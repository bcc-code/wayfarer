import { graphqlRequest, checkGraphQLResponse } from '../lib/graphql.js';

const PROFILE_PAGE_QUERY = `
query ProfilePage($ageFilter: LeaderboardFilter) {
  me {
    id
    name
    consentStatus {
      pendingConsents {
        __typename
        id
        key
        version
        title
        body {
          html
        }
        shortText
        url
        managementType
        managedBy
      }
    }
  }
  myCurrentProject {
    id
    name
    infoMessage {
      markdown
      html
    }
    infoMessageStart
    infoMessageEnd
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
    achievements {
      __typename
      id
      name
      descriptionPending
      descriptionCompleted
      imagePendingObject {
        url
        width
        height
        blurhash
      }
      imageCompletedObject {
        url
        width
        height
        blurhash
      }
      hidden
      achievedAt
      celebratedAt
      points
    }
    myPoints
    leaderboard(entityType: PERSONS, filter: $ageFilter) {
      me {
        rank
      }
    }
    myTeam {
      superTeam {
        id
        name
        color
        imageObject {
          url
          blurhash
        }
      }
    }
  }
}
`;

/**
 * Execute the ProfilePage query (home page)
 */
export function profilePage(baseUrl, token) {
    const response = graphqlRequest(baseUrl, PROFILE_PAGE_QUERY, {}, token, 'ProfilePage');
    checkGraphQLResponse(response, 'ProfilePage');
    return response;
}
