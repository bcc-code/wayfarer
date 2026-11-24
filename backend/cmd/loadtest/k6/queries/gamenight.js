import http from 'k6/http';
import { graphqlRequest, checkGraphQLResponse, parseResponse } from '../lib/graphql.js';

// The church admin's betting flow on the gamenight page
// (frontend/layers/admin/app/pages/admin/my-church/gamenights/gamenight-[gamenight].vue),
// with the operations from frontend/app/graphql/{mutations/quizes.gql,
// mutations/challenges.gql, queries/quizzes.gql}.

const SESSION_FIELDS = `
    id
    name
    state
    openAt
    lockAt
    finishAt
    createdAt
    accessCount
    quiz {
      id
      name
    }
    createdBy {
      id
      name
    }
`;

const CREATE_QUIZ_SESSION_MUTATION = `
mutation CreateQuizSession($input: CreateQuizSessionInput!) {
  createQuizSession(input: $input) {
    id
    name
    state
    openAt
    lockAt
    finishAt
    createdAt
    quiz {
      id
      name
    }
  }
}
`;

const OPEN_QUIZ_SESSION_MUTATION = `
mutation OpenQuizSession($id: ID!) {
  openQuizSession(id: $id) {${SESSION_FIELDS}  }
}
`;

const LOCK_QUIZ_SESSION_MUTATION = `
mutation LockQuizSession($id: ID!) {
  lockQuizSession(id: $id) {${SESSION_FIELDS}  }
}
`;

const FINISH_QUIZ_SESSION_MUTATION = `
mutation FinishQuizSession($id: ID!) {
  finishQuizSession(id: $id) {${SESSION_FIELDS}  }
}
`;

const GRANT_QUIZ_SESSION_ACCESS_ASYNC_MUTATION = `
mutation GrantQuizSessionAccessAsync($input: GrantQuizSessionAccessInput!) {
  grantQuizSessionAccessAsync(input: $input) {
    id
    operationType
    status
    totalCount
    successCount
    errorMessage
  }
}
`;

const QUIZ_DETAILS_QUERY = `
query QuizDetails($id: ID!) {
  quiz(id: $id) {
    id
    name
    challenge {
      id
    }
    project {
      id
    }
  }
}
`;

const BULK_ENROLL_USERS_IN_CHALLENGE_MUTATION = `
mutation BulkEnrollUsersInChallenge($challengeId: ID!, $target: EnrollmentTargetInput!) {
  bulkEnrollUsersInChallenge(challengeId: $challengeId, target: $target) {
    id
  }
}
`;

const TAGS = { kind: 'admin' };

function run(baseUrl, token, query, variables, name, field, extraParams) {
    const response = graphqlRequest(baseUrl, query, variables, token, name, {
        tags: { name, ...TAGS },
        ...extraParams,
    });
    if (!checkGraphQLResponse(response, name)) {
        return null;
    }
    const data = parseResponse(response);
    return data ? data[field] : null;
}

/**
 * The page's startBetting(): create and open the session, grant the church
 * access (a background bulk job), then bulk-enroll the church in the
 * challenge. Each step only runs if the previous one worked, like the page.
 * @returns {{sessionId: string|null, ok: boolean}}
 */
export function startBetting(baseUrl, token, quizId, churchId, sessionName) {
    const session = run(baseUrl, token, CREATE_QUIZ_SESSION_MUTATION,
        { input: { quizId, name: sessionName } }, 'CreateQuizSession', 'createQuizSession');
    if (!session) {
        return { sessionId: null, ok: false };
    }
    const opened = run(baseUrl, token, OPEN_QUIZ_SESSION_MUTATION, { id: session.id }, 'OpenQuizSession', 'openQuizSession');
    if (!opened) {
        return { sessionId: session.id, ok: false };
    }
    const job = run(baseUrl, token, GRANT_QUIZ_SESSION_ACCESS_ASYNC_MUTATION,
        { input: { sessionId: session.id, churchIds: [churchId] } },
        'GrantQuizSessionAccessAsync', 'grantQuizSessionAccessAsync');
    if (!job || job.status === 'FAILED') {
        return { sessionId: session.id, ok: false };
    }
    const quiz = run(baseUrl, token, QUIZ_DETAILS_QUERY, { id: quizId }, 'QuizDetails', 'quiz');
    if (!quiz) {
        return { sessionId: session.id, ok: false };
    }
    const enrolled = run(baseUrl, token, BULK_ENROLL_USERS_IN_CHALLENGE_MUTATION, {
        challengeId: quiz.challenge.id,
        target: { churchInProject: { churchId, projectId: quiz.project.id } },
    }, 'BulkEnrollUsersInChallenge', 'bulkEnrollUsersInChallenge');
    return { sessionId: session.id, ok: Boolean(enrolled) };
}

function sessionMutation(baseUrl, token, query, sessionId, name, field, doubleClick, timeout) {
    const params = timeout ? { timeout } : {};
    if (!doubleClick) {
        const result = run(baseUrl, token, query, { id: sessionId }, name, field, params);
        return { ok: Boolean(result), state: result ? result.state : null };
    }
    // A double click sends the mutation twice at once
    const req = {
        method: 'POST',
        url: `${baseUrl}/graphql`,
        body: JSON.stringify({ query, variables: { id: sessionId } }),
        params: {
            headers: { 'Content-Type': 'application/json', 'Authorization': `Bearer ${token}` },
            tags: { name: `${name}DoubleClick`, ...TAGS },
            ...params,
        },
    };
    // One of the two may legitimately fail (e.g. "can only lock session from
    // OPEN state"), so neither counts as a GraphQL error; at least one must work
    const results = http.batch([req, req]).map((r) => (r.status === 200 ? parseResponse(r) : null));
    const result = results.find((r) => r && r[field]);
    return { ok: Boolean(result), state: result ? result[field].state : null };
}

/** The page's lockBetting() */
export function lockBetting(baseUrl, token, sessionId, doubleClick) {
    return sessionMutation(baseUrl, token, LOCK_QUIZ_SESSION_MUTATION, sessionId, 'LockQuizSession', 'lockQuizSession', doubleClick);
}

/**
 * The page's finishBetting(). Finishing settles the session's bets inside
 * the request, so it gets a long timeout.
 */
export function finishBetting(baseUrl, token, sessionId, doubleClick, timeout) {
    return sessionMutation(baseUrl, token, FINISH_QUIZ_SESSION_MUTATION, sessionId, 'FinishQuizSession', 'finishQuizSession', doubleClick, timeout);
}
