import { graphqlRequest, checkGraphQLResponse } from '../lib/graphql.js';

const STANDINGS_UNIT_QUERY = `
query StandingsUnitPage {
  myCurrentProject {
    id
    myTeam {
      id
      name
      memberLeaderboard {
        id
        name
        tags
        rank
        score
      }
    }
  }
}
`;

/**
 * Execute the StandingsUnitPage query
 * @param {string} baseUrl - Base URL of the GraphQL API
 * @param {string} token - JWT token for authorization
 * @returns {object} HTTP response
 */
export function standingsUnitPage(baseUrl, token) {
    const response = graphqlRequest(baseUrl, STANDINGS_UNIT_QUERY, {}, token, 'StandingsUnitPage');
    checkGraphQLResponse(response, 'StandingsUnitPage');
    return response;
}
