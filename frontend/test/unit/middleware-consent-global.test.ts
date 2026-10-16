import { describe, it, expect, vi, beforeEach } from 'vitest'
import { ConsentManagementType } from '../../app/api/generated'
import {
  blockingConsentIds,
  unansweredLocalConsents,
} from '../../app/utils/consentGate'

/**
 * The consent gate, running the REAL `middleware/03.consent.global.ts` against
 * a stubbed urql client.
 *
 * Two tiers: a required consent redirects from anywhere, an unanswered optional
 * one only from the front page. The required rule is not "pending" — a REJECTED
 * required consent blocks too.
 */

const navigateTo = vi.fn((to: unknown) => ({ __redirect: to }))
const query = vi.fn()

vi.mock('../../app/plugins/0.urql', () => ({
  getUrqlClient: () => ({ query }),
}))

type ConsentStub = {
  id: string
  isRequired: boolean
  managementType: ConsentManagementType
}

function consent(
  isRequired: boolean,
  managementType = ConsentManagementType.Local,
): ConsentStub {
  return { id: 'CN01', isRequired, managementType }
}

/** Shapes the ConsentGate result the middleware reads. */
function respond(opts: {
  pending?: ConsentStub[]
  rejected?: ConsentStub[]
  error?: boolean
}) {
  query.mockReturnValue({
    toPromise: async () =>
      opts.error
        ? { error: new Error('offline'), data: undefined }
        : {
            error: undefined,
            data: {
              me: {
                consentStatus: {
                  pendingConsents: opts.pending ?? [],
                  rejectedConsents: (opts.rejected ?? []).map((c) => ({
                    id: 'UC01',
                    consent: c,
                  })),
                },
              },
            },
          },
  })
}

const onboardingFlag = { value: true }

async function loadMiddleware() {
  vi.stubGlobal('defineNuxtRouteMiddleware', (fn: unknown) => fn)
  vi.stubGlobal('navigateTo', navigateTo)
  vi.stubGlobal('useLocalStorage', () => onboardingFlag)
  vi.stubGlobal('blockingConsentIds', blockingConsentIds)
  vi.stubGlobal('unansweredLocalConsents', unansweredLocalConsents)
  vi.resetModules()
  const mod = await import('../../app/middleware/03.consent.global')
  return mod.default as (to: { path: string }) => Promise<unknown>
}

beforeEach(() => {
  vi.clearAllMocks()
  onboardingFlag.value = true
  respond({})
})

describe('consent gate', () => {
  it('redirects to the consent screen when a required consent is pending', async () => {
    respond({ pending: [consent(true)] })
    const middleware = await loadMiddleware()

    expect(await middleware({ path: '/challenges' })).toEqual({
      __redirect: { name: 'settings-consent' },
    })
  })

  it('redirects when a required consent was REJECTED, not just pending', async () => {
    respond({ rejected: [consent(true)] })
    const middleware = await loadMiddleware()

    expect(await middleware({ path: '/challenges' })).toEqual({
      __redirect: { name: 'settings-consent' },
    })
  })

  it('lets an optional consent through, pending or rejected', async () => {
    respond({ pending: [consent(false)], rejected: [consent(false)] })
    const middleware = await loadMiddleware()

    expect(await middleware({ path: '/challenges' })).toBeUndefined()
  })

  it('ignores remote consents, which cannot be accepted in the app', async () => {
    respond({ pending: [consent(true, ConsentManagementType.Remote)] })
    const middleware = await loadMiddleware()

    expect(await middleware({ path: '/challenges' })).toBeUndefined()
  })

  it.each([
    '/login',
    '/auth0-callback',
    '/logout-callback',
    '/settings/consent',
    '/admin/consents',
  ])('never gates %s', async (path) => {
    respond({ pending: [consent(true)] })
    const middleware = await loadMiddleware()

    expect(await middleware({ path })).toBeUndefined()
    expect(query).not.toHaveBeenCalled()
  })

  it('pulls the user in from the front page for an unanswered optional consent', async () => {
    respond({ pending: [consent(false)] })
    const middleware = await loadMiddleware()

    expect(await middleware({ path: '/' })).toEqual({
      __redirect: { name: 'settings-consent' },
    })
    // Presented as onboarding, the way the first consents were.
    expect(onboardingFlag.value).toBe(false)
  })

  it('does not pull the user in from other routes for an optional consent', async () => {
    respond({ pending: [consent(false)] })
    const middleware = await loadMiddleware()

    expect(await middleware({ path: '/challenges' })).toBeUndefined()
    expect(onboardingFlag.value).toBe(true)
  })

  it('fails open when the query errors, rather than locking the user out', async () => {
    respond({ error: true })
    const middleware = await loadMiddleware()

    expect(await middleware({ path: '/challenges' })).toBeUndefined()
  })
})
