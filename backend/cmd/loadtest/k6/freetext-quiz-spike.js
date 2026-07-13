import { SharedArray } from 'k6/data';
import { sleep, check } from 'k6';
import exec from 'k6/execution';
import { Counter } from 'k6/metrics';

import { parseResponse } from './lib/graphql.js';
import { coldLoad } from './queries/bootstrap.js';
import { activeChallengesPage } from './queries/challenges.js';
import { challengePage } from './queries/challenge.js';
import { startQuizSession, submitTextAnswer, finalizeQuiz } from './queries/quiz.js';

// Free-text quiz spike: SPIKE_USERS distinct users enter the app within
// SPIKE_WINDOW seconds, land on the challenges page, open the free-text quiz
// challenge, answer its single question and finalize. The quiz awards no
// points and no achievements (see scripts/insert_freetext_quiz.sql).
//
// Each iteration is one distinct user (token indexed by iteration number),
// so every user runs the journey exactly once.

const config = JSON.parse(open('../config.json'));
const tokens = new SharedArray('tokens', function () {
    return config.tokens;
});
const baseUrl = config.baseUrl;

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
            // Journeys include think time; let in-flight users finish
            gracefulStop: '2m',
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

    // App entry: cold load landing on the challenges page
    coldLoad(baseUrl, token, () => activeChallengesPage(baseUrl, token));

    // Scan the challenge list, then open the quiz challenge
    sleep(Math.random() * 2 + 1);
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

    // Type the free-text answer
    const typingSeconds = Math.random() * 6 + 2;
    sleep(typingSeconds);

    const answerText = `${ANSWERS[iteration % ANSWERS.length]} (#${iteration})`;
    const question = submission.orderedQuestions[0];
    const answer = submitTextAnswer(baseUrl, token, submission.id, question.id, answerText, Math.round(typingSeconds));
    if (!answer) {
        fail(iteration, 'failed to submit free-text answer');
        return;
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
