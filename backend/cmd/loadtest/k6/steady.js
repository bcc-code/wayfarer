import { SharedArray } from 'k6/data';

import { userJourney } from './lib/journey.js';

const config = JSON.parse(open('../config.json'));
const tokens = new SharedArray('tokens', function () {
    return config.tokens;
});
const baseUrl = config.baseUrl;

// Steady load: fast ramp then hold
export const options = {
    scenarios: {
        steady_load: {
            executor: 'ramping-vus',
            startVUs: 0,
            stages: [
                { duration: '5s', target: parseInt(__ENV.STEADY_VUS) || 50 },
                { duration: __ENV.DURATION || '5m', target: parseInt(__ENV.STEADY_VUS) || 50 },
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
    console.log(`Steady load: ${tokens.length} users, target ${baseUrl}`);
}
