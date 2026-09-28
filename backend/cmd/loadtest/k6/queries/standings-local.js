import { graphqlRequest, checkGraphQLResponse } from '../lib/graphql.js';

const STANDINGS_LOCAL_QUERY = `
query StandingsLocalPage($filter: LeaderboardFilter, $first: Int) {
  me {
    church {
      id
      name
    }
  }
  myCurrentProject {
    id
    personLeaderboard: leaderboard(entityType: PERSONS, filter: $filter, first: $first) {
      totalCount
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
    unitLeaderboard: leaderboard(entityType: TEAMS, filter: $filter, first: $first) {
      totalCount
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
 * Execute the StandingsLocalPage query (dual leaderboard: persons + teams)
 */
export function standingsLocalPage(baseUrl, token, first = 50) {
    const variables = {
        first: first,
    };
    const response = graphqlRequest(baseUrl, STANDINGS_LOCAL_QUERY, variables, token, 'StandingsLocalPage');
    checkGraphQLResponse(response, 'StandingsLocalPage');
    return response;
}
