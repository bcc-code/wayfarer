import http from 'k6/http';
import { check, sleep } from 'k6';
import { tokensByUserID } from './tokens.js';

// Extreme load test - simulates event start with thousands of concurrent users
// This tests system behavior under maximum realistic load
export const options = {
  stages: [
    { duration: '5s', target: 1000 },  // Quick ramp to 1000
    { duration: '5s', target: 10000 },  // Extreme spike to 5000 users in 5s
    { duration: '20s', target: 5000 }, // Sustain 5000 users
    { duration: '10s', target: 2000 }, // Drop to 2000
    { duration: '10s', target: 0 },    // Ramp down
  ],
  thresholds: {
    http_req_duration: ['p(95)<1000', 'p(99)<2000'], // More lenient under extreme load
    http_req_failed: ['rate<0.10'], // Allow up to 10% errors during spike
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

// GraphQL query for extreme load testing
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

export default function () {
  // Each VU picks a user and sticks with it (simulates real user behavior)
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
      scenario: 'extreme_load',
    },
  };

  const res = http.post(url, payload, params);

  // Relaxed checks for extreme load
  check(res, {
    'status is 200': (r) => r.status === 200,
    'response received': (r) => r.body && r.body.length > 0,
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
  });

  // Minimal sleep - aggressive load
  sleep(0.05 + Math.random() * 0.1); // 50-150ms between requests
}
