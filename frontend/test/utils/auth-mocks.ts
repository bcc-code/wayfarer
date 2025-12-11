import { ref, computed, type Ref } from 'vue'
import { vi } from 'vitest'
import type { GetMeQuery } from '../../app/api/generated'
import { RoleType } from '../../app/api/generated'

/**
 * Mock implementation of the useAuth composable
 *
 * @example
 * const auth = mockUseAuth({
 *   token: ref('mock-token'),
 *   me: ref({ id: 'US01', roles: [{ role: RoleType.Admin }] })
 * })
 */
export function mockUseAuth(
  overrides: Partial<ReturnType<typeof createAuthState>> = {},
) {
  const state = createAuthState()

  // If me ref is overridden, recreate computed properties to use the new ref
  if (overrides.me) {
    const meRef = overrides.me
    return {
      ...state,
      ...overrides,
      isSuperAdmin: computed(
        () =>
          meRef.value?.roles.some(
            (role) => role.role === RoleType.Superadmin,
          ) ?? false,
      ),
      isAdmin: computed(
        () =>
          meRef.value?.roles.some((role) => role.role === RoleType.Admin) ??
          false,
      ),
      isChurchAdmin: computed(
        () =>
          meRef.value?.roles.some(
            (role) => role.role === RoleType.ChurchAdmin,
          ) ?? false,
      ),
      isProjectAdmin: computed(
        () =>
          meRef.value?.roles.some(
            (role) => role.role === RoleType.ProjectAdmin,
          ) ?? false,
      ),
      isTeamLead: computed(
        () =>
          meRef.value?.roles.some((role) => role.role === RoleType.TeamLead) ??
          false,
      ),
    }
  }

  return {
    ...state,
    ...overrides,
  }
}

/**
 * Creates a default auth state object with all properties
 */
function createAuthState() {
  const token = ref<string | null>(null)
  const me = ref<GetMeQuery['me'] | null>(null)
  const isLoading = ref(false)

  const isSuperAdmin = computed(
    () =>
      me.value?.roles.some((role) => role.role === RoleType.Superadmin) ??
      false,
  )

  const isAdmin = computed(
    () => me.value?.roles.some((role) => role.role === RoleType.Admin) ?? false,
  )

  const isChurchAdmin = computed(
    () =>
      me.value?.roles.some((role) => role.role === RoleType.ChurchAdmin) ??
      false,
  )

  const isProjectAdmin = computed(
    () =>
      me.value?.roles.some((role) => role.role === RoleType.ProjectAdmin) ??
      false,
  )

  const isTeamLead = computed(
    () =>
      me.value?.roles.some((role) => role.role === RoleType.TeamLead) ?? false,
  )

  return {
    token,
    me,
    isLoading,
    isSuperAdmin,
    isAdmin,
    isChurchAdmin,
    isProjectAdmin,
    isTeamLead,
    setAccessToken: vi.fn((value: string) => {
      token.value = value
      isLoading.value = false
    }),
    getAccessTokenSilently: vi.fn(async () => {
      // Wait for loading to complete
      let attempts = 0
      while (isLoading.value && attempts < 100) {
        await new Promise((resolve) => setTimeout(resolve, 10))
        attempts++
      }
      return token.value
    }),
    loginWithRedirect: vi.fn(),
  }
}

/**
 * Mock implementation of useCookie
 * Returns a ref that can be read/written like a real cookie
 *
 * @example
 * const cookie = mockUseCookie('initial-value')
 * cookie.value = 'new-value'
 * expect(cookie.value).toBe('new-value')
 */
export function mockUseCookie<T = string>(
  initialValue: T | null = null,
): Ref<T | null> {
  return ref(initialValue)
}

/**
 * Creates a mock JWT token with the specified payload
 * The token format is valid (3 parts, base64url encoded) but not cryptographically signed
 *
 * @example
 * const token = createMockToken({ user_id: 'US01', exp: Date.now() / 1000 + 3600 })
 */
export function createMockToken(payload: Record<string, unknown> = {}): string {
  const header = {
    alg: 'HS256',
    typ: 'JWT',
  }

  const defaultPayload = {
    user_id: 'US01ARZ3NDEKTSV4RRFFQ69G5FAV',
    user_roles: ['user'],
    iss: 'wayfarer',
    iat: Math.floor(Date.now() / 1000),
    exp: Math.floor(Date.now() / 1000) + 3600, // 1 hour from now
    ...payload,
  }

  // Base64url encode (not base64 - uses different characters)
  const base64urlEncode = (obj: unknown): string => {
    const json = JSON.stringify(obj)
    const base64 = btoa(json)
    return base64.replace(/\+/g, '-').replace(/\//g, '_').replace(/=/g, '')
  }

  const headerEncoded = base64urlEncode(header)
  const payloadEncoded = base64urlEncode(defaultPayload)
  const signature = 'mock-signature'

  return `${headerEncoded}.${payloadEncoded}.${signature}`
}

/**
 * Creates a mock user object for the me query
 *
 * @example
 * const user = createMockUser({
 *   id: 'US01',
 *   roles: [{ role: RoleType.Admin, scope: null }]
 * })
 */
export function createMockUser(
  overrides: Partial<GetMeQuery['me']> = {},
): GetMeQuery['me'] {
  return {
    id: 'US01ARZ3NDEKTSV4RRFFQ69G5FAV',
    name: 'Test User',
    email: 'test@example.com',
    avatarUrl: null,
    roles: [{ role: RoleType.User, scope: null }],
    church: {
      id: 'CH01ARZ3NDEKTSV4RRFFQ69G5FAV',
      name: 'Test Church',
    },
    ...overrides,
  } as GetMeQuery['me']
}

/**
 * Creates a mock fetch response for testing
 *
 * @example
 * const mockFetch = vi.fn().mockResolvedValue(
 *   createMockFetchResponse({ token: 'new-token' })
 * )
 */
export function createMockFetchResponse<T = unknown>(
  data: T,
  ok = true,
  status = 200,
) {
  return {
    ok,
    status,
    json: async () => data,
    text: async () => JSON.stringify(data),
  }
}

/**
 * Mock implementation of navigateTo
 * Records navigation calls for testing
 *
 * @example
 * const navigate = mockNavigateTo()
 * await navigate('/admin')
 * expect(navigate.mock.calls[0][0]).toBe('/admin')
 */
export function mockNavigateTo() {
  return vi.fn((to: string | { path: string }, options?: unknown) => {
    // Return a promise like the real navigateTo
    return Promise.resolve()
  })
}

/**
 * Mock implementation of createError
 *
 * @example
 * const error = mockCreateError({ statusCode: 403 })
 */
export function mockCreateError() {
  return vi.fn((error: { statusCode: number; message?: string }) => {
    return new Error(
      `Error ${error.statusCode}: ${error.message || 'Forbidden'}`,
    )
  })
}

/**
 * Mock implementation of window.location
 * Useful for testing redirect flows
 *
 * @example
 * const location = mockWindowLocation()
 * location.pathname = '/admin/projects'
 */
export function mockWindowLocation(pathname = '/') {
  return {
    pathname,
    href: `http://localhost:3000${pathname}`,
    search: '',
    hash: '',
  }
}

/**
 * Mock implementation of useRoute
 * Returns a route object with query parameters
 *
 * @example
 * const route = mockUseRoute({ token: 'abc123', redirect: '/admin' })
 */
export function mockUseRoute(query: Record<string, string | string[]> = {}) {
  return {
    query,
    path: '/callback',
    name: 'callback',
    params: {},
    meta: {},
  }
}

/**
 * Mock implementation of useRuntimeConfig
 *
 * @example
 * const config = mockUseRuntimeConfig()
 */
export function mockUseRuntimeConfig() {
  return {
    public: {
      apiUrl: 'http://localhost:8080/graphql',
      tokenUrl: 'http://localhost:8080/token',
      loginUrl: 'https://login.example.com/auth',
    },
  }
}

/**
 * Mock implementation of $fetch
 * Use with vi.fn() to control responses
 *
 * @example
 * const mockFetch = vi.fn().mockResolvedValue({ token: 'new-token' })
 * global.$fetch = mockFetch
 */
export function mockFetch() {
  return vi.fn()
}
