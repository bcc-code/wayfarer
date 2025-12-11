import { graphqlRequest, checkGraphQLResponse } from '../lib/graphql.js';

const STANDINGS_LOCAL_QUERY = `
query StandingsLocalPage($entityType: LeaderboardEntityType!, $filter: LeaderboardFilter) {
  me {
    church {
      id
      name
    }
  }
  myCurrentProject {
    id
    leaderboard(entityType: $entityType, filter: $filter) {
      edges {
        node {
          id
          name
          score
          rank
          tags
        }
      }
      me {
        id
        name
        score
        rank
        tags
      }
    }
  }
}
`;

/**
 * Execute the StandingsLocalPage query
 * @param {string} baseUrl - Base URL of the GraphQL API
 * @param {string} token - JWT token for authorization
 * @param {string} entityType - Leaderboard entity type (PERSONS, TEAMS)
 * @returns {object} HTTP response
 */
export function standingsLocalPage(baseUrl, token, entityType = 'PERSONS') {
    const variables = {
        entityType: entityType,
    };
    const response = graphqlRequest(baseUrl, STANDINGS_LOCAL_QUERY, variables, token, 'StandingsLocalPage');
    checkGraphQLResponse(response, 'StandingsLocalPage');
    return response;
}
