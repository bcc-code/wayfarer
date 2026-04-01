import { sleep } from 'k6';

import { profilePage } from '../queries/profile.js';
import { activeChallengesPage, completedChallengesPage } from '../queries/challenges.js';
import { standingsGlobalPage } from '../queries/standings-global.js';
import { standingsLocalPage } from '../queries/standings-local.js';
import { standingsUnitPage } from '../queries/standings-unit.js';

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
 * Simulate a realistic user journey:
 * 1. Land on the home page (profile)
 * 2. Visit challenges and standings in random order
 */
export function userJourney(baseUrl, token) {
    // 1. Home page (always first)
    profilePage(baseUrl, token);
    sleep(Math.random() * 2 + 1);

    // 2. Build the two page visits: challenges and standings
    const pages = [
        () => {
            activeChallengesPage(baseUrl, token);
            completedChallengesPage(baseUrl, token);
        },
        () => randomStandingsPage(baseUrl, token),
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
