import http from 'k6/http';
import { check } from 'k6';
import { graphqlRequest, checkGraphQLResponse, parseResponse, graphqlErrors } from '../lib/graphql.js';

const SUBMIT_BET_MUTATION = `
mutation SubmitQuizAnswer($submissionId: ID!, $input: SubmitQuizAnswerInput!) {
    submitQuizAnswer(submissionId: $submissionId, input: $input) {
        id
        answeredAt
        betAmount
        ... on PredefinedResponse {
            isCorrect
        }
    }
}
`;

// startQuizSession with the betting limits the app's bet slider uses
const START_BETTING_SESSION_MUTATION = `
mutation StartQuizSession($sessionId: ID!) {
    startQuizSession(sessionId: $sessionId) {
        id
        completedAt
        orderedQuestions {
            __typename
            id
            questionOrder
            bettingEnabled
            bettingMinAbsolute
            bettingMaxAbsolute
            bettingMinPercentage
            bettingMaxPercentage
            ... on PredefinedQuestion {
                predefinedAnswers {
                    id
                    answerOrder
                }
            }
        }
    }
}
`;

const LOCK_SESSION_MUTATION = `
mutation LockQuizSession($id: ID!) {
    lockQuizSession(id: $id) { id state }
}
`;

const FINISH_SESSION_MUTATION = `
mutation FinishQuizSession($id: ID!) {
    finishQuizSession(id: $id) { id state }
}
`;

// The one rejection the server may legitimately give a bet under load: the
// user no longer has the points (stale client state, or a concurrent bet)
const REJECTED_FOR_POINTS = 'exceeds available points';

/**
 * Start (or resume) the user's submission in a session, with betting limits
 * @returns {object|null} submission, or null on error
 */
export function startBettingSession(baseUrl, token, sessionId) {
    const response = graphqlRequest(baseUrl, START_BETTING_SESSION_MUTATION, { sessionId }, token, 'StartQuizSession');
    if (!checkGraphQLResponse(response, 'StartQuizSession')) {
        return null;
    }
    const data = parseResponse(response);
    return data ? data.startQuizSession : null;
}

function betRequest(baseUrl, token, submissionId, questionId, answerId, betAmount, timeSpentSeconds, name) {
    return {
        method: 'POST',
        url: `${baseUrl}/graphql`,
        body: JSON.stringify({
            query: SUBMIT_BET_MUTATION,
            variables: {
                submissionId,
                input: { questionId, selectedAnswerIds: [answerId], betAmount, timeSpentSeconds },
            },
        }),
        params: {
            headers: { 'Content-Type': 'application/json', 'Authorization': `Bearer ${token}` },
            tags: { name },
        },
    };
}

/**
 * Classify a bet response: 'accepted' (with the stored response), 'rejected'
 * (not enough points) or 'error' (anything else, logged and counted as a
 * GraphQL error)
 */
function classify(response, name) {
    let body = null;
    try {
        body = JSON.parse(response.body);
    } catch {
        // handled below
    }
    if (response.status === 200 && body && body.data && body.data.submitQuizAnswer && !body.errors) {
        check(response, { [`${name}: accepted`]: () => true });
        return { status: 'accepted', answer: body.data.submitQuizAnswer };
    }
    const errors = body && body.errors ? JSON.stringify(body.errors) : '';
    if (errors.includes(REJECTED_FOR_POINTS)) {
        return { status: 'rejected', message: errors };
    }
    graphqlErrors.add(1, { operation: name });
    check(response, { [`${name}: accepted`]: () => false });
    const message = response.status !== 200 ? `HTTP ${response.status} ${response.error || ''}` : errors || 'invalid JSON';
    console.error(`${name}: ${message}`);
    return { status: 'error', message };
}

/**
 * Answer a predefined question with a bet (0 = no bet; betting questions
 * require betAmount to be sent)
 */
export function submitBet(baseUrl, token, submissionId, questionId, answerId, betAmount, timeSpentSeconds) {
    const name = betAmount > 0 ? 'SubmitQuizAnswerBet' : 'SubmitQuizAnswerNoBet';
    const req = betRequest(baseUrl, token, submissionId, questionId, answerId, betAmount, timeSpentSeconds, name);
    return classify(http.post(req.url, req.body, req.params), name);
}

/**
 * Double tap: the same bet sent twice at once, like a retry on a flaky
 * connection. Returns both classified results.
 */
export function submitBetTwice(baseUrl, token, submissionId, questionId, answerId, betAmount, timeSpentSeconds) {
    const name = 'SubmitQuizAnswerDoubleTap';
    const req = betRequest(baseUrl, token, submissionId, questionId, answerId, betAmount, timeSpentSeconds, name);
    return http.batch([req, req]).map((r) => classify(r, name));
}

/**
 * Lock, then finish a quiz session as an admin. Finishing settles every bet
 * in the session synchronously, so it gets a long timeout.
 * @returns {{locked: boolean, finished: boolean, finishMs: number, status: number}}
 */
export function lockAndFinishSession(baseUrl, adminToken, sessionId, finishTimeout) {
    const lock = graphqlRequest(baseUrl, LOCK_SESSION_MUTATION, { id: sessionId }, adminToken, 'LockQuizSession');
    const locked = checkGraphQLResponse(lock, 'LockQuizSession');
    const finish = graphqlRequest(baseUrl, FINISH_SESSION_MUTATION, { id: sessionId }, adminToken, 'FinishQuizSession',
        { timeout: finishTimeout });
    const finished = checkGraphQLResponse(finish, 'FinishQuizSession');
    return { locked, finished, finishMs: finish.timings.duration, status: finish.status };
}
