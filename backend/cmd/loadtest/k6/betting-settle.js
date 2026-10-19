import { SharedArray } from 'k6/data';
import { Trend } from 'k6/metrics';

import { parseResponse } from './lib/graphql.js';
import { challengePage } from './queries/challenge.js';
import { lockAndFinishSession } from './queries/betting.js';

// Settlement step of a sharded betting-quiz-rampspike.js run: once every
// shard has finished (each ran with SETTLE=0), lock and finish the session as
// the admin. Finishing settles every bet inside the request, so its duration
// is the settlement time an admin would see.

const meta = new SharedArray('meta', function () {
    const parsed = JSON.parse(open('../config.json'));
    return [{
        baseUrl: parsed.baseUrl,
        userToken: parsed.tokens[0].token,
        adminToken: parsed.adminToken ? parsed.adminToken.token : null,
    }];
})[0];
const baseUrl = __ENV.BASE_URL || meta.baseUrl;
const challengeId = __ENV.CHALLENGE_ID || 'CL01LOADTESTBETTING000000000';
const finishTimeout = __ENV.FINISH_TIMEOUT || '30m';

const settlementDuration = new Trend('settlement_duration', true);

export const options = {
    scenarios: {
        settle: { executor: 'per-vu-iterations', vus: 1, iterations: 1, maxDuration: '40m' },
    },
    thresholds: {
        'checks{check:FinishQuizSession: status is 200}': ['rate==1'],
    },
};

export default function () {
    if (!meta.adminToken) {
        throw new Error('config.json has no adminToken: run tokengen with -admin-user');
    }
    // The session is found through a user's view of the challenge, like the
    // rampspike's setup()
    const data = parseResponse(challengePage(baseUrl, meta.userToken, challengeId));
    const session = data && data.challenge && data.challenge.quiz && data.challenge.quiz.userActiveSession;
    if (!session) {
        throw new Error(`No OPEN session for challenge ${challengeId}`);
    }
    console.log(`Finishing session ${session.id} (settles all bets, timeout ${finishTimeout})...`);
    const res = lockAndFinishSession(baseUrl, meta.adminToken, session.id, finishTimeout);
    settlementDuration.add(res.finishMs);
    console.log(
        `Session finished: locked=${res.locked} finished=${res.finished} HTTP ${res.status} ` +
        `in ${(res.finishMs / 1000).toFixed(1)}s`
    );
}
