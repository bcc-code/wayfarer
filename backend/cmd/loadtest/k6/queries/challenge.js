import { graphqlRequest, checkGraphQLResponse, parseResponse } from '../lib/graphql.js';

// Mirrors frontend/app/graphql/queries/pages/challenge.gql with the
// QuizQuestionUserFields fragment inlined.
const CHALLENGE_PAGE_QUERY = `
query ChallengePage($challengeId: ID!) {
  myCurrentProject {
    myPoints
  }
  challenge(id: $challengeId) {
    __typename
    id
    name
    description
    requiresTeamMembership
    requiresSuperTeamMembership
    userEnrolledAt
    userCompletedAt
    ... on SimpleChallenge {
      allowSelfCompletion
    }
    ... on PluginChallenge {
      pluginChallengeId
    }
    ... on ExternalChallenge {
      url
    }
    ... on QuizChallenge {
      quiz {
        id
        name
        description
        timeoutSeconds
        randomizeQuestions
        revealCorrectAnswers
        allowRetakes
        completionPoints
        endTime
        userCanStart
        userActiveSubmission {
          id
        }
        userActiveSession {
          id
          state
        }
        userSubmissions {
          id
          startedAt
          completedAt
          expiresAt
          isExpired
          score
          maxScore
          scorePercentage
          pointsAwarded
          orderedQuestions {
            __typename
            id
            questionText
            questionOrder
            timeoutSeconds
            bettingEnabled
            bettingMinAbsolute
            bettingMaxAbsolute
            bettingMinPercentage
            bettingMaxPercentage
            ... on PredefinedQuestion {
              allowMultipleSelection
              predefinedAnswers {
                id
                answerText
                answerOrder
                isCorrect
                translationStatus {
                  languageCode
                  fields
                }
              }
            }
            ... on NumberQuestion {
              minValue
              maxValue
              stepValue
            }
            ... on OrderingQuestion {
              orderingItems {
                id
                itemText
                correctOrder
              }
            }
          }
          responses {
            __typename
            id
            answeredAt
            timeSpentSeconds
            betAmount
            pointsEarned
            question {
              id
            }
            ... on FreeTextResponse {
              textResponse
            }
            ... on JsonResponse {
              jsonResponse
            }
            ... on NumberResponse {
              numberResponse
            }
            ... on PredefinedResponse {
              isCorrect
              selectedAnswers {
                id
                answerText
                answerOrder
                isCorrect
              }
            }
            ... on OrderingResponse {
              isCorrect
              submittedOrder
            }
          }
        }
      }
    }
  }
}
`;

// Mirrors frontend/app/graphql/mutations/challenges.gql — fired by the
// challenge page when opened with ?enroll=true (QR code entry).
const ENROLL_IN_CHALLENGE_MUTATION = `
mutation EnrollInChallenge($challengeId: ID!) {
  enrollInChallenge(challengeId: $challengeId) {
    id
    userEnrolledAt
  }
}
`;

/**
 * Execute the EnrollInChallenge mutation (self-enroll via QR code)
 * @param {string} baseUrl - Base URL of the GraphQL API
 * @param {string} token - JWT token for authorization
 * @param {string} challengeId - ID of the challenge to enroll in
 * @returns {object|null} Enrolled challenge data or null on error
 */
export function enrollInChallenge(baseUrl, token, challengeId) {
    const response = graphqlRequest(baseUrl, ENROLL_IN_CHALLENGE_MUTATION, { challengeId }, token, 'EnrollInChallenge');
    if (checkGraphQLResponse(response, 'EnrollInChallenge')) {
        const data = parseResponse(response);
        return data ? data.enrollInChallenge : null;
    }
    return null;
}

/**
 * Execute the ChallengePage query
 * @param {string} baseUrl - Base URL of the GraphQL API
 * @param {string} token - JWT token for authorization
 * @param {string} challengeId - ID of the challenge to fetch
 * @returns {object} HTTP response
 */
export function challengePage(baseUrl, token, challengeId) {
    const variables = { challengeId };
    const response = graphqlRequest(baseUrl, CHALLENGE_PAGE_QUERY, variables, token, 'ChallengePage');
    checkGraphQLResponse(response, 'ChallengePage');
    return response;
}

// k6's error code for a request that hit its timeout
const REQUEST_TIMEOUT = 1050;

/**
 * The ChallengePage refetch a Firestore ping triggers. Today's frontend aborts
 * an in-flight refetch when the next ping arrives
 * (frontend/test/unit/firestoreRefetchCoalescing.test.ts), so the request
 * gets the time until that ping as its timeout: k6 then drops it like the
 * browser's abort. Tagged kind=refetch; a cancelled refetch is not an error.
 * @returns {{cancelled: boolean, ok: boolean, data: object|null, error: string|null, durationMs: number}}
 */
export function challengePageRefetch(baseUrl, token, challengeId, timeoutMs) {
    const response = graphqlRequest(baseUrl, CHALLENGE_PAGE_QUERY, { challengeId }, token, 'ChallengePageRefetch', {
        tags: { name: 'ChallengePageRefetch', kind: 'refetch' },
        timeout: `${Math.max(1, Math.round(timeoutMs))}ms`,
    });
    if (response.error_code === REQUEST_TIMEOUT) {
        return { cancelled: true, ok: false, data: null, durationMs: response.timings.duration };
    }
    let body = null;
    try {
        body = JSON.parse(response.body);
    } catch {
        // reported below
    }
    const data = response.status === 200 && body && !body.errors ? body.data : null;
    const error = data ? null
        : `HTTP ${response.status}${response.error ? ` ${response.error}` : ''}${body && body.errors ? ` ${JSON.stringify(body.errors.map((e) => e.message))}` : ''}`;
    return { cancelled: false, ok: Boolean(data), data, error, durationMs: response.timings.duration };
}
