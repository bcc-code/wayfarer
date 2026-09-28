import { describe, it, expect, vi, beforeEach, afterEach } from 'vitest'
import { ref, type Ref } from 'vue'
import { createMockToken } from '../utils/auth-mocks'

/**
 * Global auth middleware tests.
 *
 * Runs the REAL `app/middleware/auth.global.ts`. The previous tests for this
 * middleware lived in `auth.test.ts` and re-implemented its logic inline
 * (`// Simulate middleware logic`), so they could not fail when the middleware
 * changed.
 *
 * `useAuth0` is module-mocked; the remaining dependencies are Nuxt
 * auto-imports and are stubbed as globals before the dynamic import.
 */

const auth0 = {
  isLoading: ref(false),
  isAuthenticated: ref(false),
  logout: vi.fn(),
}

vi.mock('@auth0/auth0-vue', () => ({ useAuth0: () => auth0 }))

const navigateTo = vi.fn()

/** Backing store for the stubbed `useLocalStorage`. */
let storage: Record<string, Ref<unknown>> = {}

function stubGlobals(tokenIsExpired = false) {
  vi.stubGlobal('defineNuxtRouteMiddleware', (fn: unknown) => fn)
  vi.stubGlobal('navigateTo', navigateTo)
  vi.stubGlobal('isTokenExpired', () => tokenIsExpired)
  vi.stubGlobal('useLocalStorage', (key: string, init: unknown) => {
    storage[key] ??= ref(typeof init === 'function' ? init() : init)
    return storage[key]
  })
}

async function loadMiddleware(tokenIsExpired = false) {
  stubGlobals(tokenIsExpired)
  vi.resetModules()
  const mod = await import('../../app/middleware/01.auth.global')
  return mod.default as (to: {
    path: string
    fullPath: string
  }) => Promise<unknown>
}

function route(path: string) {
  return { path, fullPath: path }
}

beforeEach(() => {
  storage = {}
  navigateTo.mockClear()
  auth0.logout.mockClear()
  auth0.isLoading.value = false
  auth0.isAuthenticated.value = false
  vi.stubGlobal('window', { location: { origin: 'https://app.test' } })
})

afterEach(() => {
  vi.unstubAllGlobals()
})

describe('middleware/auth.global', () => {
  it.each([['/login'], ['/auth0-callback'], ['/logout-callback']])(
    'skips the auth check on %s',
    async (path) => {
      const middleware = await loadMiddleware()

      await middleware(route(path))

      expect(navigateTo).not.toHaveBeenCalled()
    },
  )

  it('lets a request through when a valid token is present', async () => {
    const middleware = await loadMiddleware()
    storage.token = ref(createMockToken())

    await middleware(route('/challenges'))

    expect(navigateTo).not.toHaveBeenCalled()
  })

  it('redirects to /login with a redirect param when unauthenticated', async () => {
    const middleware = await loadMiddleware()
    auth0.isAuthenticated.value = false

    await middleware(route('/standings'))

    expect(navigateTo).toHaveBeenCalledWith(
      { path: '/login', query: { redirect: '/standings' } },
      { replace: true },
    )
  })

  it('omits the redirect param when landing on the root', async () => {
    const middleware = await loadMiddleware()

    await middleware(route('/'))

    expect(navigateTo).toHaveBeenCalledWith(
      { path: '/login', query: { redirect: undefined } },
      { replace: true },
    )
  })

  it('sends an Auth0-authenticated user without a token to the callback', async () => {
    const middleware = await loadMiddleware()
    auth0.isAuthenticated.value = true

    await middleware(route('/challenges'))

    expect(navigateTo).toHaveBeenCalledWith('/auth0-callback', {
      replace: true,
    })
  })

  it('clears an expired token and re-exchanges it', async () => {
    const middleware = await loadMiddleware(true)
    storage.token = ref(createMockToken())
    auth0.isAuthenticated.value = true

    await middleware(route('/challenges'))

    expect(storage.token!.value).toBeNull()
    expect(navigateTo).toHaveBeenCalledWith('/auth0-callback', {
      replace: true,
    })
  })

  describe('redirect-loop guard', () => {
    it('logs out after more than three rapid callback redirects', async () => {
      const middleware = await loadMiddleware()
      auth0.isAuthenticated.value = true

      // Four attempts inside the 5s window: the fourth trips the guard.
      for (let i = 0; i < 4; i++) {
        await middleware(route('/challenges'))
      }

      expect(auth0.logout).toHaveBeenCalledWith({
        logoutParams: { returnTo: 'https://app.test/login' },
      })
      // The guard resets its counters so the next attempt starts clean.
      expect(storage.auth_redirect_attempts!.value).toBe(0)
      expect(storage.auth_redirect_time!.value).toBe(0)
    })

    it('does not trip when redirects are spaced beyond the window', async () => {
      const middleware = await loadMiddleware()
      auth0.isAuthenticated.value = true

      await middleware(route('/challenges'))
      // Push the previous attempt outside the 5s window.
      storage.auth_redirect_time!.value = Date.now() - 10_000
      await middleware(route('/challenges'))

      expect(auth0.logout).not.toHaveBeenCalled()
      expect(storage.auth_redirect_attempts!.value).toBe(1)
    })

    it('clears the counters once a valid token is present', async () => {
      const middleware = await loadMiddleware()
      storage.auth_redirect_attempts = ref(3)
      storage.auth_redirect_time = ref(Date.now())
      storage.token = ref(createMockToken())

      await middleware(route('/challenges'))

      expect(storage.auth_redirect_attempts.value).toBe(0)
      expect(storage.auth_redirect_time.value).toBe(0)
    })
  })

  it('falls through to the token checks when Auth0 init stalls', async () => {
    const middleware = await loadMiddleware()
    // `until(...).toBe(false, { throwOnTimeout: true })` rejects; the
    // middleware swallows it and carries on to the token checks.
    auth0.isLoading.value = true

    const navigation = middleware(route('/challenges'))
    // Resolve the stall so the 10s timeout does not have to elapse.
    auth0.isLoading.value = false
    await navigation

    expect(navigateTo).toHaveBeenCalledWith(
      { path: '/login', query: { redirect: '/challenges' } },
      { replace: true },
    )
  })
})
