import { SharedArray } from 'k6/data';

import { userJourney } from './lib/journey.js';

const config = JSON.parse(open('../config.json'));
const tokens = new SharedArray('tokens', function () {
    return config.tokens;
});
const baseUrl = config.baseUrl;

// Spike test: fast ramp to high VUs
export const options = {
    scenarios: {
        spike_test: {
            executor: 'ramping-vus',
            startVUs: 0,
            stages: [
                { duration: '5s', target: 5000 },
                { duration: __ENV.SPIKE_HOLD || '2m', target: 5000 },
                { duration: '5s', target: 0 },
            ],
        },
    },
    thresholds: {
        http_req_duration: ['p(95)<500', 'p(99)<1000'],
        http_req_failed: ['rate<0.01'],
        graphql_errors: ['rate<0.01'],
    },
};

function getRandomToken() {
    return tokens[Math.floor(Math.random() * tokens.length)];
}

export default function () {
    const { token } = getRandomToken();
    userJourney(baseUrl, token);
}

export function setup() {
    console.log(`Spike test: ${tokens.length} users, target ${baseUrl}`);
}
