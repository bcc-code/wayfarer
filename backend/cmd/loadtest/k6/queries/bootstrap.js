import { graphqlRequest, checkGraphQLResponse } from '../lib/graphql.js';

// Queries fired by the SPA on every cold app load, before any page query:
// GetMe (useAuth), CurrentProject (default layout) and GetFirebaseToken
// (Firestore realtime sync bootstrap).

const GET_ME_QUERY = `
query GetMe {
  me {
    id
    name
    email
    image
    membersId
    language
    church {
      id
      name
      country
      category
    }
    gender
    birthdate
    age
    createdAt
    roles {
      id
      role
      scope {
        id
        type
        church {
          id
        }
        team {
          id
        }
        project {
          id
        }
      }
    }
  }
}
`;

const CURRENT_PROJECT_QUERY = `
query CurrentProject {
  myCurrentProject {
    id
    branding {
      logoImage {
        url
        width
        height
        blurhash
      }
      bannerImage {
        url
        width
        height
        blurhash
      }
      rounding
      colors {
        light {
          accent
          accentContrast
          onAccent
          backgroundDefault
          backgroundRaised
          backgroundIndent
          textDefault
          textMuted
          textHint
          shadowDefault
          shadowBlank
          borderDefault
        }
        dark {
          accent
          accentContrast
          onAccent
          backgroundDefault
          backgroundRaised
          backgroundIndent
          textDefault
          textMuted
          textHint
          shadowDefault
          shadowBlank
          borderDefault
        }
      }
    }
    activeChallengesCount
    leaderboard(entityType: PERSONS, first: 1) {
      totalCount
    }
  }
}
`;

const GET_FIREBASE_TOKEN_QUERY = `
query GetFirebaseToken {
  firebaseToken {
    token
    expiresIn
  }
}
`;

/**
 * Execute the GetMe query (fired by useAuth on every cold load)
 */
export function getMe(baseUrl, token) {
    const response = graphqlRequest(baseUrl, GET_ME_QUERY, {}, token, 'GetMe');
    checkGraphQLResponse(response, 'GetMe');
    return response;
}

/**
 * Execute the CurrentProject query (fired by the default layout on every cold load)
 */
export function currentProject(baseUrl, token) {
    const response = graphqlRequest(baseUrl, CURRENT_PROJECT_QUERY, {}, token, 'CurrentProject');
    checkGraphQLResponse(response, 'CurrentProject');
    return response;
}

/**
 * Execute the GetFirebaseToken query (Firestore sync bootstrap).
 * Skipped when SKIP_FIREBASE_TOKEN is set, for environments without
 * Firebase credentials configured.
 */
export function getFirebaseToken(baseUrl, token) {
    if (__ENV.SKIP_FIREBASE_TOKEN) {
        return null;
    }
    const response = graphqlRequest(baseUrl, GET_FIREBASE_TOKEN_QUERY, {}, token, 'GetFirebaseToken');
    checkGraphQLResponse(response, 'GetFirebaseToken');
    return response;
}

/**
 * Simulate a cold app load: the bootstrap queries plus the landing page query.
 * The browser fires GetMe, CurrentProject and the page query as a parallel
 * burst, then GetFirebaseToken once the user is known; we issue them
 * back-to-back without sleeps.
 * @param {string} baseUrl - Base URL of the GraphQL API
 * @param {string} token - JWT token for authorization
 * @param {function} pageQueryFn - Executes the landing page query, returns its response
 * @returns {object} The landing page query response
 */
export function coldLoad(baseUrl, token, pageQueryFn) {
    getMe(baseUrl, token);
    currentProject(baseUrl, token);
    const pageResponse = pageQueryFn ? pageQueryFn() : null;
    getFirebaseToken(baseUrl, token);
    return pageResponse;
}
