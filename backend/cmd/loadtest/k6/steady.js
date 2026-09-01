import { SharedArray } from 'k6/data';

import { userJourney } from './lib/journey.js';

// open()+JSON.parse() must happen inside the SharedArray callback — see
// freetext-quiz-spike.js for why parsing outside it blows up per-VU RAM.
const tokens = new SharedArray('tokens', function () {
    return JSON.parse(open('../config.json')).tokens;
});
// Simulated Auth0 tokens (tokengen -auth0-count) for the REALISM auth-dance
// fraction. SharedArray callbacks must return a non-empty array, so a null
// placeholder marks "not generated".
const auth0Tokens = new SharedArray('auth0Tokens', function () {
    const parsed = JSON.parse(open('../config.json')).auth0Tokens;
    return parsed && parsed.length > 0 ? parsed : [null];
});
const baseUrl = new SharedArray('baseUrl', function () {
    return [JSON.parse(open('../config.json')).baseUrl];
})[0];

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

function getRandomAuth0Token() {
    const entry = auth0Tokens[Math.floor(Math.random() * auth0Tokens.length)];
    return entry ? entry.token : null;
}

export default function () {
    const { token } = getRandomToken();
    userJourney(baseUrl, token, getRandomAuth0Token());
}

export function setup() {
    console.log(`Steady load: ${tokens.length} users, target ${baseUrl}`);
}
