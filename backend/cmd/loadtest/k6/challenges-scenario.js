import { sleep } from 'k6';
import { SharedArray } from 'k6/data';

import { coldLoad } from './queries/bootstrap.js';
import { activeChallengesPage } from './queries/challenges.js';
import { clickRandomChallenge } from './lib/journey.js';

// open()+JSON.parse() must happen inside the SharedArray callback — see
// freetext-quiz-spike.js for why parsing outside it blows up per-VU RAM.
const tokens = new SharedArray('tokens', function () {
    return JSON.parse(open('../config.json')).tokens;
});
const baseUrl = new SharedArray('baseUrl', function () {
    return [JSON.parse(open('../config.json')).baseUrl];
})[0];

// Challenges page focus: users cold-load /challenges, scan the list,
// click one challenge and leave the app.
export const options = {
    scenarios: {
        challenges_page: {
            executor: 'ramping-vus',
            startVUs: 0,
            stages: [
                { duration: '5s', target: parseInt(__ENV.CHALLENGES_VUS) || 50 },
                { duration: __ENV.DURATION || '5m', target: parseInt(__ENV.CHALLENGES_VUS) || 50 },
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

    // Cold load of /challenges: GetMe + CurrentProject + ActiveChallengesPage
    // + GetFirebaseToken
    const response = coldLoad(baseUrl, token, () => activeChallengesPage(baseUrl, token));

    // User scans the list before clicking
    sleep(Math.random() * 3 + 2);

    // Click one challenge: external -> leaves the app with no request,
    // otherwise the detail page fires ChallengePage. Session ends here.
    clickRandomChallenge(baseUrl, token, response);
}

export function setup() {
    console.log(`Challenges page load: ${tokens.length} users, target ${baseUrl}`);
}
