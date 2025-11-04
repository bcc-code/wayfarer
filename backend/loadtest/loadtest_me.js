import http from 'k6/http';
import { check, sleep } from 'k6';
import { tokensByUserID } from './tokens.js';

// Test configuration
export const options = {
  vus: 100, // 10 virtual users
  duration: '20s', // Run for 30 seconds
  thresholds: {
    http_req_duration: ['p(95)<100'], // 95% of requests must complete below 100ms
    http_req_failed: ['rate<0.01'], // Error rate must be below 1%
  },
};

// GraphQL query for the me endpoint
const query = `
  query {
    projects {
      id
    }
    me {
      roles {
        id
        assignedBy {
          name
        }
        role
        assignedAt
        scope {
          id
          type
          team {
            name
            id
          }
          project {
            name
            id
            description
          }
        }
        user {
          id
        }
      }
      id
      name
      email
      church {
        id
        name
        country
      }
      projects {
        id
        name
        description
        branding {
          logo
          colors {
            primary
            secondary
            tertiary
          }
        }
      }
    }
  }
`;

const url = 'http://localhost:8080/graphql';

// Get a random user ID for this virtual user
const userIDs = Object.keys(tokensByUserID);

export default function () {
  // Select a random user for each request
  const userID = userIDs[Math.floor(Math.random() * userIDs.length)];
  const token = tokensByUserID[userID];

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
    'has user data': (r) => {
      try {
        const body = JSON.parse(r.body);
        return body.data && body.data.me && body.data.me.id;
      } catch (e) {
        return false;
      }
    },
    'has church data': (r) => {
      try {
        const body = JSON.parse(r.body);
        return body.data && body.data.me && body.data.me.church;
      } catch (e) {
        return false;
      }
    },
    'has projects data': (r) => {
      try {
        const body = JSON.parse(r.body);
        return body.data && body.data.me && Array.isArray(body.data.me.projects);
      } catch (e) {
        return false;
      }
    },
    'has roles data': (r) => {
      try {
        const body = JSON.parse(r.body);
        return body.data && body.data.me && Array.isArray(body.data.me.roles);
      } catch (e) {
        return false;
      }
    },
    'has root projects': (r) => {
      try {
        const body = JSON.parse(r.body);
        return body.data && Array.isArray(body.data.projects);
      } catch (e) {
        return false;
      }
    },
  });

  // Small sleep between iterations (optional, remove for max throughput)
  sleep(0.1);
}
