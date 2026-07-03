import { ref } from 'vue'
import { vi } from 'vitest'

// The global auth middleware (app/middleware/auth.global.ts) runs during Nuxt
// app initialization for every mounted component. It calls useAuth0(), which is
// only provided by the Auth0 plugin at runtime. Stub the module here so the
// middleware initializes cleanly instead of throwing during mount.
// The Sentry Nuxt module registers a client plugin (sentry.client.config.ts)
// that calls browser-only APIs (e.g. replayIntegration) unavailable in the test
// build. No-op the whole module so Sentry init doesn't throw during mount.
vi.mock('@sentry/nuxt', () => ({
  init: vi.fn(),
  replayIntegration: vi.fn(),
  browserTracingIntegration: vi.fn(),
  captureException: vi.fn(),
  captureMessage: vi.fn(),
  setUser: vi.fn(),
  setContext: vi.fn(),
  setTag: vi.fn(),
  addBreadcrumb: vi.fn(),
}))

vi.mock('@auth0/auth0-vue', () => ({
  // The auth0 plugin (app/plugins/1.auth0.ts) installs this as a Vue plugin.
  createAuth0: () => ({ install: () => {} }),
  useAuth0: () => ({
    isLoading: ref(false),
    isAuthenticated: ref(false),
    user: ref(null),
    loginWithRedirect: vi.fn(),
    logout: vi.fn(),
    getAccessTokenSilently: vi.fn(),
  }),
}))
