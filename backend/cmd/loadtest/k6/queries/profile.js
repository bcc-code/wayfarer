import { graphqlRequest, checkGraphQLResponse } from '../lib/graphql.js';

const PROFILE_PAGE_QUERY = `
query ProfilePage {
  me {
    id
    name
    consentStatus {
      pendingConsents {
        __typename
        id
        key
        version
        title
        body {
          html
        }
        url
        managementType
        managedBy
      }
    }
  }
  myCurrentProject {
    id
    name
    achievements {
      id
      name
      descriptionPending
      descriptionCompleted
      imagePending
      imageCompleted
      hidden
      achievedAt
      points
    }
    leaderboard(entityType: PERSONS) {
      me {
        score
        rank
      }
    }
  }
}
`;

/**
 * Execute the ProfilePage query
 * @param {string} baseUrl - Base URL of the GraphQL API
 * @param {string} token - JWT token for authorization
 * @returns {object} HTTP response
 */
export function profilePage(baseUrl, token) {
    const response = graphqlRequest(baseUrl, PROFILE_PAGE_QUERY, {}, token, 'ProfilePage');
    checkGraphQLResponse(response, 'ProfilePage');
    return response;
}
