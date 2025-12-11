import { SharedArray } from 'k6/data';
import { sleep } from 'k6';

import { challengesPage } from './queries/challenges.js';
import { profilePage } from './queries/profile.js';
import { standingsGlobalPage } from './queries/standings-global.js';
import { standingsLocalPage } from './queries/standings-local.js';
import { standingsUnitPage } from './queries/standings-unit.js';

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
                { duration: '5s', target: parseInt(__ENV.SPIKE_PEAK) || 500 },
                { duration: __ENV.SPIKE_HOLD || '2m', target: parseInt(__ENV.SPIKE_PEAK) || 500 },
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
    const rand = Math.random();

    if (rand < 0.30) {
        challengesPage(baseUrl, token);
    } else if (rand < 0.50) {
        profilePage(baseUrl, token);
    } else if (rand < 0.70) {
        standingsGlobalPage(baseUrl, token);
    } else if (rand < 0.85) {
        standingsLocalPage(baseUrl, token);
    } else {
        standingsUnitPage(baseUrl, token);
    }

    sleep(Math.random() * 0.5 + 0.1);
}

export function setup() {
    console.log(`Spike test: ${tokens.length} users, target ${baseUrl}`);
}
