import { graphqlRequest, checkGraphQLResponse } from '../lib/graphql.js';

const CHALLENGE_PAGE_QUERY = `
query ChallengePage($challengeId: ID!) {
  challenge(id: $challengeId) {
    __typename
    id
    name
    description
    userEnrolledAt
    userCompletedAt
    ... on SimpleChallenge {
      allowSelfCompletion
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
        publishedAt
        endTime
        userCanStart
        userActiveSubmission {
          id
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
          orderedQuestions {
            __typename
            id
            questionText
            questionOrder
            timeoutSeconds
            ... on NumberQuestion {
              minValue
              maxValue
              stepValue
            }
            ... on PredefinedQuestion {
              allowMultipleSelection
              predefinedAnswers {
                id
                answerText
                answerOrder
                isCorrect
              }
            }
          }
          responses {
            __typename
            id
            answeredAt
            timeSpentSeconds
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
