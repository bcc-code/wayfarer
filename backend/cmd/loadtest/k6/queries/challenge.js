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
