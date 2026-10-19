import { getUrqlClient } from '~/plugins/0.urql'
import { ConsentGateDocument, type ConsentGateQuery } from '~/api/generated'

// Auth pages have no user to gate, and the consent screen is where the gate
// sends people — bouncing it would loop.
const EXEMPT_PATHS = [
  '/login',
  '/auth0-callback',
  '/logout-callback',
  '/settings/consent',
]

/**
 * The one place that decides whether the user is sent to the consent screen.
 *
 * Two tiers, deliberately different in reach:
 *  - a required consent blocks the project, so it redirects from any route;
 *  - an unanswered optional one is the onboarding nudge, so it only pulls the
 *    user in from the front page, as it did when it lived in `index.vue`.
 */
export default defineNuxtRouteMiddleware(async (to) => {
  if (EXEMPT_PATHS.includes(to.path)) return
  if (to.path.startsWith('/admin')) return

  const client = getUrqlClient()
  if (!client) return

  const result = await client
    .query<ConsentGateQuery>(
      ConsentGateDocument,
      {},
      { requestPolicy: 'cache-first' },
    )
    .toPromise()

  // Fail open: offline or a failed query must not lock a user out of the app.
  if (result.error) return

  const status = result.data?.me.consentStatus

  if (blockingConsentIds(status).size > 0) {
    return navigateTo({ name: 'settings-consent' })
  }

  if (to.path === '/' && unansweredLocalConsents(status).length > 0) {
    // Re-enter onboarding presentation: a consent published after the user
    // finished onboarding has to be presented the same way the first ones were.
    useLocalStorage('hasCompletedOnboarding', false).value = false
    return navigateTo({ name: 'settings-consent' })
  }
})
