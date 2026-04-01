import { graphqlRequest, checkGraphQLResponse } from '../lib/graphql.js';

const ACTIVE_CHALLENGES_PAGE_QUERY = `
query ActiveChallengesPage {
  myCurrentProject {
    myTeam {
      joinCode
    }
    activeChallenges {
      __typename
      id
      name
      description
      imageObject {
        url
        width
        height
        blurhash
      }
      buttonText
      notificationText
      publishedAt
      visibleAt
      startedAt
      endTime
      requiresTeamMembership
      requiresSuperTeamMembership
      userCompletedAt
      userEnrolledAt
      ... on SimpleChallenge {
        allowSelfCompletion
      }
      ... on ExternalChallenge {
        url
      }
      ... on QuizChallenge {
        quiz {
          timeoutSeconds
          randomizeQuestions
          revealCorrectAnswers
          allowRetakes
          completionPoints
          endTime
          userCanStart
          userActiveSubmission {
            id
            startedAt
            completedAt
            expiresAt
            isExpired
            autoSubmitted
            score
            maxScore
            scorePercentage
            pointsAwarded
          }
          userActiveSession {
            id
            state
            openAt
            lockAt
            finishAt
            userHasAccess
          }
          userSubmissions {
            id
            startedAt
            completedAt
            expiresAt
            isExpired
            autoSubmitted
            score
            maxScore
            scorePercentage
            pointsAwarded
          }
        }
      }
    }
  }
}
`;

const COMPLETED_CHALLENGES_PAGE_QUERY = `
query CompletedChallengesPage {
  myCurrentProject {
    completedChallenges {
      __typename
      id
      name
      description
      imageObject {
        url
        width
        height
        blurhash
      }
      buttonText
      notificationText
      publishedAt
      visibleAt
      startedAt
      endTime
      requiresTeamMembership
      requiresSuperTeamMembership
      userCompletedAt
      userEnrolledAt
      ... on SimpleChallenge {
        allowSelfCompletion
      }
      ... on ExternalChallenge {
        url
      }
      ... on QuizChallenge {
        quiz {
          timeoutSeconds
          randomizeQuestions
          revealCorrectAnswers
          allowRetakes
          completionPoints
          endTime
          userCanStart
          userActiveSubmission {
            id
            startedAt
            completedAt
            expiresAt
            isExpired
            autoSubmitted
            score
            maxScore
            scorePercentage
            pointsAwarded
          }
          userActiveSession {
            id
            state
            openAt
            lockAt
            finishAt
            userHasAccess
          }
          userSubmissions {
            id
            startedAt
            completedAt
            expiresAt
            isExpired
            autoSubmitted
            score
            maxScore
            scorePercentage
            pointsAwarded
          }
        }
      }
    }
  }
}
`;

/**
 * Execute the ActiveChallengesPage query
 */
export function activeChallengesPage(baseUrl, token) {
    const response = graphqlRequest(baseUrl, ACTIVE_CHALLENGES_PAGE_QUERY, {}, token, 'ActiveChallengesPage');
    checkGraphQLResponse(response, 'ActiveChallengesPage');
    return response;
}

/**
 * Execute the CompletedChallengesPage query
 */
export function completedChallengesPage(baseUrl, token) {
    const response = graphqlRequest(baseUrl, COMPLETED_CHALLENGES_PAGE_QUERY, {}, token, 'CompletedChallengesPage');
    checkGraphQLResponse(response, 'CompletedChallengesPage');
    return response;
}
