import { graphqlRequest, checkGraphQLResponse } from '../lib/graphql.js';

const CHALLENGES_PAGE_QUERY = `
query ChallengesPage {
  myCurrentProject {
    challenges {
      id
      name
      description
      image
      buttonText
      publishedAt
      endTime
      visibleAt
    }
  }
}
`;

/**
 * Execute the ChallengesPage query
 * @param {string} baseUrl - Base URL of the GraphQL API
 * @param {string} token - JWT token for authorization
 * @returns {object} HTTP response
 */
export function challengesPage(baseUrl, token) {
    const response = graphqlRequest(baseUrl, CHALLENGES_PAGE_QUERY, {}, token, 'ChallengesPage');
    checkGraphQLResponse(response, 'ChallengesPage');
    return response;
}
