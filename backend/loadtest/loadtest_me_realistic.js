import http from 'k6/http';
import { check, sleep } from 'k6';

// Test configuration - realistic high-concurrency load
// Simulates rapid user influx (e.g., youth camp event start)
export const options = {
  stages: [
    { duration: '10s', target: 3000 }, // Rapid ramp: 0 → 3000 users in 10 seconds
    { duration: '30s', target: 3000 }, // Sustained load at 3000 users
    { duration: '20s', target: 1000 }, // Drop to 1000 users
    { duration: '10s', target: 0 },    // Ramp down
  ],
  thresholds: {
    http_req_duration: ['p(95)<500', 'p(99)<1000'], // 95% under 500ms, 99% under 1s
    http_req_failed: ['rate<0.05'], // Error rate must be below 5%
  },
};

// Real user IDs from the database
const userIDs = [
  'US01K8XV6EN42TPEATDT708X51KE',
  'US01K8XV6EPTGRWZPZKV3SZ3MTRP',
  'US01K8XV6ERG3MF9AJBJFGDPVE2V',
  'US01K8XV6ET6PNZKPFHA0DKDPZ2T',
  'US01K8XV6EVTEFD65RGJJB05WX5J',
  'US01K8XV6EXEY4EBTF1TMR0810H6',
  'US01K8XV6EZ2J1X50A2H6P7DZ603',
  'US01K8XV6F0PWEY5AH5QKTQXK4R1',
  'US01K8XV6F2EPF5RCPJKJMDP5W5P',
  'US01K8XV6F44F7NB36Y04RE2PV3C',
  'US01K8XV6F5QN5MT7MQMXTDXDFEK',
  'US01K8XV6F7DD1D45P82XH3ZM475',
  'US01K8XV6F92EGGBAQ4W4S8AM2DG',
  'US01K8XV6FAP0A24Y63ADQRG2438',
  'US01K8XV6FCB4181JPQMNW0KKH3E',
  'US01K8XV6FE12VEMWJ6AG6P7DVC7',
  'US01K8XV6FFQB2HWCWJ31Z83XTQ5',
  'US01K8XV6FHCE50SK3NMMKXDQDB5',
  'US01K8XV6FK27NPRX2T624S6X536',
  'US01K8XV6FMPEBDQ4A0K71DWK64A',
  'US01K8XV6FPBZ9BERE7T81V6XRQ9',
  'US01K8XV6FR0ENFDZNMWGVBY64PT',
  'US01K8XV6FSMP0H1RMMY7RM8QNZG',
  'US01K8XV6FV9XV9BR9GBJ4HPGX21',
  'US01K8XV6FWY3MVWPRH6X8D7NJZM',
  'US01K8XV6FYKGW34RSJQ4RCX80DY',
  'US01K8XV6G09YPN7YRR1HZE22R0G',
  'US01K8XV6G1Z6XAYJ3J069JB592M',
  'US01K8XV6G3NY703WZKDSWXHQREP',
  'US01K8XV6G5AWG0VKX1VYWTXJ4S9',
  'US01K8XV6G6Z5ZJH2RQVJT9RM4CY',
  'US01K8XV6G8N833MHEYH7HGRB3SV',
  'US01K8XV6GABX9EPSA2PRNSWYERJ',
  'US01K8XV6GC1PBVV2BZ72T2QPQSM',
];

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

// TODO: Generate JWT tokens for each user
// Map of user ID to JWT token
const tokensByUserID = {};
userIDs.forEach(userID => {
  tokensByUserID[userID] = 'YOUR_JWT_TOKEN_HERE';
});

export default function () {
  // Select a user ID for this virtual user
  // Each VU gets assigned to a user and sticks with it
  const userIndex = (__VU - 1) % userIDs.length;
  const userID = userIDs[userIndex];
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

  // Realistic think time between requests (0.5-2 seconds)
  sleep(0.5 + Math.random() * 1.5);
}
