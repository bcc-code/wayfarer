import http from 'k6/http';
import { check, sleep } from 'k6';
import { Counter, Trend } from 'k6/metrics';
import { tokensByUserID, allUserIDs } from './all_user_tokens.js';

// Custom metrics for tracking per-query performance
const queryDuration = new Trend('query_duration', true);
const queryErrors = new Counter('query_errors');

// Test configuration - realistic high-concurrency load
// Simulates rapid user influx (e.g., youth camp event start)
export const options = {
  stages: [
    { duration: '5s', target: 3000 },  // Ramp up to 3000 users over 2 minutes
    { duration: '1m', target: 7000 },  // Sustained load at 3000 users for 5 minutes
    { duration: '1m', target: 0 },     // Ramp down to 0
  ],
  thresholds: {
    http_req_duration: ['p(95)<500', 'p(99)<1000'], // 95% under 500ms, 99% under 1s
    http_req_failed: ['rate<0.05'], // Error rate must be below 5%
    'query_duration{query:StandingsPage}': ['p(95)<500'],
    'query_duration{query:ChallengesPage}': ['p(95)<500'],
    'query_duration{query:ProfilePage}': ['p(95)<500'],
    'query_duration{query:UnitPage}': ['p(95)<500'],
    'query_duration{query:AdminHomePage}': ['p(95)<500'],
    'query_duration{query:AdminProjectsPage}': ['p(95)<500'],
    'query_duration{query:AdminUsersPage}': ['p(95)<500'],
    'query_duration{query:GetMe}': ['p(95)<500'],
    'query_duration{query:CurrentProject}': ['p(95)<500'],
    'query_duration{query:AdminSidebar}': ['p(95)<500'],
  },
};

// All GraphQL queries from the frontend (read-only)
const queries = {
  StandingsPage: {
    query: `
      query StandingsPage {
        myCurrentProject {
          id
          leaderboard(type: TOTAL) {
            name
            description
            score
            image
          }
        }
      }
    `,
    variables: null,
  },

  ChallengesPage: {
    query: `
      query ChallengesPage {
        myCurrentProject {
          challenges {
            id
            name
            description
            userCompletedAt
            image
            url
            buttonText
            publishedAt
            endTime
          }
        }
      }
    `,
    variables: null,
  },

  ProfilePage: {
    query: `
      query ProfilePage {
        me {
          id
          name
          image
          church {
            id
            name
          }
          projects {
            id
            achievements {
              id
              name
              image
              hidden
              achievedAt
              points
            }
          }
        }
      }
    `,
    variables: null,
  },

  UnitPage: {
    query: `
      query UnitPage {
        myCurrentProject {
          id
          myTeam {
            id
            name
            superTeam {
              id
              name
            }
            leaderboard(type: TOTAL) {
              name
              description
              score
              image
            }
          }
        }
      }
    `,
    variables: null,
  },

  AdminHomePage: {
    query: `
      query AdminHomePage {
        me {
          id
          name
        }
        projects {
          edges {
            node {
              id
              name
              description
              endDate
              startDate
              branding {
                logo
                rounding
                colors {
                  primary
                  secondary
                  tertiary
                }
              }
            }
          }
        }
      }
    `,
    variables: null,
  },

  AdminProjectsPage: {
    query: `
      query AdminProjectsPage {
        projects {
          edges {
            node {
              id
              name
              description
              endDate
              startDate
              branding {
                logo
                colors {
                  primary
                }
              }
            }
          }
        }
      }
    `,
    variables: null,
  },

  AdminUsersPage: {
    query: `
      query AdminUsersPage($filter: UserFilter, $first: Int, $after: String, $last: Int, $before: String) {
        users(
          filter: $filter
          first: $first
          after: $after
          last: $last
          before: $before
        ) {
          totalCount
          pageInfo {
            hasNextPage
            hasPreviousPage
            startCursor
            endCursor
          }
          edges {
            cursor
            node {
              id
              name
              email
              image
              church {
                name
              }
              roles {
                id
                role
              }
            }
          }
        }
      }
    `,
    variables: {
      first: 15,
    },
  },

  GetMe: {
    query: `
      query GetMe {
        me {
          id
          name
          email
          image
          membersId
          church {
            id
            name
            country
            category
          }
          gender
          birthdate
          roles {
            id
            role
            scope {
              id
              type
              church {
                id
              }
              team {
                id
              }
              project {
                id
              }
            }
          }
        }
      }
    `,
    variables: null,
  },

  CurrentProject: {
    query: `
      query CurrentProject {
        myCurrentProject {
          branding {
            logo
            colors {
              primary
            }
            rounding
          }
        }
      }
    `,
    variables: null,
  },

  AdminSidebar: {
    query: `
      query AdminSidebar {
        projects {
          edges {
            node {
              id
              name
              endDate
              startDate
            }
          }
        }
      }
    `,
    variables: null,
  },
};

// Get array of query names for random selection
// Exclude admin queries since no users have admin role in test data
const allQueryNames = Object.keys(queries);
const queryNames = allQueryNames.filter(name => !name.startsWith('Admin'));

const url = 'http://localhost:8080/graphql';

export default function () {
  // Randomly select a user for this iteration
  const userID = allUserIDs[Math.floor(Math.random() * allUserIDs.length)];
  const token = tokensByUserID[userID];

  // Randomly select a query to execute
  const queryName = queryNames[Math.floor(Math.random() * queryNames.length)];
  const queryConfig = queries[queryName];

  const payload = JSON.stringify({
    query: queryConfig.query,
    variables: queryConfig.variables,
  });

  const params = {
    headers: {
      'Content-Type': 'application/json',
      'Authorization': `Bearer ${token}`,
    },
    tags: {
      user_id: userID,
      query_name: queryName,
    },
  };

  const startTime = Date.now();
  const res = http.post(url, payload, params);
  const duration = Date.now() - startTime;

  // Record custom metrics
  queryDuration.add(duration, { query: queryName });

  // Check that the request was successful
  const checks = check(res, {
    'status is 200': (r) => r.status === 200,
    'no errors': (r) => {
      try {
        const body = JSON.parse(r.body);
        return !body.errors;
      } catch (e) {
        return false;
      }
    },
    'has data property': (r) => {
      try {
        const body = JSON.parse(r.body);
        // GraphQL always returns { data: ... }, even if fields are null
        // This is valid when users lack context (no project, no team, etc.)
        return 'data' in body;
      } catch (e) {
        return false;
      }
    },
  });

  // Only count as error if we got GraphQL errors or non-200 status
  if (!checks['status is 200'] || !checks['no errors']) {
    queryErrors.add(1, { query: queryName });
  }

  // Realistic think time between requests (0.5-2 seconds)
  sleep(0.5 + Math.random() * 1.5);
}

// Summary report at the end
export function handleSummary(data) {
  return {
    'stdout': textSummary(data, { indent: ' ', enableColors: true }),
  };
}

function textSummary(data, config) {
  const indent = config.indent || '';
  const enableColors = config.enableColors || false;

  let summary = `\n${indent}Random Query Load Test Summary\n`;
  summary += `${indent}================================\n\n`;

  // Overall metrics
  const httpReqs = data.metrics.http_reqs.values.count;
  const httpFailRate = data.metrics.http_req_failed.values.rate;
  const httpDurationP95 = data.metrics.http_req_duration.values['p(95)'];
  const httpDurationP99 = data.metrics.http_req_duration.values['p(99)'];

  summary += `${indent}Overall:\n`;
  summary += `${indent}  Total Requests: ${httpReqs}\n`;
  summary += `${indent}  Error Rate: ${(httpFailRate * 100).toFixed(2)}%\n`;
  summary += `${indent}  P95 Duration: ${httpDurationP95.toFixed(2)}ms\n`;
  summary += `${indent}  P99 Duration: ${httpDurationP99.toFixed(2)}ms\n\n`;

  // Per-query metrics
  summary += `${indent}Per-Query Performance:\n`;
  for (const queryName of queryNames) {
    const metricName = `query_duration{query:${queryName}}`;
    const metric = data.metrics[metricName];

    if (metric && metric.values) {
      const count = metric.values.count;
      const avg = metric.values.avg;
      const p95 = metric.values['p(95)'];

      summary += `${indent}  ${queryName}:\n`;
      summary += `${indent}    Count: ${count}\n`;
      summary += `${indent}    Avg: ${avg.toFixed(2)}ms\n`;
      summary += `${indent}    P95: ${p95.toFixed(2)}ms\n`;
    }
  }

  summary += `\n${indent}Test completed successfully!\n`;

  return summary;
}
