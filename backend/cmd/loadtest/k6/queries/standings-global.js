import { graphqlRequest, checkGraphQLResponse } from '../lib/graphql.js';

const STANDINGS_GLOBAL_QUERY = `
query StandingsGlobalPage($entityType: LeaderboardEntityType!, $filter: LeaderboardFilter, $first: Int) {
  myCurrentProject {
    id
    leaderboard(entityType: $entityType, filter: $filter, first: $first) {
      edges {
        node {
          id
          name
          description
          score
          rank
          tags
        }
      }
      me {
        id
        name
        description
        score
        rank
        tags
      }
    }
  }
}
`;

/**
 * Execute the StandingsGlobalPage query
 */
export function standingsGlobalPage(baseUrl, token, entityType = 'PERSONS', first = 50) {
    const variables = {
        entityType: entityType,
        first: first,
    };
    const response = graphqlRequest(baseUrl, STANDINGS_GLOBAL_QUERY, variables, token, 'StandingsGlobalPage');
    checkGraphQLResponse(response, 'StandingsGlobalPage');
    return response;
}
