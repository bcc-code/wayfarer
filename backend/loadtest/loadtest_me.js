import http from 'k6/http';
import { check, sleep } from 'k6';

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
    me {
      id
      name
      email
      age
      gender
      image
      church {
        id
        name
        country
        category
      }
      projects {
        id
        name
        description
        startDate
        endDate
        branding {
          logo
          colors {
            primary
            secondary
            tertiary
          }
          rounding
        }
      }
    }
  }
`;

const url = 'http://localhost:8080/graphql/user';

// TODO: Replace with a valid JWT token for testing
const token = 'YOUR_JWT_TOKEN_HERE';

export default function () {
  const payload = JSON.stringify({
    query: query,
  });

  const params = {
    headers: {
      'Content-Type': 'application/json',
      'Authorization': `Bearer ${token}`,
    },
  };

  const res = http.post(url, payload, params);

  // Check that the request was successful
  check(res, {
    'status is 200': (r) => r.status === 200,
    'no errors': (r) => {
      const body = JSON.parse(r.body);
      return !body.errors;
    },
    'has user data': (r) => {
      const body = JSON.parse(r.body);
      return body.data && body.data.me && body.data.me.id;
    },
    'has church data': (r) => {
      const body = JSON.parse(r.body);
      return body.data && body.data.me && body.data.me.church;
    },
    'has projects data': (r) => {
      const body = JSON.parse(r.body);
      return body.data && body.data.me && Array.isArray(body.data.me.projects);
    },
  });

  // Small sleep between iterations (optional, remove for max throughput)
  sleep(0.1);
}
