import http from 'k6/http';
import { check, sleep } from 'k6';
import exec from 'k6/execution';

// Test configuration
export const options = {
  vus: 1000, // 1000 virtual users
  duration: '30s', // Run for 30 seconds
  thresholds: {
    http_req_duration: ['p(95)<500'], // 95% of requests must complete below 500ms
    http_req_failed: ['rate<0.01'], // Error rate must be below 1%
  },
};

// GraphQL query for leaderboards with me and projects
const query = `
  query {
    me {
      id
      name
    }
    projects {
      totalCount
      edges {
        node {
          name
          id
          leaderboard(entityType: CHURCHES, first: 10, after: "10") {
            me {
              score
              rank
              name
            }
            pageInfo {
              endCursor
            }
            edges {
              node {
                score
                id
                name
                rank
                isMe
              }
            }
          }
        }
      }
    }
  }
`;

const url = 'http://localhost:8080/graphql';

// Specific user ID for the test
const userID = 'US01K9VZ8663JJ561SR15CFPSKGZ';

// Pre-generate token using the gentoken tool
// This will be executed once during test setup
const token = generateToken(userID);

export default function () {
  const payload = JSON.stringify({
    query: query,
  });

  const params = {
    headers: {
      'Content-Type': 'application/json',
      'Authorization': `Bearer ${token}`,
    },
    tags: {
      user_id: userID,
    },
  };

  const res = http.post(url, payload, params);

  // Check that the request was successful
  check(res, {
    'status is 200': (r) => r.status === 200,
    'no errors': (r) => {
      try {
        const body = JSON.parse(r.body);
        return !body.errors;
      } catch (e) {
        return false;
      }
    },
    'has me data': (r) => {
      try {
        const body = JSON.parse(r.body);
        return body.data && body.data.me && body.data.me.id === userID;
      } catch (e) {
        return false;
      }
    },
    'has projects data': (r) => {
      try {
        const body = JSON.parse(r.body);
        return body.data && body.data.projects && body.data.projects.edges;
      } catch (e) {
        return false;
      }
    },
    'has leaderboard data': (r) => {
      try {
        const body = JSON.parse(r.body);
        const projects = body.data.projects.edges;
        if (projects.length === 0) return true; // No projects is valid
        return projects.every(edge =>
          edge.node.leaderboard &&
          Array.isArray(edge.node.leaderboard.edges)
        );
      } catch (e) {
        return false;
      }
    },
    'has leaderboard me data': (r) => {
      try {
        const body = JSON.parse(r.body);
        const projects = body.data.projects.edges;
        if (projects.length === 0) return true; // No projects is valid
        return projects.every(edge =>
          edge.node.leaderboard &&
          edge.node.leaderboard.me !== null
        );
      } catch (e) {
        return false;
      }
    },
  });

  // Small sleep between iterations (optional, remove for max throughput)
  sleep(0.1);
}

// Helper function to generate JWT token using Go
function generateToken(userId) {
  // This will be replaced by the actual token at runtime
  // Run this command before starting k6:
  // export TOKEN=$(go run cmd/gentoken/main.go US01K9VZ8663JJ561SR15CFPSKGZ)
  const token = __ENV.TOKEN;
  if (!token) {
    throw new Error('TOKEN environment variable not set. Run: export TOKEN=$(go run cmd/gentoken/main.go ' + userId + ')');
  }
  return token;
}
