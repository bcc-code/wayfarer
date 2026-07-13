import { sleep } from 'k6';

import { coldLoad } from '../queries/bootstrap.js';
import { profilePage } from '../queries/profile.js';
import { activeChallengesPage, completedChallengesPage } from '../queries/challenges.js';
import { challengePage } from '../queries/challenge.js';
import { standingsPage } from '../queries/standings-page.js';
import { standingsGlobalPage } from '../queries/standings-global.js';
import { standingsLocalPage } from '../queries/standings-local.js';
import { standingsUnitPage } from '../queries/standings-unit.js';
import { parseResponse } from './graphql.js';

/**
 * Pick a random standings page variant (global/local/unit)
 */
function randomStandingsPage(baseUrl, token) {
    const rand = Math.random();
    if (rand < 0.40) {
        standingsGlobalPage(baseUrl, token);
    } else if (rand < 0.70) {
        standingsLocalPage(baseUrl, token);
    } else {
        standingsUnitPage(baseUrl, token);
    }
}

/**
 * Simulate clicking a random challenge from an ActiveChallengesPage response.
 * External challenges link straight out of the app (no request); all other
 * types open the in-app detail page, which fires ChallengePage.
 * @param {string} baseUrl - Base URL of the GraphQL API
 * @param {string} token - JWT token for authorization
 * @param {object} activeChallengesResponse - HTTP response from activeChallengesPage
 */
export function clickRandomChallenge(baseUrl, token, activeChallengesResponse) {
    const data = parseResponse(activeChallengesResponse);
    const challenges = data && data.myCurrentProject && data.myCurrentProject.activeChallenges;
    if (!challenges || challenges.length === 0) {
        return;
    }
    const challenge = challenges[Math.floor(Math.random() * challenges.length)];
    if (challenge.__typename === 'ExternalChallenge') {
        // User follows the external link and leaves the app
        return;
    }
    challengePage(baseUrl, token, challenge.id);
}

/**
 * Simulate a realistic user session (SPA semantics: bootstrap queries fire
 * once on the cold load, in-app navigation only fires page queries):
 * 1. Cold load on the home page (GetMe + CurrentProject + ProfilePage + GetFirebaseToken)
 * 2. Visit challenges and standings in random order
 *    - challenges: active list, ~40% open the completed tab, ~50% click a challenge
 *    - standings: wrapper query + a random tab (global/local/unit)
 */
export function userJourney(baseUrl, token) {
    // 1. Cold app load, landing on the home page
    coldLoad(baseUrl, token, () => profilePage(baseUrl, token));
    sleep(Math.random() * 2 + 1);

    // 2. Build the two page visits: challenges and standings
    const pages = [
        () => {
            const response = activeChallengesPage(baseUrl, token);
            if (Math.random() < 0.4) {
                completedChallengesPage(baseUrl, token);
            }
            if (Math.random() < 0.5) {
                clickRandomChallenge(baseUrl, token, response);
            }
        },
        () => {
            standingsPage(baseUrl, token);
            randomStandingsPage(baseUrl, token);
        },
    ];

    // Randomize visit order
    if (Math.random() < 0.5) {
        pages.reverse();
    }

    pages[0]();
    sleep(Math.random() * 2 + 1);
    pages[1]();
    sleep(Math.random() * 2 + 1);
}
