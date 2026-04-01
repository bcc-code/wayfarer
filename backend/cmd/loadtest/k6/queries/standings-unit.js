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
        score
        rank
        tags
      }
      superTeam {
        id
        name
        color
        imageObject {
          url
          blurhash
        }
      }
    }
  }
}
`;

/**
 * Execute the StandingsUnitPage query
 */
export function standingsUnitPage(baseUrl, token) {
    const response = graphqlRequest(baseUrl, STANDINGS_UNIT_QUERY, {}, token, 'StandingsUnitPage');
    checkGraphQLResponse(response, 'StandingsUnitPage');
    return response;
}
