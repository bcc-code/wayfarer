import { graphqlRequest, checkGraphQLResponse } from '../lib/graphql.js';

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
