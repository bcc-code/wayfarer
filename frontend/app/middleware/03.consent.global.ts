import { getUrqlClient } from '~/plugins/0.urql'
import {
  BlockingConsentsDocument,
  ConsentManagementType,
  type BlockingConsentsQuery,
} from '~/api/generated'

// Auth pages have no user to gate, and the consent screen is where the gate
// sends people — bouncing it would loop.
const EXEMPT_PATHS = [
  '/login',
  '/auth0-callback',
  '/logout-callback',
  '/settings/consent',
]

/**
 * A required consent the user has not accepted blocks the whole project, so the
 * gate runs on every route rather than only on the front page.
 *
 * Only LOCAL consents count. A remote one cannot be accepted in the app, so
 * gating on it would strand the user on a screen with no accept button —
 * the API is meant to reject that combination, but until it does, failing open
 * here is the safer half.
 */
function isBlocking(data: BlockingConsentsQuery | undefined): boolean {
  const status = data?.me.consentStatus
  if (!status) return false

  const local = (c: { managementType: ConsentManagementType }) =>
    c.managementType === ConsentManagementType.Local

  const pending = status.pendingConsents.filter((c) => c.isRequired && local(c))
  const rejected = status.rejectedConsents.filter(
    (entry) => entry.consent.isRequired && local(entry.consent),
  )

  return pending.length > 0 || rejected.length > 0
}

export default defineNuxtRouteMiddleware(async (to) => {
  if (EXEMPT_PATHS.includes(to.path)) return
  if (to.path.startsWith('/admin')) return

  const client = getUrqlClient()
  if (!client) return

  const result = await client
    .query<BlockingConsentsQuery>(
      BlockingConsentsDocument,
      {},
      { requestPolicy: 'cache-first' },
    )
    .toPromise()

  // Fail open: offline or a failed query must not lock a user out of the app.
  if (result.error) return

  if (isBlocking(result.data)) {
    return navigateTo({ name: 'settings-consent' })
  }
})
