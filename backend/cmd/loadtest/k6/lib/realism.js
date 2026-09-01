import { authDance } from '../queries/bootstrap.js';

// Realistic-auth simulation, off by default. REALISM=1 is the master switch
// enabling production-like defaults (3% of journeys do the Auth0 dance, 50%
// refresh their Firebase token mid-session); the individual fractions can
// always be set explicitly to override either default.
export const REALISM = ['1', 'true', 'yes'].includes(String(__ENV.REALISM).toLowerCase());

function fraction(name, realismDefault) {
    if (__ENV[name] !== undefined && __ENV[name] !== '') {
        return parseFloat(__ENV[name]) || 0;
    }
    return REALISM ? realismDefault : 0;
}

export const AUTH_DANCE_FRACTION = fraction('AUTH_DANCE_FRACTION', 0.03);
export const FIREBASE_REFRESH_FRACTION = fraction('FIREBASE_REFRESH_FRACTION', 0.5);

/**
 * With probability AUTH_DANCE_FRACTION, start the session expired: perform
 * the Auth0 dance and continue with the freshly minted Wayfarer JWT.
 * @param {string} baseUrl - Base URL of the API
 * @param {string} token - Pre-minted Wayfarer JWT (fallback)
 * @param {string|null} auth0Token - Simulated Auth0 JWT, if available
 * @returns {string} The token to use for the rest of the journey
 */
export function maybeAuthDance(baseUrl, token, auth0Token) {
    if (!auth0Token || Math.random() >= AUTH_DANCE_FRACTION) {
        return token;
    }
    return authDance(baseUrl, auth0Token) || token;
}
