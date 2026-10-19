import { SharedArray } from 'k6/data';
import { sleep, check } from 'k6';
import exec from 'k6/execution';
import { Counter, Trend } from 'k6/metrics';

import { parseResponse } from './lib/graphql.js';
import { coldLoad, getFirebaseToken } from './queries/bootstrap.js';
import { maybeAuthDance, FIREBASE_REFRESH_FRACTION } from './lib/realism.js';
import { challengePage, enrollInChallenge } from './queries/challenge.js';
import { finalizeQuiz } from './queries/quiz.js';
import { startBettingSession, submitBet, submitBetTwice, lockAndFinishSession } from './queries/betting.js';

// Betting variant of freetext-quiz-rampspike.js: the same front-loaded QR
// stampede (cumulative 1k@1s, 5k@2s, 8k@4s, 10k@10s, same token slices and
// knobs), but every user answers the five PREDEFINED betting questions from
// bench/betting_box.sh with a bet, and after the last journey the admin
// finishes the session, which settles every bet (teardown).
//
// Per question a user places (fractions configurable):
//   90%  a sensible bet: within the question's limits and the points the user
//        has left (myCurrentProject.myPoints from the challenge page, minus the
//        stakes already placed in this quiz — what the app's bet slider uses)
//   5%   no bet (betAmount 0)
//   3%   a double tap: the sensible bet sent twice at once
//   2%   a bet over the available points, which must be rejected; the user
//        then retries with a sensible bet, as after the app's error message
//
// Answers are right with probability CORRECT_RATE (answer_order 1 is the
// correct one in the fixture), so settlement pays out wins and losses.
//
// Needs >= 12,100 tokens (tokengen -limit 13000) and, unless SETTLE=0, an
// admin token (tokengen -admin-user US00000000000000000000000001).

const tokens = new SharedArray('tokens', function () {
    return JSON.parse(open('../config.json')).tokens;
});
// See freetext-quiz-rampspike.js: index-aligned Auth0 tokens for the dance
const auth0Tokens = new SharedArray('auth0Tokens', function () {
    const parsed = JSON.parse(open('../config.json')).auth0Tokens;
    return parsed && parsed.length > 0 ? parsed : [null];
});

function auth0TokenFor(idx, userId) {
    const entry = idx < auth0Tokens.length ? auth0Tokens[idx] : null;
    return entry && entry.userId === userId ? entry.token : null;
}

// Never parse config.json outside a SharedArray: init runs once per VU and
// materializing 13k tokens per VU OOMs the generator
const meta = new SharedArray('meta', function () {
    const parsed = JSON.parse(open('../config.json'));
    return [{ baseUrl: parsed.baseUrl, adminToken: parsed.adminToken ? parsed.adminToken.token : null }];
})[0];
const baseUrl = __ENV.BASE_URL || meta.baseUrl;
const adminToken = meta.adminToken;

const challengeId = __ENV.CHALLENGE_ID || 'CL01LOADTESTBETTING000000000';
const expectPoints = (__ENV.EXPECT_POINTS || '1') !== '0';
const scale = parseFloat(__ENV.RAMP_SCALE || '1');
const thinkScale = parseFloat(__ENV.THINK_SCALE || '1');
const thinkUniform = (__ENV.THINK_UNIFORM || '').match(/^([\d.]+)-([\d.]+)$/);
const preEnrolled = ['1', 'true', 'yes'].includes(String(__ENV.PRE_ENROLLED).toLowerCase());
const settle = (__ENV.SETTLE || '1') !== '0';
// SHARDS/SHARD split one run over several generators: shard i of n runs 1/n
// of every phase's arrival rate and takes every n-th token of each phase's
// slice, so shards never share a user. Run them with SETTLE=0 and settle once
// all shards are done (betting-settle.js).
const shards = Math.max(1, parseInt(__ENV.SHARDS || '1'));
const shard = parseInt(__ENV.SHARD || '0');
// Long enough for a slow settlement to finish, so its duration is measured
const finishTimeout = __ENV.FINISH_TIMEOUT || '30m';

const correctRate = parseFloat(__ENV.CORRECT_RATE || '0.5');
const noBetFraction = parseFloat(__ENV.NO_BET_FRACTION || '0.05');
const doubleTapFraction = parseFloat(__ENV.DOUBLE_TAP_FRACTION || '0.03');
const overBetFraction = parseFloat(__ENV.OVER_BET_FRACTION || '0.02');

function thinkTime(baseMin, baseRange) {
    if (thinkUniform) {
        const lo = parseFloat(thinkUniform[1]);
        const hi = parseFloat(thinkUniform[2]);
        return Math.random() * (hi - lo) + lo;
    }
    return (Math.random() * baseRange + baseMin) * thinkScale;
}

const quizCompletions = new Counter('quiz_completions');
const quizSkipped = new Counter('quiz_skipped');
const quizFailures = new Counter('quiz_failures');
const betsAccepted = new Counter('bets_accepted');
const betsNone = new Counter('bets_none');
const betsDoubleTap = new Counter('bets_double_tap');
// Over-limit bets the server correctly refused
const betsRejectedExpected = new Counter('bets_rejected_expected');
// Sensible bets refused for points, or over-limit bets accepted: a bug
const betsWrongOutcome = new Counter('bets_wrong_outcome');
const betAmount = new Trend('bet_amount');
const settlementDuration = new Trend('settlement_duration', true);

function phase(rate, timeUnit, duration, startTime, vus, tokenBase) {
    return {
        executor: 'constant-arrival-rate',
        rate: Math.max(1, Math.round((rate * scale) / shards)),
        timeUnit,
        duration,
        startTime,
        preAllocatedVUs: Math.max(10, Math.round((vus * scale) / shards)),
        maxVUs: Math.max(10, Math.round((vus * scale) / shards)),
        gracefulStop: '4m',
        env: { TOKEN_BASE: String(Math.round(tokenBase * scale)) },
        exec: 'journey',
    };
}

export const options = {
    scenarios: {
        burst1_1k_s1: phase(1000, '1s', '1s', '0s', 1200, 0),
        burst2_4k_s2: phase(4000, '1s', '1s', '1s', 4400, 1300),
        burst3_3k_s3s4: phase(3000, '2s', '2s', '2s', 3300, 6100),
        burst4_2k_s5s10: phase(2000, '6s', '6s', '4s', 2200, 9700),
    },
    // Settlement runs inside the finishQuizSession request
    teardownTimeout: '40m',
    thresholds: {
        http_req_duration: ['p(95)<500', 'p(99)<1000'],
        http_req_failed: ['rate<0.01'],
        graphql_errors: ['rate<0.01'],
        quiz_failures: ['count<20'],
        bets_wrong_outcome: ['count<1'],
        'http_req_duration{name:AuthCallback}': ['p(95)<1000'],
        'http_req_duration{name:GetMe}': ['p(95)<500'],
        'http_req_duration{name:CurrentProject}': ['p(95)<500'],
        'http_req_duration{name:ChallengePage}': ['p(95)<500'],
        'http_req_duration{name:EnrollInChallenge}': ['p(95)<500'],
        'http_req_duration{name:StartQuizSession}': ['p(95)<500'],
        'http_req_duration{name:SubmitQuizAnswerBet}': ['p(95)<500'],
        'http_req_duration{name:SubmitQuizAnswerNoBet}': ['p(95)<500'],
        'http_req_duration{name:SubmitQuizAnswerDoubleTap}': ['p(95)<1000'],
        'http_req_duration{name:FinalizeQuiz}': ['p(95)<1000'],
    },
};

function fail(iteration, message) {
    quizFailures.add(1);
    console.error(`iter ${iteration}: ${message}`);
}

// sensibleBet returns a random bet the app's slider would allow for the
// question, given the points the user has left; 0 when no bet fits
export function sensibleBet(question, available) {
    let max = available;
    if (question.bettingMaxAbsolute != null) {
        max = Math.min(max, question.bettingMaxAbsolute);
    }
    if (question.bettingMaxPercentage != null) {
        max = Math.min(max, Math.floor((available * question.bettingMaxPercentage) / 100));
    }
    let min = 1;
    if (question.bettingMinAbsolute != null) {
        min = Math.max(min, question.bettingMinAbsolute);
    }
    if (question.bettingMinPercentage != null) {
        min = Math.max(min, Math.ceil((available * question.bettingMinPercentage) / 100));
    }
    if (max < min) {
        return 0;
    }
    return min + Math.floor(Math.random() * (max - min + 1));
}

export function journey() {
    const tokenBase = parseInt(__ENV.TOKEN_BASE) || 0;
    const iteration = exec.scenario.iterationInTest;
    const idx = tokenBase + iteration * shards + shard;
    if (idx >= tokens.length) {
        fail(iteration, `token slice overrun: index ${idx} >= ${tokens.length} tokens (run tokengen with a higher -limit)`);
        return;
    }
    let { token } = tokens[idx];

    token = maybeAuthDance(baseUrl, token, auth0TokenFor(idx, tokens[idx].userId));
    const refreshFirebaseToken = Math.random() < FIREBASE_REFRESH_FRACTION;

    const coldPage = coldLoad(baseUrl, token, () => challengePage(baseUrl, token, challengeId));

    let pageResponse = coldPage;
    if (!preEnrolled) {
        enrollInChallenge(baseUrl, token, challengeId);
        pageResponse = challengePage(baseUrl, token, challengeId);
    }
    const pageData = parseResponse(pageResponse);
    const quiz = pageData && pageData.challenge && pageData.challenge.quiz;
    if (!quiz || !quiz.userActiveSession) {
        fail(iteration, `no active session on challenge ${challengeId}`);
        return;
    }

    const submission = startBettingSession(baseUrl, token, quiz.userActiveSession.id);
    if (!submission || !submission.orderedQuestions || submission.orderedQuestions.length === 0) {
        fail(iteration, 'failed to start quiz session');
        return;
    }
    if (submission.completedAt) {
        quizSkipped.add(1);
        return;
    }

    const project = pageData.myCurrentProject;
    if (!project || project.myPoints == null) {
        fail(iteration, 'challenge page has no myCurrentProject.myPoints');
        return;
    }
    // Stakes stay reserved until settlement, so every accepted bet lowers
    // what the user can still bet in this quiz
    let available = Math.max(0, project.myPoints);

    for (let i = 0; i < submission.orderedQuestions.length; i++) {
        const question = submission.orderedQuestions[i];
        if (question.__typename !== 'PredefinedQuestion') {
            fail(iteration, `unexpected question type ${question.__typename}`);
            return;
        }
        const thinkSeconds = thinkTime(5, 10);
        sleep(thinkSeconds);
        const timeSpent = Math.round(thinkSeconds);

        const answers = question.predefinedAnswers;
        const correct = answers.find((a) => a.answerOrder === 1);
        const wrong = answers.filter((a) => a.answerOrder !== 1);
        const answerId = Math.random() < correctRate || wrong.length === 0
            ? correct.id
            : wrong[Math.floor(Math.random() * wrong.length)].id;

        const roll = Math.random();
        let bet = sensibleBet(question, available);
        let result;
        if (roll < noBetFraction) {
            bet = 0;
            result = submitBet(baseUrl, token, submission.id, question.id, answerId, 0, timeSpent);
        } else if (roll < noBetFraction + doubleTapFraction) {
            betsDoubleTap.add(1);
            const results = submitBetTwice(baseUrl, token, submission.id, question.id, answerId, bet, timeSpent);
            // Both are the same bet on the same answer: the second replaces the
            // first, so the stake counts once
            result = results.find((r) => r.status === 'accepted') || results[0];
            for (const r of results) {
                if (r.status === 'rejected') {
                    betsWrongOutcome.add(1);
                    console.error(`iter ${iteration}: double-tap bet ${bet} of ${available} rejected: ${r.message}`);
                }
            }
        } else {
            if (roll < noBetFraction + doubleTapFraction + overBetFraction) {
                const over = available + 1 + Math.floor(Math.random() * 20);
                const overResult = submitBet(baseUrl, token, submission.id, question.id, answerId, over, timeSpent);
                if (overResult.status === 'rejected') {
                    betsRejectedExpected.add(1);
                } else if (overResult.status === 'accepted') {
                    betsWrongOutcome.add(1);
                    console.error(`iter ${iteration}: bet ${over} accepted with only ${available} available`);
                    available = Math.max(0, available - over);
                    continue;
                }
                // After the error message the user picks a bet that fits
            }
            result = submitBet(baseUrl, token, submission.id, question.id, answerId, bet, timeSpent);
            if (result.status === 'rejected') {
                betsWrongOutcome.add(1);
                console.error(`iter ${iteration}: sensible bet ${bet} of ${available} rejected: ${result.message}`);
            }
        }

        if (result.status !== 'accepted') {
            if (result.status === 'error') {
                fail(iteration, `failed to answer question ${question.id}`);
                return;
            }
            continue;
        }
        if (bet > 0) {
            betsAccepted.add(1);
            betAmount.add(bet);
            available -= bet;
        } else {
            betsNone.add(1);
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
    check(finalized, { 'quiz completed': (f) => Boolean(f.completedAt) });
    if (expectPoints) {
        check(finalized, { 'points awarded': (f) => (f.pointsAwarded || 0) > 0 });
    }
}

export function setup() {
    console.log(
        `Betting ramped spike: cumulative 1k@1s, 5k@2s, 8k@4s, 10k@10s (scale ${scale}); ${tokens.length} tokens; ` +
        `shard ${shard + 1}/${shards}; settle=${settle}; target ${baseUrl}`
    );
    if (settle && !adminToken) {
        throw new Error('config.json has no adminToken: run tokengen with -admin-user, or set SETTLE=0');
    }
    const { token } = tokens[0];
    const data = parseResponse(challengePage(baseUrl, token, challengeId));
    const quiz = data && data.challenge && data.challenge.quiz;
    if (!quiz) {
        throw new Error(`Challenge ${challengeId} not found — run bench/betting_box.sh reset first`);
    }
    if (!quiz.userActiveSession) {
        throw new Error(`No OPEN session for challenge ${challengeId}`);
    }
    console.log(`Quiz: ${quiz.name}, session ${quiz.userActiveSession.id}, completionPoints=${quiz.completionPoints}`);
    return { sessionId: quiz.userActiveSession.id };
}

// Finishing the session settles every bet inside the request, like the admin
// pressing "finish" after the camp quiz
export function teardown(data) {
    if (!settle || !data || !data.sessionId) {
        return;
    }
    console.log(`Finishing session ${data.sessionId} (settles all bets, timeout ${finishTimeout})...`);
    const res = lockAndFinishSession(baseUrl, adminToken, data.sessionId, finishTimeout);
    settlementDuration.add(res.finishMs);
    console.log(
        `Session finished: locked=${res.locked} finished=${res.finished} HTTP ${res.status} ` +
        `in ${(res.finishMs / 1000).toFixed(1)}s`
    );
}
