import { SharedArray } from 'k6/data';
import { sleep, check } from 'k6';
import exec from 'k6/execution';
import { Counter } from 'k6/metrics';

import { parseResponse } from './lib/graphql.js';
import { coldLoad } from './queries/bootstrap.js';
import { challengePage, enrollInChallenge } from './queries/challenge.js';
import { startQuizSession, submitTextAnswer, submitAnswer, submitNumberAnswer, finalizeQuiz } from './queries/quiz.js';

// Free-text quiz spike: SPIKE_USERS distinct users scan a QR code within
// SPIKE_WINDOW seconds, landing directly on the quiz challenge URL
// (/challenges/{id}?enroll=true — see AdminChallengeQrModal.vue). Each user
// cold-loads the app on the challenge page, self-enrolls, answers the four
// questions (2x free text, 1x multiple choice, 1x number) and finalizes. The
// quiz awards no points and no achievements (see
// scripts/insert_freetext_quiz.sql).
//
// Each iteration is one distinct user (token indexed by iteration number),
// so every user runs the journey exactly once.

// open()+JSON.parse() must happen inside the SharedArray callback: k6 runs
// this whole file's init code once per VU, so parsing outside the callback
// makes every VU hold its own full copy of the token array instead of the
// single shared one (with tens of thousands of VUs, that's gigabytes of RAM).
const tokens = new SharedArray('tokens', function () {
    return JSON.parse(open('../config.json')).tokens;
});
const baseUrl = new SharedArray('baseUrl', function () {
    return [JSON.parse(open('../config.json')).baseUrl];
})[0];

const challengeId = __ENV.CHALLENGE_ID || 'CL01LOADTESTFREETEXT00000000';
const spikeUsers = parseInt(__ENV.SPIKE_USERS) || 10000;
const spikeWindowSeconds = parseInt(__ENV.SPIKE_WINDOW) || 10;
// Skip the first N tokens — lets repeated small runs use fresh users
// without re-running the seed script (users keep completed submissions)
const tokenOffset = parseInt(__ENV.TOKEN_OFFSET) || 0;

const quizCompletions = new Counter('quiz_completions');
const quizSkipped = new Counter('quiz_skipped');
const quizFailures = new Counter('quiz_failures');

export const options = {
    scenarios: {
        freetext_quiz_spike: {
            executor: 'constant-arrival-rate',
            rate: Math.ceil(spikeUsers / spikeWindowSeconds),
            timeUnit: '1s',
            duration: `${spikeWindowSeconds}s`,
            // Journeys (think time + typing) outlast the arrival window, so
            // nearly all users are active at once; pre-allocated headroom
            // avoids dropped iterations at the tail (on-demand VU init lags)
            preAllocatedVUs: Math.ceil(spikeUsers * 1.1) + 5,
            maxVUs: Math.ceil(spikeUsers * 1.1) + 5,
            // Journeys include up to ~150s of think time across the four
            // questions; let in-flight users finish
            gracefulStop: '4m',
        },
    },
    thresholds: {
        http_req_duration: ['p(95)<500', 'p(99)<1000'],
        http_req_failed: ['rate<0.01'],
        graphql_errors: ['rate<0.01'],
        quiz_failures: ['count<1'],
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

export default function () {
    const iteration = exec.scenario.iterationInTest;
    const { token } = tokens[(iteration + tokenOffset) % tokens.length];

    // QR code entry: cold load landing directly on the challenge page
    coldLoad(baseUrl, token, () => challengePage(baseUrl, token, challengeId));

    // ?enroll=true fires the enroll mutation, then the page refetches
    enrollInChallenge(baseUrl, token, challengeId);
    const pageData = parseResponse(challengePage(baseUrl, token, challengeId));
    const quiz = pageData && pageData.challenge && pageData.challenge.quiz;
    if (!quiz || !quiz.userActiveSession) {
        fail(iteration, `no active session on challenge ${challengeId} — run scripts/insert_freetext_quiz.sql`);
        return;
    }

    const submission = startQuizSession(baseUrl, token, quiz.userActiveSession.id);
    if (!submission || !submission.orderedQuestions || submission.orderedQuestions.length === 0) {
        fail(iteration, 'failed to start quiz session');
        return;
    }
    if (submission.completedAt) {
        // User already took the quiz in a previous run — re-run the seed
        // script to reset submissions, or set TOKEN_OFFSET for fresh users
        quizSkipped.add(1);
        return;
    }

    // Answer every question in order, with per-type think time
    for (let i = 0; i < submission.orderedQuestions.length; i++) {
        const question = submission.orderedQuestions[i];
        let answer = null;

        switch (question.__typename) {
            case 'FreeTextQuestion': {
                // Reading the question + typing the answer
                const thinkSeconds = Math.random() * 30 + 20;
                sleep(thinkSeconds);
                const answerText = `${ANSWERS[(iteration + i) % ANSWERS.length]} (#${iteration})`;
                answer = submitTextAnswer(baseUrl, token, submission.id, question.id, answerText, Math.round(thinkSeconds));
                break;
            }
            case 'PredefinedQuestion': {
                const thinkSeconds = Math.random() * 20 + 10;
                sleep(thinkSeconds);
                // Every seeded answer is correct, so select all of them
                const selectedIds = question.predefinedAnswers.map((a) => a.id);
                answer = submitAnswer(baseUrl, token, submission.id, question.id, selectedIds, Math.round(thinkSeconds));
                if (answer) {
                    check(answer, { 'multiple choice graded correct': (a) => a.isCorrect === true });
                }
                break;
            }
            case 'NumberQuestion': {
                const thinkSeconds = Math.random() * 10 + 5;
                sleep(thinkSeconds);
                const min = question.minValue == null ? 1 : question.minValue;
                const max = question.maxValue == null ? 100 : question.maxValue;
                const value = Math.floor(Math.random() * (max - min + 1)) + min;
                answer = submitNumberAnswer(baseUrl, token, submission.id, question.id, value, Math.round(thinkSeconds));
                break;
            }
            default:
                fail(iteration, `unexpected question type ${question.__typename} on question ${question.id}`);
                return;
        }

        if (!answer) {
            fail(iteration, `failed to submit ${question.__typename} answer for question ${question.id}`);
            return;
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
        'no points awarded': (f) => (f.pointsAwarded || 0) === 0,
        'score is zero': (f) => (f.score || 0) === 0,
    });
}

export function setup() {
    console.log(
        `Free-text quiz spike: ${spikeUsers} users over ${spikeWindowSeconds}s (${Math.ceil(
            spikeUsers / spikeWindowSeconds
        )}/s), ${tokens.length} tokens, target ${baseUrl}`
    );
    if (spikeUsers > tokens.length) {
        console.warn(`Only ${tokens.length} tokens for ${spikeUsers} users — tokens will be reused (run make loadtest-gen-all)`);
    }

    // Fail fast if the quiz challenge or session is missing
    const { token } = tokens[0];
    const data = parseResponse(challengePage(baseUrl, token, challengeId));
    const quiz = data && data.challenge && data.challenge.quiz;
    if (!quiz) {
        throw new Error(`Challenge ${challengeId} not found — run scripts/insert_freetext_quiz.sql first`);
    }
    if (!quiz.userActiveSession) {
        throw new Error(`No open session with access on quiz ${quiz.id} — run scripts/insert_freetext_quiz.sql first`);
    }
    console.log(`Quiz found: ${quiz.name}, session ${quiz.userActiveSession.id}, completionPoints=${quiz.completionPoints}`);
}
