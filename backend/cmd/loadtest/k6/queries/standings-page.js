import { graphqlRequest, checkGraphQLResponse } from '../lib/graphql.js';

// Wrapper query fired by the standings page itself (pages/standings.vue),
// before the selected tab component runs its own query.
const STANDINGS_PAGE_QUERY = `
query StandingsPage {
  myCurrentProject {
    myTeam {
      id
    }
  }
}
`;

/**
 * Execute the StandingsPage query
 */
export function standingsPage(baseUrl, token) {
    const response = graphqlRequest(baseUrl, STANDINGS_PAGE_QUERY, {}, token, 'StandingsPage');
    checkGraphQLResponse(response, 'StandingsPage');
    return response;
}
