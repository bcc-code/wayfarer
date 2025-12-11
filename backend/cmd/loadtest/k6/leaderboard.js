import { SharedArray } from 'k6/data';

import { standingsGlobalPage } from './queries/standings-global.js';
import { standingsLocalPage } from './queries/standings-local.js';
import { standingsUnitPage } from './queries/standings-unit.js';

const config = JSON.parse(open('../config.json'));
const tokens = new SharedArray('tokens', function () {
    return config.tokens;
});
const baseUrl = config.baseUrl;

// Leaderboard stress: constant arrival rate
export const options = {
    scenarios: {
        leaderboard_stress: {
            executor: 'constant-arrival-rate',
            rate: parseInt(__ENV.LEADERBOARD_RPS) || 100,
            timeUnit: '1s',
            duration: __ENV.LEADERBOARD_DURATION || '5m',
            preAllocatedVUs: parseInt(__ENV.LEADERBOARD_RPS) || 100,
            maxVUs: 1000,
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

    if (rand < 0.50) {
        standingsGlobalPage(baseUrl, token);
    } else if (rand < 0.80) {
        standingsLocalPage(baseUrl, token);
    } else {
        standingsUnitPage(baseUrl, token);
    }
}

export function setup() {
    console.log(`Leaderboard stress: ${tokens.length} users, target ${baseUrl}`);
}
