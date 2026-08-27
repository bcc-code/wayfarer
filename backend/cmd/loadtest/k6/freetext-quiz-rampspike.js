import { SharedArray } from 'k6/data';
import { sleep, check } from 'k6';
import exec from 'k6/execution';
import { Counter } from 'k6/metrics';

import { parseResponse } from './lib/graphql.js';
import { coldLoad, getFirebaseToken } from './queries/bootstrap.js';
import { maybeAuthDance, FIREBASE_REFRESH_FRACTION } from './lib/realism.js';
import { challengePage, enrollInChallenge } from './queries/challenge.js';
import { startQuizSession, submitTextAnswer, submitAnswer, submitNumberAnswer, finalizeQuiz } from './queries/quiz.js';

// Ramped QR-code stampede: instead of a flat arrival rate, model a real
// scan-the-poster curve that is steep at the front and tapers off, hitting
// these CUMULATIVE arrival counts:
//
//   end of second 1  ->  1,000 users
//   end of second 2  ->  5,000 users
//   end of second 4  ->  8,000 users
//   end of second 10 -> 10,000 users
//
// Implemented as four time-phased constant-arrival-rate scenarios, each with
// its own slice of the token pool (TOKEN_BASE) so no user is reused: k6's
// iterationInTest is per-scenario, so without distinct bases the scenarios
// would all start from token 0 and collide.
//
//   burst1  1,000 over 1s (1,000/s)   tokens     0..1,299
//   burst2  4,000 over 1s (4,000/s)   tokens 1,300..6,099
//   burst3  3,000 over 2s (1,500/s)   tokens 6,100..9,699
//   burst4  2,000 over 6s (  333/s)   tokens 9,700..12,099
//
// Needs >= 12,100 tokens: run tokengen with -limit 13000.
//
// Unlike freetext-quiz-spike.js, this variant expects the quiz to AWARD
// POINTS by default (EXPECT_POINTS=1), so the score_journal write and its
// trigger fan-out are exercised. Set EXPECT_POINTS=0 for the zero-point A/B.

const tokens = new SharedArray('tokens', function () {
    return JSON.parse(open('../config.json')).tokens;
});
// Simulated Auth0 tokens (tokengen -auth0-count) for the REALISM auth-dance
// fraction. Quiz users are single-use per token slice, so the dance must mint
// a token for the SAME user, never a random one: tokengen orders both arrays
// by user id, so when every user got an auth0 token they are index-aligned —
// verified per lookup via the userId guard in auth0TokenFor(). Never copy
// SharedArray entries into a per-VU map here: module init runs once per VU
// and materializing 13k tokens per VU OOMs the generator.
// SharedArray callbacks must return a non-empty array; null marks "not generated".
const auth0Tokens = new SharedArray('auth0Tokens', function () {
    const parsed = JSON.parse(open('../config.json')).auth0Tokens;
    return parsed && parsed.length > 0 ? parsed : [null];
});

function auth0TokenFor(idx, userId) {
    const entry = idx < auth0Tokens.length ? auth0Tokens[idx] : null;
    return entry && entry.userId === userId ? entry.token : null;
}
// BASE_URL overrides the baseUrl baked into config.json by tokengen, so a
// config generated on the server box (baseUrl 127.0.0.1) can be replayed
// from an off-box load generator without re-minting tokens.
const baseUrl = __ENV.BASE_URL || new SharedArray('baseUrl', function () {
    return [JSON.parse(open('../config.json')).baseUrl];
})[0];

const challengeId = __ENV.CHALLENGE_ID || 'CL01LOADTESTFREETEXT00000000';
const expectPoints = (__ENV.EXPECT_POINTS || '1') !== '0';
// RAMP_SCALE scales every phase's arrival rate while preserving the curve's
// shape, so the same profile can be run at reduced load (e.g. to separate
// server cost from load-generator contention when both share a machine).
const scale = parseFloat(__ENV.RAMP_SCALE || '1');
// THINK_SCALE compresses per-question think time (e.g. 0.02 turns 20-50s
// reads into <1s) for fast iteration runs. NOTE: this changes the workload —
// finalizes then land at ~arrival rate instead of trickling over the tail,
// so it's a HARSHER scoring-path test, not comparable to THINK_SCALE=1 runs.
const thinkScale = parseFloat(__ENV.THINK_SCALE || '1');

const quizCompletions = new Counter('quiz_completions');
const quizSkipped = new Counter('quiz_skipped');
const quizFailures = new Counter('quiz_failures');

// preAllocatedVUs must cover the whole in-flight journey (~95-150s of think
// time), not just the arrival window, or the tail drops iterations.
function phase(rate, timeUnit, duration, startTime, vus, tokenBase) {
    return {
        executor: 'constant-arrival-rate',
        rate: Math.max(1, Math.round(rate * scale)),
        timeUnit,
        duration,
        startTime,
        preAllocatedVUs: Math.max(10, Math.round(vus * scale)),
        maxVUs: Math.max(10, Math.round(vus * scale)),
        gracefulStop: '4m',
        // Token slices scale with RAMP_SCALE so bursts don't overrun into each
        // other's users. Total tokens needed ≈ 12,100 * scale (18,150 at 1.5).
        env: { TOKEN_BASE: String(Math.round(tokenBase * scale)) },
        exec: 'journey',
    };
}

export const options = {
    scenarios: {
        // Token bases carry ~30% headroom over each phase's iteration count:
        // constant-arrival-rate can overshoot its nominal total slightly, and a
        // slice that overruns into the next one reuses a user who has already
        // completed the quiz, which surfaces as "submission already completed".
        burst1_1k_s1: phase(1000, '1s', '1s', '0s', 1200, 0),
        burst2_4k_s2: phase(4000, '1s', '1s', '1s', 4400, 1300),
        burst3_3k_s3s4: phase(3000, '2s', '2s', '2s', 3300, 6100),
        burst4_2k_s5s10: phase(2000, '6s', '6s', '4s', 2200, 9700),
    },
    thresholds: {
        http_req_duration: ['p(95)<500', 'p(99)<1000'],
        http_req_failed: ['rate<0.01'],
        graphql_errors: ['rate<0.01'],
        // count<1 is unreachable at 10k scale (a single loopback RST trips it);
        // allow a handful and rely on http_req_failed/graphql_errors for signal
        quiz_failures: ['count<20'],
        'http_req_duration{name:AuthCallback}': ['p(95)<1000'],
        'http_req_duration{name:GetMe}': ['p(95)<500'],
        'http_req_duration{name:CurrentProject}': ['p(95)<500'],
        'http_req_duration{name:GetFirebaseToken}': ['p(95)<500'],
        'http_req_duration{name:ChallengePage}': ['p(95)<500'],
        'http_req_duration{name:EnrollInChallenge}': ['p(95)<500'],
        'http_req_duration{name:StartQuizSession}': ['p(95)<500'],
        'http_req_duration{name:SubmitQuizAnswer}': ['p(95)<500'],
        'http_req_duration{name:FinalizeQuiz}': ['p(95)<1000'],
    },
};

const ANSWERS = [
    'That kindness matters more than being right.',
    'Patience is something you practice, not something you have.',
    'I learned to listen before answering.',
    'Small daily habits add up to big changes.',
    'Being grateful changes how the whole day feels.',
    'Helping someone else is the fastest way to feel better.',
    'It is okay to ask for help.',
    'Courage is acting despite the fear, not without it.',
];

function fail(iteration, message) {
    quizFailures.add(1);
    console.error(`iter ${iteration}: ${message}`);
}

export function journey() {
    const tokenBase = parseInt(__ENV.TOKEN_BASE) || 0;
    const iteration = exec.scenario.iterationInTest;
    const idx = tokenBase + iteration;
    if (idx >= tokens.length) {
        fail(iteration, `token slice overrun: index ${idx} >= ${tokens.length} tokens (run tokengen with a higher -limit)`);
        return;
    }
    let { token } = tokens[idx];

    // A fraction of sessions start expired: Auth0 dance for the SAME user,
    // then continue with the freshly minted Wayfarer JWT.
    token = maybeAuthDance(baseUrl, token, auth0TokenFor(idx, tokens[idx].userId));
    // A fraction of sessions outlive their 1h Firebase custom token during the
    // long quiz think time and re-fetch it at the quiz midpoint.
    const refreshFirebaseToken = Math.random() < FIREBASE_REFRESH_FRACTION;

    coldLoad(baseUrl, token, () => challengePage(baseUrl, token, challengeId));

    enrollInChallenge(baseUrl, token, challengeId);
    const pageData = parseResponse(challengePage(baseUrl, token, challengeId));
    const quiz = pageData && pageData.challenge && pageData.challenge.quiz;
    if (!quiz || !quiz.userActiveSession) {
        fail(iteration, `no active session on challenge ${challengeId}`);
        return;
    }

    const submission = startQuizSession(baseUrl, token, quiz.userActiveSession.id);
    if (!submission || !submission.orderedQuestions || submission.orderedQuestions.length === 0) {
        fail(iteration, 'failed to start quiz session');
        return;
    }
    if (submission.completedAt) {
        quizSkipped.add(1);
        return;
    }

    for (let i = 0; i < submission.orderedQuestions.length; i++) {
        const question = submission.orderedQuestions[i];
        let answer = null;

        switch (question.__typename) {
            case 'FreeTextQuestion': {
                const thinkSeconds = (Math.random() * 30 + 20) * thinkScale;
                sleep(thinkSeconds);
                const answerText = `${ANSWERS[(iteration + i) % ANSWERS.length]} (#${tokenBase + iteration})`;
                answer = submitTextAnswer(baseUrl, token, submission.id, question.id, answerText, Math.round(thinkSeconds));
                break;
            }
            case 'PredefinedQuestion': {
                const thinkSeconds = (Math.random() * 20 + 10) * thinkScale;
                sleep(thinkSeconds);
                const selectedIds = question.predefinedAnswers.map((a) => a.id);
                answer = submitAnswer(baseUrl, token, submission.id, question.id, selectedIds, Math.round(thinkSeconds));
                if (answer) {
                    check(answer, { 'multiple choice graded correct': (a) => a.isCorrect === true });
                }
                break;
            }
            case 'NumberQuestion': {
                const thinkSeconds = (Math.random() * 10 + 5) * thinkScale;
                sleep(thinkSeconds);
                const min = question.minValue == null ? 1 : question.minValue;
                const max = question.maxValue == null ? 100 : question.maxValue;
                const value = Math.floor(Math.random() * (max - min + 1)) + min;
                answer = submitNumberAnswer(baseUrl, token, submission.id, question.id, value, Math.round(thinkSeconds));
                break;
            }
            default:
                fail(iteration, `unexpected question type ${question.__typename}`);
                return;
        }

        if (!answer) {
            fail(iteration, `failed to submit ${question.__typename} answer for question ${question.id}`);
            return;
        }

        if (refreshFirebaseToken && i === Math.floor(submission.orderedQuestions.length / 2)) {
            getFirebaseToken(baseUrl, token);
        }
    }

    const finalized = finalizeQuiz(baseUrl, token, submission.id);
    if (!finalized) {
        fail(iteration, 'failed to finalize quiz');
        return;
    }

    quizCompletions.add(1);
    check(finalized, {
        'quiz completed': (f) => Boolean(f.completedAt),
    });
    if (expectPoints) {
        check(finalized, { 'points awarded': (f) => (f.pointsAwarded || 0) > 0 });
    } else {
        check(finalized, { 'no points awarded': (f) => (f.pointsAwarded || 0) === 0 });
    }
}

export function setup() {
    console.log(
        `Ramped spike: cumulative 1k@1s, 5k@2s, 8k@4s, 10k@10s; ${tokens.length} tokens; ` +
        `expectPoints=${expectPoints}; target ${baseUrl}`
    );
    const { token } = tokens[0];
    const data = parseResponse(challengePage(baseUrl, token, challengeId));
    const quiz = data && data.challenge && data.challenge.quiz;
    if (!quiz) {
        throw new Error(`Challenge ${challengeId} not found — run setup_fixtures.sh first`);
    }
    if (!quiz.userActiveSession) {
        throw new Error(`No OPEN session for challenge ${challengeId}`);
    }
    console.log(`Quiz: ${quiz.name}, session ${quiz.userActiveSession.id}, completionPoints=${quiz.completionPoints}`);
}
