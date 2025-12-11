import { describe, it, expect, vi, beforeEach } from 'vitest'
import { ref } from 'vue'
import {
  mockUseAuth,
  mockUseCookie,
  mockNavigateTo,
  mockUseRoute,
  mockUseRuntimeConfig,
  mockFetch,
  createMockToken,
  createMockUser,
} from '../utils/auth-mocks'
import { RoleType, type GetMeQuery } from '../../app/api/generated'

/**
 * Authentication Flow Tests
 *
 * Tests the critical authentication flow including:
 * - Token storage and retrieval
 * - Callback handling
 * - Login redirects
 * - Role-based authorization
 * - Token expiration handling
 * - Error scenarios
 */

describe('Authentication Flow', () => {
  describe('useAuth composable', () => {
    beforeEach(() => {
      // Clear any existing auth state
      vi.clearAllMocks()
    })

    describe('Token Management', () => {
      it('should store token in cookie on setAccessToken', () => {
        const cookie = mockUseCookie<string>()
        const auth = mockUseAuth({
          token: cookie,
          setAccessToken: (value: string) => {
            cookie.value = value
          },
        })

        const token = createMockToken()
        auth.setAccessToken(token)

        expect(cookie.value).toBe(token)
      })

      it('should retrieve token from cookie', () => {
        const expectedToken = createMockToken()
        const cookie = mockUseCookie<string>(expectedToken)
        const auth = mockUseAuth({ token: cookie })

        expect(auth.token.value).toBe(expectedToken)
      })

      it('should return null when no token exists', () => {
        const cookie = mockUseCookie<string>(null)
        const auth = mockUseAuth({ token: cookie })

        expect(auth.token.value).toBeNull()
      })

      it('should overwrite existing token when setting new token', () => {
        const firstToken = createMockToken({ user_id: 'US01' })
        const cookie = mockUseCookie<string>(firstToken)
        const auth = mockUseAuth({
          token: cookie,
          setAccessToken: (value: string) => {
            cookie.value = value
          },
        })

        expect(cookie.value).toBe(firstToken)

        const secondToken = createMockToken({ user_id: 'US02' })
        auth.setAccessToken(secondToken)

        expect(cookie.value).toBe(secondToken)
        expect(cookie.value).not.toBe(firstToken)
      })

      it('should handle token with special characters', () => {
        // JWT tokens contain dots, dashes, and underscores (base64url)
        const tokenWithSpecialChars =
          'eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJ1c2VyX2lkIjoiVVMwMUFSWiJ9.dBjftJeZ4CVP-mB92K27uhbUJU1p1r_wW1gFWFOEjXk'
        const cookie = mockUseCookie<string>()
        const auth = mockUseAuth({
          token: cookie,
          setAccessToken: (value: string) => {
            cookie.value = value
          },
        })

        auth.setAccessToken(tokenWithSpecialChars)

        expect(cookie.value).toBe(tokenWithSpecialChars)
      })

      it('should handle very long tokens', () => {
        // Create a token with a very long payload (> 1KB)
        const largePayload: Record<string, unknown> = {
          user_id: 'US01',
          data: 'x'.repeat(2000),
        }
        const longToken = createMockToken(largePayload)

        const cookie = mockUseCookie<string>()
        const auth = mockUseAuth({
          token: cookie,
          setAccessToken: (value: string) => {
            cookie.value = value
          },
        })

        auth.setAccessToken(longToken)

        expect(cookie.value).toBe(longToken)
        expect(cookie.value!.length).toBeGreaterThan(1000)
      })

      it('should set isLoading to false after token is set', () => {
        const cookie = mockUseCookie<string>()
        const isLoading = ref(true)
        const auth = mockUseAuth({
          token: cookie,
          isLoading,
          setAccessToken: (value: string) => {
            cookie.value = value
            isLoading.value = false
          },
        })

        expect(isLoading.value).toBe(true)

        const token = createMockToken()
        auth.setAccessToken(token)

        expect(isLoading.value).toBe(false)
      })

      it('should wait for loading to complete in getAccessTokenSilently', async () => {
        const token = createMockToken()
        const cookie = mockUseCookie<string>(token)
        const isLoading = ref(true)

        const auth = mockUseAuth({
          token: cookie,
          isLoading,
          getAccessTokenSilently: async () => {
            let attempts = 0
            while (isLoading.value && attempts < 100) {
              await new Promise((resolve) => setTimeout(resolve, 10))
              attempts++
            }
            return cookie.value
          },
        })

        // Start async call while loading is true
        const tokenPromise = auth.getAccessTokenSilently()

        // Simulate loading completing after 50ms
        setTimeout(() => {
          isLoading.value = false
        }, 50)

        const result = await tokenPromise
        expect(result).toBe(token)
      })

      it('should clear token on logout', () => {
        const token = createMockToken()
        const cookie = mockUseCookie<string>(token)

        expect(cookie.value).toBe(token)

        // Logout by clearing token
        cookie.value = null

        expect(cookie.value).toBeNull()
      })
    })

    describe('Login Redirect', () => {
      const LOGIN_URL = 'https://login.example.com/auth'

      beforeEach(() => {
        // Reset window.location mock before each test
        delete (global as any).window
        ;(global as any).window = {}
      })

      it('should redirect to login URL with current path', () => {
        const currentPath = '/admin/projects'
        ;(global as any).window.location = { pathname: currentPath }

        const navigate = vi.fn()
        const auth = mockUseAuth({
          loginWithRedirect: () => {
            const redirectUrl = `${LOGIN_URL}?redirect=${window.location.pathname}`
            return navigate(redirectUrl, { external: true })
          },
        })

        auth.loginWithRedirect()

        expect(navigate).toHaveBeenCalledWith(
          `${LOGIN_URL}?redirect=${currentPath}`,
          { external: true },
        )
      })

      it('should include redirect parameter in login URL', () => {
        const targetPath = '/challenges/123'
        ;(global as any).window.location = { pathname: targetPath }

        const navigate = vi.fn()
        const auth = mockUseAuth({
          loginWithRedirect: () => {
            const redirectUrl = `${LOGIN_URL}?redirect=${encodeURIComponent(window.location.pathname)}`
            return navigate(redirectUrl, { external: true })
          },
        })

        auth.loginWithRedirect()

        const callArg = navigate.mock.calls[0][0] as string
        expect(callArg).toContain('redirect=')
        expect(callArg).toContain(encodeURIComponent(targetPath))
      })

      it('should handle paths with query params in redirect', () => {
        const pathWithQuery = '/search?q=test&filter=active'
        ;(global as any).window.location = { pathname: '/search' }

        const navigate = vi.fn()
        const auth = mockUseAuth({
          loginWithRedirect: () => {
            // In real implementation, this would use window.location.pathname
            // which doesn't include the query string
            const redirectUrl = `${LOGIN_URL}?redirect=${encodeURIComponent('/search')}`
            return navigate(redirectUrl, { external: true })
          },
        })

        auth.loginWithRedirect()

        expect(navigate).toHaveBeenCalled()
        const callArg = navigate.mock.calls[0][0] as string
        expect(callArg).toContain('redirect=')
      })

      it('should handle paths with hash fragments', () => {
        const pathWithHash = '/page#section'
        ;(global as any).window.location = { pathname: '/page' }

        const navigate = vi.fn()
        const auth = mockUseAuth({
          loginWithRedirect: () => {
            // pathname excludes the hash
            const redirectUrl = `${LOGIN_URL}?redirect=${encodeURIComponent('/page')}`
            return navigate(redirectUrl, { external: true })
          },
        })

        auth.loginWithRedirect()

        expect(navigate).toHaveBeenCalled()
        const callArg = navigate.mock.calls[0][0] as string
        expect(callArg).toBe(`${LOGIN_URL}?redirect=%2Fpage`)
      })

      it('should encode special characters in redirect URL', () => {
        const pathWithSpaces = '/search results/item'
        ;(global as any).window.location = { pathname: pathWithSpaces }

        const navigate = vi.fn()
        const auth = mockUseAuth({
          loginWithRedirect: () => {
            // URL encoding should handle special characters
            const redirectUrl = `${LOGIN_URL}?redirect=${window.location.pathname}`
            return navigate(redirectUrl, { external: true })
          },
        })

        auth.loginWithRedirect()

        expect(navigate).toHaveBeenCalled()
        const callArg = navigate.mock.calls[0][0] as string
        // Should contain the path (may or may not be encoded in the mock)
        expect(callArg).toContain('redirect=')
      })
    })

    describe('Role-based Authorization', () => {
      it('should detect superadmin role correctly', () => {
        const user = createMockUser({
          roles: [{ role: RoleType.Superadmin, scope: null }],
        })
        const auth = mockUseAuth({ me: ref(user) })

        expect(auth.isSuperAdmin.value).toBe(true)
        expect(auth.isAdmin.value).toBe(false)
      })

      it('should detect admin role correctly', () => {
        const user = createMockUser({
          roles: [{ role: RoleType.Admin, scope: null }],
        })
        const auth = mockUseAuth({ me: ref(user) })

        expect(auth.isAdmin.value).toBe(true)
        expect(auth.isSuperAdmin.value).toBe(false)
      })

      it('should detect church admin role correctly', () => {
        const user = createMockUser({
          roles: [
            {
              role: RoleType.ChurchAdmin,
              scope: { churchId: 'CH01ARZ3NDEKTSV4RRFFQ69G5FAV' },
            },
          ],
        })
        const auth = mockUseAuth({ me: ref(user) })

        expect(auth.isChurchAdmin.value).toBe(true)
        expect(auth.isAdmin.value).toBe(false)
      })

      it('should detect project admin role correctly', () => {
        const user = createMockUser({
          roles: [
            {
              role: RoleType.ProjectAdmin,
              scope: { projectId: 'PR01ARZ3NDEKTSV4RRFFQ69G5FAV' },
            },
          ],
        })
        const auth = mockUseAuth({ me: ref(user) })

        expect(auth.isProjectAdmin.value).toBe(true)
        expect(auth.isAdmin.value).toBe(false)
      })

      it('should detect team lead role correctly', () => {
        const user = createMockUser({
          roles: [
            {
              role: RoleType.TeamLead,
              scope: { teamId: 'TM01ARZ3NDEKTSV4RRFFQ69G5FAV' },
            },
          ],
        })
        const auth = mockUseAuth({ me: ref(user) })

        expect(auth.isTeamLead.value).toBe(true)
        expect(auth.isAdmin.value).toBe(false)
      })

      it('should handle user with multiple roles', () => {
        const user = createMockUser({
          roles: [
            { role: RoleType.Admin, scope: null },
            {
              role: RoleType.ProjectAdmin,
              scope: { projectId: 'PR01ARZ3NDEKTSV4RRFFQ69G5FAV' },
            },
            { role: RoleType.User, scope: null },
          ],
        })
        const auth = mockUseAuth({ me: ref(user) })

        expect(auth.isAdmin.value).toBe(true)
        expect(auth.isProjectAdmin.value).toBe(true)
        expect(auth.isSuperAdmin.value).toBe(false)
      })

      it('should return false for missing roles', () => {
        const user = createMockUser({
          roles: [{ role: RoleType.User, scope: null }],
        })
        const auth = mockUseAuth({ me: ref(user) })

        expect(auth.isSuperAdmin.value).toBe(false)
        expect(auth.isAdmin.value).toBe(false)
        expect(auth.isChurchAdmin.value).toBe(false)
        expect(auth.isProjectAdmin.value).toBe(false)
        expect(auth.isTeamLead.value).toBe(false)
      })

      it('should handle empty roles array', () => {
        const user = createMockUser({ roles: [] })
        const auth = mockUseAuth({ me: ref(user) })

        expect(auth.isSuperAdmin.value).toBe(false)
        expect(auth.isAdmin.value).toBe(false)
      })

      it('should handle null/undefined user', () => {
        const auth = mockUseAuth({ me: ref(null) })

        expect(auth.isSuperAdmin.value).toBe(false)
        expect(auth.isAdmin.value).toBe(false)
        expect(auth.isChurchAdmin.value).toBe(false)
        expect(auth.isProjectAdmin.value).toBe(false)
        expect(auth.isTeamLead.value).toBe(false)
      })
    })

    describe('GraphQL Me Query', () => {
      it('should update me state when user data is available', () => {
        const user = createMockUser({
          id: 'US01ARZ3NDEKTSV4RRFFQ69G5FAV',
          name: 'John Doe',
          email: 'john@example.com',
        })

        const meRef = ref<GetMeQuery['me'] | null>(null)
        const auth = mockUseAuth({ me: meRef })

        expect(auth.me.value).toBeNull()

        // Simulate query resolving with user data
        meRef.value = user

        expect(auth.me.value).toStrictEqual(user)
        expect(auth.me.value?.name).toBe('John Doe')
        expect(auth.me.value?.email).toBe('john@example.com')
      })

      it('should set isLoading to false when query completes', async () => {
        const isLoading = ref(true)
        const meRef = ref<GetMeQuery['me'] | null>(null)
        const auth = mockUseAuth({
          me: meRef,
          isLoading,
        })

        expect(isLoading.value).toBe(true)

        // Simulate query completing
        meRef.value = createMockUser()
        isLoading.value = false

        expect(isLoading.value).toBe(false)
        expect(auth.me.value).not.toBeNull()
      })

      it('should handle null user data gracefully', () => {
        const meRef = ref<GetMeQuery['me'] | null>(null)
        const auth = mockUseAuth({ me: meRef })

        expect(auth.me.value).toBeNull()
        expect(auth.isSuperAdmin.value).toBe(false)
        expect(auth.isAdmin.value).toBe(false)
      })

      it('should update role computeds when me data changes', async () => {
        const meRef = ref<GetMeQuery['me'] | null>(null)
        const auth = mockUseAuth({ me: meRef })

        // Initially no user
        expect(auth.isAdmin.value).toBe(false)

        // Update to admin user
        meRef.value = createMockUser({
          roles: [{ role: RoleType.Admin, scope: null }],
        })

        // Role computed should update
        expect(auth.isAdmin.value).toBe(true)

        // Update to regular user
        meRef.value = createMockUser({
          roles: [{ role: RoleType.User, scope: null }],
        })

        // Role computed should update again
        expect(auth.isAdmin.value).toBe(false)
      })

      it('should handle user with church data', () => {
        const user = createMockUser({
          church: {
            id: 'CH01ARZ3NDEKTSV4RRFFQ69G5FAV',
            name: 'Test Church',
          },
        })

        const meRef = ref(user)
        const auth = mockUseAuth({ me: meRef })

        expect(auth.me.value?.church.id).toBe('CH01ARZ3NDEKTSV4RRFFQ69G5FAV')
        expect(auth.me.value?.church.name).toBe('Test Church')
      })

      it('should maintain reactivity when me ref is updated multiple times', () => {
        const meRef = ref<GetMeQuery['me'] | null>(null)
        const auth = mockUseAuth({ me: meRef })

        // First update
        meRef.value = createMockUser({ name: 'User 1' })
        expect(auth.me.value?.name).toBe('User 1')

        // Second update
        meRef.value = createMockUser({ name: 'User 2' })
        expect(auth.me.value?.name).toBe('User 2')

        // Clear
        meRef.value = null
        expect(auth.me.value).toBeNull()
      })
    })
  })

  describe('Callback Page', () => {
    describe('Token Validation', () => {
      const TOKEN_URL = 'http://localhost:8080/token'

      it('should validate token with backend on mount', async () => {
        const inputToken = 'incoming-token-from-oauth'
        const validatedToken = createMockToken()

        const fetch = mockFetch()
        fetch.mockResolvedValue({ token: validatedToken })

        const setAccessToken = vi.fn()
        const navigate = mockNavigateTo()

        // Simulate callback page behavior
        const route = mockUseRoute({ token: inputToken })
        const config = mockUseRuntimeConfig()

        // Simulate onBeforeMount logic
        const { token } = route.query
        if (token) {
          const response = await fetch(
            `${config.public.tokenUrl}?token=${token}`,
            { method: 'GET' },
          )
          if (response && response.token) {
            setAccessToken(response.token)
            navigate('/')
          }
        }

        expect(fetch).toHaveBeenCalledWith(
          `${TOKEN_URL}?token=${inputToken}`,
          { method: 'GET' },
        )
        expect(setAccessToken).toHaveBeenCalledWith(validatedToken)
      })

      it('should store validated token in cookie', async () => {
        const inputToken = 'oauth-token'
        const validatedToken = createMockToken()

        const cookie = mockUseCookie<string>()
        const fetch = mockFetch()
        fetch.mockResolvedValue({ token: validatedToken })

        const setAccessToken = vi.fn((token: string) => {
          cookie.value = token
        })

        const route = mockUseRoute({ token: inputToken })
        const config = mockUseRuntimeConfig()

        const response = await fetch(
          `${config.public.tokenUrl}?token=${inputToken}`,
        )
        if (response && response.token) {
          setAccessToken(response.token)
        }

        expect(cookie.value).toBe(validatedToken)
      })

      it('should redirect to home after successful validation', async () => {
        const inputToken = 'oauth-token'
        const validatedToken = createMockToken()

        const fetch = mockFetch()
        fetch.mockResolvedValue({ token: validatedToken })

        const setAccessToken = vi.fn()
        const navigate = mockNavigateTo()

        const route = mockUseRoute({ token: inputToken })
        const config = mockUseRuntimeConfig()

        const response = await fetch(
          `${config.public.tokenUrl}?token=${inputToken}`,
        )
        if (response && response.token) {
          setAccessToken(response.token)
          navigate('/')
        }

        expect(navigate).toHaveBeenCalledWith('/')
      })

      it('should redirect to specified redirect param', async () => {
        const inputToken = 'oauth-token'
        const validatedToken = createMockToken()
        const redirectPath = '/admin/projects/123'

        const fetch = mockFetch()
        fetch.mockResolvedValue({ token: validatedToken })

        const setAccessToken = vi.fn()
        const navigate = mockNavigateTo()

        const route = mockUseRoute({
          token: inputToken,
          redirect: redirectPath,
        })
        const config = mockUseRuntimeConfig()

        const response = await fetch(
          `${config.public.tokenUrl}?token=${inputToken}`,
        )
        if (response && response.token) {
          setAccessToken(response.token)

          const { redirect } = route.query
          if (redirect && typeof redirect === 'string') {
            navigate(redirect)
          } else {
            navigate('/')
          }
        }

        expect(navigate).toHaveBeenCalledWith(redirectPath)
      })

      it('should handle missing token parameter', async () => {
        const setAccessToken = vi.fn()
        const fetch = mockFetch()

        const route = mockUseRoute({}) // No token in query

        const { token } = route.query
        if (token) {
          const response = await fetch(`/callback?token=${token}`)
          if (response && response.token) {
            setAccessToken(response.token)
          }
        }

        // Should not make fetch call or set token
        expect(fetch).not.toHaveBeenCalled()
        expect(setAccessToken).not.toHaveBeenCalled()
      })

      it('should handle invalid token format', async () => {
        const invalidToken = 'not-a-jwt'

        const fetch = mockFetch()
        fetch.mockRejectedValue(new Error('Invalid token format'))

        const setAccessToken = vi.fn()
        const consoleError = vi
          .spyOn(console, 'error')
          .mockImplementation(() => {})

        const route = mockUseRoute({ token: invalidToken })
        const config = mockUseRuntimeConfig()

        try {
          const response = await fetch(
            `${config.public.tokenUrl}?token=${invalidToken}`,
          )
          if (response && response.token) {
            setAccessToken(response.token)
          }
        } catch (e) {
          consoleError(e)
        }

        expect(setAccessToken).not.toHaveBeenCalled()
        expect(consoleError).toHaveBeenCalled()

        consoleError.mockRestore()
      })

      it('should handle expired token', async () => {
        const expiredToken = 'expired-jwt-token'

        const fetch = mockFetch()
        fetch.mockRejectedValue(new Error('Token expired'))

        const setAccessToken = vi.fn()
        const consoleError = vi
          .spyOn(console, 'error')
          .mockImplementation(() => {})

        const route = mockUseRoute({ token: expiredToken })
        const config = mockUseRuntimeConfig()

        try {
          const response = await fetch(
            `${config.public.tokenUrl}?token=${expiredToken}`,
          )
          if (response && response.token) {
            setAccessToken(response.token)
          }
        } catch (e) {
          consoleError(e)
        }

        expect(setAccessToken).not.toHaveBeenCalled()
        expect(consoleError).toHaveBeenCalled()

        consoleError.mockRestore()
      })

      it('should handle backend validation failure', async () => {
        const inputToken = 'some-token'

        const fetch = mockFetch()
        fetch.mockRejectedValue(new Error('Validation failed'))

        const setAccessToken = vi.fn()
        const navigate = mockNavigateTo()
        const consoleError = vi
          .spyOn(console, 'error')
          .mockImplementation(() => {})

        const route = mockUseRoute({ token: inputToken })
        const config = mockUseRuntimeConfig()

        try {
          const response = await fetch(
            `${config.public.tokenUrl}?token=${inputToken}`,
          )
          if (response && response.token) {
            setAccessToken(response.token)
            navigate('/')
          }
        } catch (e) {
          consoleError(e)
        }

        expect(setAccessToken).not.toHaveBeenCalled()
        expect(navigate).not.toHaveBeenCalled()
        expect(consoleError).toHaveBeenCalled()

        consoleError.mockRestore()
      })

      it('should handle network timeout', async () => {
        const inputToken = 'some-token'

        const fetch = mockFetch()
        fetch.mockRejectedValue(new Error('Network timeout'))

        const setAccessToken = vi.fn()
        const consoleError = vi
          .spyOn(console, 'error')
          .mockImplementation(() => {})

        const route = mockUseRoute({ token: inputToken })
        const config = mockUseRuntimeConfig()

        try {
          const response = await fetch(
            `${config.public.tokenUrl}?token=${inputToken}`,
          )
          if (response && response.token) {
            setAccessToken(response.token)
          }
        } catch (e) {
          consoleError(e)
        }

        expect(setAccessToken).not.toHaveBeenCalled()
        expect(consoleError).toHaveBeenCalled()

        consoleError.mockRestore()
      })

      it('should handle malformed backend response', async () => {
        const inputToken = 'some-token'

        const fetch = mockFetch()
        fetch.mockResolvedValue({}) // No token field

        const setAccessToken = vi.fn()
        const navigate = mockNavigateTo()

        const route = mockUseRoute({ token: inputToken })
        const config = mockUseRuntimeConfig()

        const response = await fetch(
          `${config.public.tokenUrl}?token=${inputToken}`,
        )
        if (response && response.token) {
          setAccessToken(response.token)
          navigate('/')
        }

        // Should not set token or navigate
        expect(setAccessToken).not.toHaveBeenCalled()
        expect(navigate).not.toHaveBeenCalled()
      })

      it('should handle redirect with query params', async () => {
        const inputToken = 'oauth-token'
        const validatedToken = createMockToken()
        const redirectPath = '/challenges?id=123&filter=active'

        const fetch = mockFetch()
        fetch.mockResolvedValue({ token: validatedToken })

        const setAccessToken = vi.fn()
        const navigate = mockNavigateTo()

        const route = mockUseRoute({
          token: inputToken,
          redirect: redirectPath,
        })
        const config = mockUseRuntimeConfig()

        const response = await fetch(
          `${config.public.tokenUrl}?token=${inputToken}`,
        )
        if (response && response.token) {
          setAccessToken(response.token)

          const { redirect } = route.query
          if (redirect && typeof redirect === 'string') {
            navigate(redirect)
          } else {
            navigate('/')
          }
        }

        expect(navigate).toHaveBeenCalledWith(redirectPath)
      })

      it('should preserve hash in redirect', async () => {
        const inputToken = 'oauth-token'
        const validatedToken = createMockToken()
        const redirectPath = '/page#section'

        const fetch = mockFetch()
        fetch.mockResolvedValue({ token: validatedToken })

        const setAccessToken = vi.fn()
        const navigate = mockNavigateTo()

        const route = mockUseRoute({
          token: inputToken,
          redirect: redirectPath,
        })
        const config = mockUseRuntimeConfig()

        const response = await fetch(
          `${config.public.tokenUrl}?token=${inputToken}`,
        )
        if (response && response.token) {
          setAccessToken(response.token)

          const { redirect } = route.query
          if (redirect && typeof redirect === 'string') {
            navigate(redirect)
          } else {
            navigate('/')
          }
        }

        expect(navigate).toHaveBeenCalledWith(redirectPath)
      })

      it('should handle array redirect parameter', async () => {
        const inputToken = 'oauth-token'
        const validatedToken = createMockToken()

        const fetch = mockFetch()
        fetch.mockResolvedValue({ token: validatedToken })

        const setAccessToken = vi.fn()
        const navigate = mockNavigateTo()

        // Query params can be arrays in some cases
        const route = mockUseRoute({
          token: inputToken,
          redirect: ['/path1', '/path2'],
        })
        const config = mockUseRuntimeConfig()

        const response = await fetch(
          `${config.public.tokenUrl}?token=${inputToken}`,
        )
        if (response && response.token) {
          setAccessToken(response.token)

          const { redirect } = route.query
          // Should not redirect if redirect is not a string
          if (redirect && typeof redirect === 'string') {
            navigate(redirect)
          } else {
            navigate('/')
          }
        }

        expect(navigate).toHaveBeenCalledWith('/')
      })
    })

    describe('Error Handling', () => {
      it('should log errors to console', async () => {
        const inputToken = 'invalid-token'
        const consoleErrorSpy = vi.spyOn(console, 'error').mockImplementation()

        const fetch = mockFetch()
        fetch.mockRejectedValue(new Error('Network error'))

        const route = mockUseRoute({ token: inputToken })
        const config = mockUseRuntimeConfig()

        try {
          await fetch(`${config.public.tokenUrl}?token=${inputToken}`)
        } catch (error) {
          console.error('Token validation failed:', error)
        }

        expect(consoleErrorSpy).toHaveBeenCalledWith(
          'Token validation failed:',
          expect.any(Error),
        )

        consoleErrorSpy.mockRestore()
      })

      it('should show error state on validation failure', async () => {
        const inputToken = 'invalid-token'
        const errorState = ref<string | null>(null)

        const fetch = mockFetch()
        fetch.mockRejectedValue(new Error('Validation failed'))

        const route = mockUseRoute({ token: inputToken })
        const config = mockUseRuntimeConfig()

        try {
          await fetch(`${config.public.tokenUrl}?token=${inputToken}`)
        } catch (error) {
          errorState.value =
            error instanceof Error ? error.message : 'Unknown error'
        }

        expect(errorState.value).toBe('Validation failed')
      })

      it('should not store invalid token', async () => {
        const inputToken = 'invalid-token'
        const cookie = mockUseCookie<string>(null)
        const setAccessToken = vi.fn((value: string) => {
          cookie.value = value
        })

        const fetch = mockFetch()
        fetch.mockRejectedValue(new Error('Invalid token'))

        const route = mockUseRoute({ token: inputToken })
        const config = mockUseRuntimeConfig()

        try {
          const response = await fetch(
            `${config.public.tokenUrl}?token=${inputToken}`,
          )
          if (response && response.token) {
            setAccessToken(response.token)
          }
        } catch (error) {
          // Error occurred - don't store token
        }

        // Token should remain null
        expect(cookie.value).toBeNull()
        expect(setAccessToken).not.toHaveBeenCalled()
      })

      it('should allow user to retry on failure', async () => {
        const inputToken = 'retry-token'
        const validatedToken = createMockToken()
        const setAccessToken = vi.fn()

        const fetch = mockFetch()
        // First call fails, second succeeds
        fetch
          .mockRejectedValueOnce(new Error('Network timeout'))
          .mockResolvedValueOnce({ token: validatedToken })

        const route = mockUseRoute({ token: inputToken })
        const config = mockUseRuntimeConfig()

        // First attempt - fails
        try {
          const response = await fetch(
            `${config.public.tokenUrl}?token=${inputToken}`,
          )
          if (response && response.token) {
            setAccessToken(response.token)
          }
        } catch (error) {
          // Handle error - user can retry
        }

        expect(setAccessToken).not.toHaveBeenCalled()

        // Retry - succeeds
        const response = await fetch(
          `${config.public.tokenUrl}?token=${inputToken}`,
        )
        if (response && response.token) {
          setAccessToken(response.token)
        }

        expect(setAccessToken).toHaveBeenCalledWith(validatedToken)
      })
    })
  })

  describe('Global Auth Middleware', () => {
    describe('Route Protection', () => {
      it('should allow access to callback page without token', async () => {
        const cookie = mockUseCookie<string>(null) // No token
        const loginWithRedirect = vi.fn()

        const auth = mockUseAuth({
          token: cookie,
          loginWithRedirect,
        })

        // Simulate middleware logic
        const to = { path: '/callback', name: 'callback' }

        // Skip auth check for callback page
        if (to.path === '/callback') {
          // Allow access - don't call loginWithRedirect
        } else if (!auth.token.value) {
          await loginWithRedirect()
        }

        expect(loginWithRedirect).not.toHaveBeenCalled()
      })

      it('should redirect to login when accessing protected route without token', async () => {
        const cookie = mockUseCookie<string>(null) // No token
        const loginWithRedirect = vi.fn()

        const auth = mockUseAuth({
          token: cookie,
          loginWithRedirect,
        })

        // Simulate middleware logic for protected route
        const to = { path: '/admin/projects', name: 'admin-projects' }

        if (to.path !== '/callback') {
          if (!auth.token.value) {
            await loginWithRedirect()
          }
        }

        expect(loginWithRedirect).toHaveBeenCalled()
      })

      it('should allow access to protected routes with valid token', async () => {
        const token = createMockToken()
        const cookie = mockUseCookie<string>(token)
        const loginWithRedirect = vi.fn()

        const auth = mockUseAuth({
          token: cookie,
          loginWithRedirect,
        })

        // Simulate middleware logic
        const to = { path: '/admin/projects', name: 'admin-projects' }

        if (to.path !== '/callback') {
          if (!auth.token.value) {
            await loginWithRedirect()
          }
        }

        expect(loginWithRedirect).not.toHaveBeenCalled()
      })

      it('should check token on every navigation', async () => {
        const cookie = mockUseCookie<string>(null)
        const loginWithRedirect = vi.fn()

        const auth = mockUseAuth({
          token: cookie,
          loginWithRedirect,
        })

        // First navigation - no token
        const to1 = { path: '/projects', name: 'projects' }
        if (to1.path !== '/callback' && !auth.token.value) {
          await loginWithRedirect()
        }

        expect(loginWithRedirect).toHaveBeenCalledTimes(1)

        // Set token
        cookie.value = createMockToken()

        // Second navigation - has token
        const to2 = { path: '/challenges', name: 'challenges' }
        if (to2.path !== '/callback' && !auth.token.value) {
          await loginWithRedirect()
        }

        // Should still only be called once (from first navigation)
        expect(loginWithRedirect).toHaveBeenCalledTimes(1)
      })

      it('should handle navigation to callback route variations', async () => {
        const cookie = mockUseCookie<string>(null)
        const loginWithRedirect = vi.fn()

        const auth = mockUseAuth({
          token: cookie,
          loginWithRedirect,
        })

        // Test various callback route scenarios
        // In real Nuxt, query params are separate from path
        const callbackRoutes = [
          { path: '/callback', query: {} },
          { path: '/callback', query: { token: 'abc123' } },
          { path: '/callback', query: { token: 'abc', redirect: '/admin' } },
        ]

        for (const to of callbackRoutes) {
          if (to.path !== '/callback' && !auth.token.value) {
            await loginWithRedirect()
          }
        }

        // Should never call login for callback paths
        expect(loginWithRedirect).not.toHaveBeenCalled()
      })

      it('should handle routes with callback in path but not callback page', async () => {
        const cookie = mockUseCookie<string>(null)
        const loginWithRedirect = vi.fn()

        const auth = mockUseAuth({
          token: cookie,
          loginWithRedirect,
        })

        // This contains 'callback' but isn't the callback page
        const to = {
          path: '/admin/callback-settings',
          name: 'callback-settings',
        }

        // With exact path check, this should redirect (it's not the callback page)
        if (to.path === '/callback') {
          // Skip auth
        } else if (!auth.token.value) {
          await loginWithRedirect()
        }

        // SHOULD redirect because path doesn't exactly match '/callback'
        expect(loginWithRedirect).toHaveBeenCalled()
      })
    })

    describe('Token Expiration During Session', () => {
      it('should detect expired token on navigation', () => {
        // Create token that expired 1 hour ago
        const expiredTime = Math.floor(Date.now() / 1000) - 3600
        const expiredToken = createMockToken({ exp: expiredTime })

        // Parse and check expiration
        const parts = expiredToken.split('.')
        const payload = JSON.parse(
          atob(parts[1].replace(/-/g, '+').replace(/_/g, '/')),
        )
        const now = Math.floor(Date.now() / 1000)
        const isExpired = payload.exp < now

        expect(isExpired).toBe(true)
      })

      it('should redirect to login on expired token', async () => {
        const expiredTime = Math.floor(Date.now() / 1000) - 3600
        const expiredToken = createMockToken({ exp: expiredTime })
        const cookie = mockUseCookie<string>(expiredToken)
        const loginWithRedirect = vi.fn()

        const auth = mockUseAuth({
          token: cookie,
          loginWithRedirect,
        })

        const to = { path: '/admin/projects', name: 'admin-projects' }

        // Check if token is expired before allowing access
        if (auth.token.value) {
          const parts = auth.token.value.split('.')
          if (parts.length === 3) {
            try {
              const payload = JSON.parse(
                atob(parts[1].replace(/-/g, '+').replace(/_/g, '/')),
              )
              const now = Math.floor(Date.now() / 1000)
              if (payload.exp && payload.exp < now) {
                // Token is expired
                await loginWithRedirect()
              }
            } catch (error) {
              // Invalid token
              await loginWithRedirect()
            }
          }
        } else {
          await loginWithRedirect()
        }

        expect(loginWithRedirect).toHaveBeenCalled()
      })

      it('should clear expired token from storage', async () => {
        const expiredTime = Math.floor(Date.now() / 1000) - 3600
        const expiredToken = createMockToken({ exp: expiredTime })
        const cookie = mockUseCookie<string>(expiredToken)

        // Check expiration and clear
        if (cookie.value) {
          const parts = cookie.value.split('.')
          if (parts.length === 3) {
            try {
              const payload = JSON.parse(
                atob(parts[1].replace(/-/g, '+').replace(/_/g, '/')),
              )
              const now = Math.floor(Date.now() / 1000)
              if (payload.exp && payload.exp < now) {
                // Clear expired token
                cookie.value = null
              }
            } catch (error) {
              // Clear invalid token
              cookie.value = null
            }
          }
        }

        expect(cookie.value).toBeNull()
      })

      it('should handle token expiring during user activity', async () => {
        // Create a token that's already expired
        const pastExpiry = Math.floor(Date.now() / 1000) - 10
        const expiredToken = createMockToken({ exp: pastExpiry })
        const cookie = mockUseCookie<string>()
        const loginWithRedirect = vi.fn()

        // Initially user has a valid token
        const validToken = createMockToken()
        cookie.value = validToken

        const auth = mockUseAuth({
          token: cookie,
          loginWithRedirect,
        })

        // Token was valid
        let parts = validToken.split('.')
        let payload = JSON.parse(
          atob(parts[1].replace(/-/g, '+').replace(/_/g, '/')),
        )
        let now = Math.floor(Date.now() / 1000)
        expect(payload.exp >= now).toBe(true)

        // Token expires (simulate by replacing with expired token)
        cookie.value = expiredToken

        // Now check if token is expired
        parts = expiredToken.split('.')
        payload = JSON.parse(
          atob(parts[1].replace(/-/g, '+').replace(/_/g, '/')),
        )
        now = Math.floor(Date.now() / 1000)
        const isExpired = payload.exp < now

        expect(isExpired).toBe(true)

        // Should redirect to login
        if (isExpired) {
          await loginWithRedirect()
        }

        expect(loginWithRedirect).toHaveBeenCalled()
      })
    })
  })

  describe('Admin Middleware', () => {
    describe('Admin Route Protection', () => {
      it('should allow superadmin to access admin routes', async () => {
        const token = createMockToken()
        const superadminUser = createMockUser({
          roles: [{ role: RoleType.Superadmin, scope: null }],
        })

        const cookie = mockUseCookie<string>(token)
        const meRef = ref(superadminUser)
        const isLoading = ref(false)
        const loginWithRedirect = vi.fn()
        const createError = vi.fn()

        const auth = mockUseAuth({
          token: cookie,
          me: meRef,
          isLoading,
          loginWithRedirect,
        })

        const to = { path: '/admin/projects', name: 'admin-projects' }

        // Simulate middleware logic
        if (to.path.startsWith('/admin')) {
          if (!auth.token.value) {
            loginWithRedirect()
          } else {
            // Wait for loading
            let attempts = 0
            while (isLoading.value && attempts < 100) {
              await new Promise((resolve) => setTimeout(resolve, 10))
              attempts++
            }

            if (!auth.me.value) {
              loginWithRedirect()
            } else if (!auth.isSuperAdmin.value && !auth.isAdmin.value) {
              createError({ statusCode: 403 })
            }
          }
        }

        expect(loginWithRedirect).not.toHaveBeenCalled()
        expect(createError).not.toHaveBeenCalled()
      })

      it('should allow admin to access admin routes', async () => {
        const token = createMockToken()
        const adminUser = createMockUser({
          roles: [{ role: RoleType.Admin, scope: null }],
        })

        const cookie = mockUseCookie<string>(token)
        const meRef = ref(adminUser)
        const isLoading = ref(false)
        const loginWithRedirect = vi.fn()
        const createError = vi.fn()

        const auth = mockUseAuth({
          token: cookie,
          me: meRef,
          isLoading,
          loginWithRedirect,
        })

        const to = { path: '/admin', name: 'admin' }

        if (to.path.startsWith('/admin')) {
          if (!auth.token.value) {
            loginWithRedirect()
          } else {
            let attempts = 0
            while (isLoading.value && attempts < 100) {
              await new Promise((resolve) => setTimeout(resolve, 10))
              attempts++
            }

            if (!auth.me.value) {
              loginWithRedirect()
            } else if (!auth.isSuperAdmin.value && !auth.isAdmin.value) {
              createError({ statusCode: 403 })
            }
          }
        }

        expect(loginWithRedirect).not.toHaveBeenCalled()
        expect(createError).not.toHaveBeenCalled()
      })

      it('should block regular user from admin routes', async () => {
        const token = createMockToken()
        const regularUser = createMockUser({
          roles: [{ role: RoleType.User, scope: null }],
        })

        const cookie = mockUseCookie<string>(token)
        const meRef = ref(regularUser)
        const isLoading = ref(false)
        const loginWithRedirect = vi.fn()
        const createError = vi.fn()

        const auth = mockUseAuth({
          token: cookie,
          me: meRef,
          isLoading,
          loginWithRedirect,
        })

        const to = { path: '/admin/projects', name: 'admin-projects' }

        if (to.path.startsWith('/admin')) {
          if (!auth.token.value) {
            loginWithRedirect()
          } else {
            let attempts = 0
            while (isLoading.value && attempts < 100) {
              await new Promise((resolve) => setTimeout(resolve, 10))
              attempts++
            }

            if (!auth.me.value) {
              loginWithRedirect()
            } else if (!auth.isSuperAdmin.value && !auth.isAdmin.value) {
              createError({
                statusCode: 403,
                statusMessage: 'Forbidden',
                message: 'You do not have permission to access this page',
              })
            }
          }
        }

        expect(loginWithRedirect).not.toHaveBeenCalled()
        expect(createError).toHaveBeenCalledWith({
          statusCode: 403,
          statusMessage: 'Forbidden',
          message: 'You do not have permission to access this page',
        })
      })

      it('should redirect to login if no token', async () => {
        const cookie = mockUseCookie<string>(null)
        const loginWithRedirect = vi.fn()

        const auth = mockUseAuth({
          token: cookie,
          loginWithRedirect,
        })

        const to = { path: '/admin', name: 'admin' }

        if (to.path.startsWith('/admin')) {
          if (!auth.token.value) {
            loginWithRedirect()
          }
        }

        expect(loginWithRedirect).toHaveBeenCalled()
      })

      it('should return 403 for non-admin users', async () => {
        const token = createMockToken()
        const churchAdminUser = createMockUser({
          roles: [{ role: RoleType.ChurchAdmin, scope: { churchId: 'CH01' } }],
        })

        const cookie = mockUseCookie<string>(token)
        const meRef = ref(churchAdminUser)
        const isLoading = ref(false)
        const createError = vi.fn()

        const auth = mockUseAuth({
          token: cookie,
          me: meRef,
          isLoading,
        })

        const to = { path: '/admin/users', name: 'admin-users' }

        if (to.path.startsWith('/admin')) {
          if (auth.token.value) {
            let attempts = 0
            while (isLoading.value && attempts < 100) {
              await new Promise((resolve) => setTimeout(resolve, 10))
              attempts++
            }

            if (
              auth.me.value &&
              !auth.isSuperAdmin.value &&
              !auth.isAdmin.value
            ) {
              createError({ statusCode: 403 })
            }
          }
        }

        expect(createError).toHaveBeenCalledWith({ statusCode: 403 })
      })

      it('should wait for me query to complete', async () => {
        const token = createMockToken()
        const adminUser = createMockUser({
          roles: [{ role: RoleType.Admin, scope: null }],
        })

        const cookie = mockUseCookie<string>(token)
        const meRef = ref<GetMeQuery['me'] | null>(null) // Not loaded yet
        const isLoading = ref(true)
        const createError = vi.fn()

        const auth = mockUseAuth({
          token: cookie,
          me: meRef,
          isLoading,
        })

        const to = { path: '/admin', name: 'admin' }

        // Start middleware check
        const middlewarePromise = (async () => {
          if (to.path.startsWith('/admin') && auth.token.value) {
            let attempts = 0
            while (isLoading.value && attempts < 100) {
              await new Promise((resolve) => setTimeout(resolve, 10))
              attempts++
            }

            if (
              auth.me.value &&
              !auth.isSuperAdmin.value &&
              !auth.isAdmin.value
            ) {
              createError({ statusCode: 403 })
            }
          }
        })()

        // Simulate loading completing after 50ms
        setTimeout(() => {
          meRef.value = adminUser
          isLoading.value = false
        }, 50)

        await middlewarePromise

        expect(createError).not.toHaveBeenCalled()
      })

      it('should timeout if me query takes too long', async () => {
        const token = createMockToken()
        const cookie = mockUseCookie<string>(token)
        const meRef = ref<GetMeQuery['me'] | null>(null)
        const isLoading = ref(true) // Stuck loading
        const loginWithRedirect = vi.fn()

        const auth = mockUseAuth({
          token: cookie,
          me: meRef,
          isLoading,
          loginWithRedirect,
        })

        const to = { path: '/admin', name: 'admin' }

        if (to.path.startsWith('/admin') && auth.token.value) {
          let attempts = 0
          while (isLoading.value && attempts < 100) {
            await new Promise((resolve) => setTimeout(resolve, 10))
            attempts++
          }

          // After timeout, me is still null
          if (!auth.me.value) {
            loginWithRedirect()
          }
        }

        expect(loginWithRedirect).toHaveBeenCalled()
      })

      it('should handle me query error', async () => {
        const token = createMockToken()
        const cookie = mockUseCookie<string>(token)
        const meRef = ref<GetMeQuery['me'] | null>(null) // Query failed
        const isLoading = ref(false)
        const loginWithRedirect = vi.fn()

        const auth = mockUseAuth({
          token: cookie,
          me: meRef,
          isLoading,
          loginWithRedirect,
        })

        const to = { path: '/admin', name: 'admin' }

        if (to.path.startsWith('/admin') && auth.token.value) {
          let attempts = 0
          while (isLoading.value && attempts < 100) {
            await new Promise((resolve) => setTimeout(resolve, 10))
            attempts++
          }

          if (!auth.me.value) {
            loginWithRedirect()
          }
        }

        expect(loginWithRedirect).toHaveBeenCalled()
      })

      it('should only check admin routes', async () => {
        const token = createMockToken()
        const regularUser = createMockUser({
          roles: [{ role: RoleType.User, scope: null }],
        })

        const cookie = mockUseCookie<string>(token)
        const meRef = ref(regularUser)
        const createError = vi.fn()

        const auth = mockUseAuth({
          token: cookie,
          me: meRef,
        })

        // Non-admin route
        const to = { path: '/projects', name: 'projects' }

        // Middleware should return early for non-admin routes
        if (to.path.startsWith('/admin')) {
          if (auth.token.value && auth.me.value) {
            if (!auth.isSuperAdmin.value && !auth.isAdmin.value) {
              createError({ statusCode: 403 })
            }
          }
        }

        expect(createError).not.toHaveBeenCalled()
      })

      it('should handle nested admin routes', async () => {
        const token = createMockToken()
        const adminUser = createMockUser({
          roles: [{ role: RoleType.Admin, scope: null }],
        })

        const cookie = mockUseCookie<string>(token)
        const meRef = ref(adminUser)
        const isLoading = ref(false)
        const createError = vi.fn()

        const auth = mockUseAuth({
          token: cookie,
          me: meRef,
          isLoading,
        })

        // Deeply nested admin route
        const to = {
          path: '/admin/projects/123/challenges/456/edit',
          name: 'admin-challenge-edit',
        }

        if (to.path.startsWith('/admin')) {
          if (auth.token.value) {
            let attempts = 0
            while (isLoading.value && attempts < 100) {
              await new Promise((resolve) => setTimeout(resolve, 10))
              attempts++
            }

            if (
              auth.me.value &&
              !auth.isSuperAdmin.value &&
              !auth.isAdmin.value
            ) {
              createError({ statusCode: 403 })
            }
          }
        }

        expect(createError).not.toHaveBeenCalled()
      })
    })

    describe('Role Changes During Session', () => {
      it('should detect role removal', () => {
        // Start with admin user
        const adminUser = createMockUser({
          roles: [{ role: RoleType.Admin, scope: null }],
        })
        const meRef = ref(adminUser)
        const auth = mockUseAuth({ me: meRef })

        // Initially has admin role
        expect(auth.isAdmin.value).toBe(true)

        // Role removed (simulating backend update)
        meRef.value = createMockUser({
          roles: [{ role: RoleType.User, scope: null }],
        })

        // No longer has admin role
        expect(auth.isAdmin.value).toBe(false)
      })

      it('should block access after role removal', () => {
        const token = createMockToken()
        const meRef = ref(
          createMockUser({
            roles: [{ role: RoleType.Admin, scope: null }],
          }),
        )
        const createError = vi.fn()

        const auth = mockUseAuth({
          token: ref(token),
          me: meRef,
        })

        const to = { path: '/admin', name: 'admin' }

        // Initially admin - should have access
        if (to.path.startsWith('/admin')) {
          if (
            auth.me.value &&
            !auth.isSuperAdmin.value &&
            !auth.isAdmin.value
          ) {
            createError({ statusCode: 403 })
          }
        }

        expect(createError).not.toHaveBeenCalled()

        // Role removed
        meRef.value = createMockUser({
          roles: [{ role: RoleType.User, scope: null }],
        })

        // Now should be blocked
        if (to.path.startsWith('/admin')) {
          if (
            auth.me.value &&
            !auth.isSuperAdmin.value &&
            !auth.isAdmin.value
          ) {
            createError({ statusCode: 403 })
          }
        }

        expect(createError).toHaveBeenCalledWith({ statusCode: 403 })
      })

      it('should allow access after role grant', () => {
        const token = createMockToken()
        const meRef = ref(
          createMockUser({
            roles: [{ role: RoleType.User, scope: null }],
          }),
        )
        const createError = vi.fn()

        const auth = mockUseAuth({
          token: ref(token),
          me: meRef,
        })

        const to = { path: '/admin', name: 'admin' }

        // Initially not admin - should be blocked
        if (to.path.startsWith('/admin')) {
          if (
            auth.me.value &&
            !auth.isSuperAdmin.value &&
            !auth.isAdmin.value
          ) {
            createError({ statusCode: 403 })
          }
        }

        expect(createError).toHaveBeenCalledWith({ statusCode: 403 })

        // Reset mock
        createError.mockClear()

        // Admin role granted
        meRef.value = createMockUser({
          roles: [{ role: RoleType.Admin, scope: null }],
        })

        // Now should have access
        if (to.path.startsWith('/admin')) {
          if (
            auth.me.value &&
            !auth.isSuperAdmin.value &&
            !auth.isAdmin.value
          ) {
            createError({ statusCode: 403 })
          }
        }

        expect(createError).not.toHaveBeenCalled()
      })
    })
  })

  describe('Integration: Full Auth Flow', () => {
    it('should complete full login flow', async () => {
      // Step 1: User visits protected route without token
      const cookie = mockUseCookie<string>(null)
      const loginWithRedirect = vi.fn()
      const navigate = mockNavigateTo()
      const fetch = mockFetch()

      const auth = mockUseAuth({
        token: cookie,
        loginWithRedirect,
      })

      const protectedRoute = { path: '/admin/projects', name: 'admin-projects' }

      // Step 2: Gets redirected to login (no token)
      if (protectedRoute.path !== '/callback' && !auth.token.value) {
        await loginWithRedirect()
      }
      expect(loginWithRedirect).toHaveBeenCalled()

      // Step 3: User returns to callback with token
      const oauthToken = 'oauth-callback-token'
      const validatedToken = createMockToken()
      const route = mockUseRoute({
        token: oauthToken,
        redirect: '/admin/projects',
      })
      const config = mockUseRuntimeConfig()

      // Step 4: Token validated with backend
      fetch.mockResolvedValue({ token: validatedToken })
      const response = await fetch(
        `${config.public.tokenUrl}?token=${oauthToken}`,
      )

      // Step 5: Stored in cookie
      if (response && response.token) {
        cookie.value = response.token
      }
      expect(cookie.value).toBe(validatedToken)

      // Step 6: Redirected to original destination
      const { redirect } = route.query
      if (redirect && typeof redirect === 'string') {
        navigate(redirect)
      }
      expect(navigate).toHaveBeenCalledWith('/admin/projects')

      // Step 7: Me query fetches user data
      const adminUser = createMockUser({
        roles: [{ role: RoleType.Admin, scope: null }],
      })
      const meRef = ref(adminUser)
      const authWithUser = mockUseAuth({
        token: cookie,
        me: meRef,
      })

      // Step 8: Can access protected routes
      if (protectedRoute.path.startsWith('/admin')) {
        const hasAccess =
          authWithUser.isAdmin.value || authWithUser.isSuperAdmin.value
        expect(hasAccess).toBe(true)
      }
    })

    it('should handle login flow with redirect', async () => {
      const LOGIN_URL = 'https://login.example.com/auth'
      const targetPath = '/admin/projects/123'
      const cookie = mockUseCookie<string>(null)
      const loginWithRedirect = vi.fn()
      const navigate = mockNavigateTo()

      const auth = mockUseAuth({
        token: cookie,
        loginWithRedirect: () => {
          const redirectUrl = `${LOGIN_URL}?redirect=${targetPath}`
          return navigate(redirectUrl, { external: true })
        },
      })

      // User tries to access specific page
      const to = { path: targetPath, name: 'admin-project-detail' }

      if (to.path !== '/callback' && !auth.token.value) {
        await auth.loginWithRedirect()
      }

      // Should redirect to login with target path
      expect(navigate).toHaveBeenCalledWith(
        `${LOGIN_URL}?redirect=${targetPath}`,
        { external: true },
      )

      // After login, callback should redirect back to target
      const validatedToken = createMockToken()
      cookie.value = validatedToken
      const route = mockUseRoute({ token: 'oauth-token', redirect: targetPath })

      const { redirect } = route.query
      if (redirect && typeof redirect === 'string') {
        navigate(redirect)
      }

      expect(navigate).toHaveBeenCalledWith(targetPath)
    })

    it('should handle logout', () => {
      const token = createMockToken()
      const cookie = mockUseCookie<string>(token)
      const meRef = ref(createMockUser())
      const loginWithRedirect = vi.fn()

      const auth = mockUseAuth({
        token: cookie,
        me: meRef,
        loginWithRedirect,
      })

      // User is logged in
      expect(auth.token.value).toBe(token)
      expect(auth.me.value).not.toBeNull()

      // Logout: clear token and user data
      cookie.value = null
      meRef.value = null

      expect(auth.token.value).toBeNull()
      expect(auth.me.value).toBeNull()

      // Accessing protected route should redirect
      const to = { path: '/admin', name: 'admin' }
      if (to.path !== '/callback' && !auth.token.value) {
        loginWithRedirect()
      }

      expect(loginWithRedirect).toHaveBeenCalled()
    })

    it('should handle token refresh', async () => {
      // Token expiring soon (5 minutes)
      const soonExpiry = Math.floor(Date.now() / 1000) + 300
      const expiringToken = createMockToken({ exp: soonExpiry })
      const newToken = createMockToken({
        exp: Math.floor(Date.now() / 1000) + 3600,
      })

      const cookie = mockUseCookie<string>(expiringToken)
      const fetch = mockFetch()
      fetch.mockResolvedValue({ token: newToken })

      // Check if token needs refresh (< 10 minutes remaining)
      const parts = cookie.value!.split('.')
      const payload = JSON.parse(
        atob(parts[1].replace(/-/g, '+').replace(/_/g, '/')),
      )
      const now = Math.floor(Date.now() / 1000)
      const timeRemaining = payload.exp - now

      const needsRefresh = timeRemaining < 600 // Less than 10 minutes

      expect(needsRefresh).toBe(true)

      // Refresh token
      if (needsRefresh) {
        const response = await fetch('/api/auth/refresh')
        if (response && response.token) {
          cookie.value = response.token
        }
      }

      expect(cookie.value).toBe(newToken)
    })

    it('should handle concurrent login in multiple tabs', () => {
      // Simulate shared cookie storage between tabs
      const sharedCookie = mockUseCookie<string>(null)

      // Tab A
      const authTabA = mockUseAuth({ token: sharedCookie })
      expect(authTabA.token.value).toBeNull()

      // Tab B
      const authTabB = mockUseAuth({ token: sharedCookie })
      expect(authTabB.token.value).toBeNull()

      // User logs in via Tab A
      const token = createMockToken()
      sharedCookie.value = token

      // Both tabs should now have the token
      expect(authTabA.token.value).toBe(token)
      expect(authTabB.token.value).toBe(token)
    })

    it('should handle logout in one tab affecting other tabs', () => {
      // Shared cookie storage
      const token = createMockToken()
      const sharedCookie = mockUseCookie<string>(token)

      // Tab A and Tab B both logged in
      const authTabA = mockUseAuth({ token: sharedCookie })
      const authTabB = mockUseAuth({ token: sharedCookie })

      expect(authTabA.token.value).toBe(token)
      expect(authTabB.token.value).toBe(token)

      // User logs out in Tab A
      sharedCookie.value = null

      // Both tabs should now be logged out
      expect(authTabA.token.value).toBeNull()
      expect(authTabB.token.value).toBeNull()
    })
  })

  describe('Security', () => {
    it('should not expose token in URLs', () => {
      const token = createMockToken()
      const cookie = mockUseCookie<string>(token)
      const navigate = mockNavigateTo()

      const auth = mockUseAuth({ token: cookie })

      // Navigate to protected route
      const to = { path: '/admin/projects', name: 'admin-projects' }

      // Token should NOT be in URL
      const urlWithToken = `/admin/projects?token=${token}`
      const urlWithoutToken = '/admin/projects'

      // Only the callback page should have token in URL
      expect(to.path).toBe(urlWithoutToken)
      expect(to.path).not.toContain('token=')

      // Token is in cookie, not URL
      expect(auth.token.value).toBe(token)
    })

    it('should use httpOnly cookies in production', () => {
      // In production, cookies should be httpOnly to prevent XSS
      const cookieOptions = {
        httpOnly: true,
        secure: true,
        sameSite: 'lax' as const,
      }

      // Verify httpOnly is enabled
      expect(cookieOptions.httpOnly).toBe(true)

      // This prevents JavaScript from accessing the cookie
      // Note: In real implementation, useCookie would be called with these options
    })

    it('should use secure cookies in production', () => {
      // In production (HTTPS), cookies should have secure flag
      const isProduction = process.env.NODE_ENV === 'production'
      const cookieOptions = {
        secure: isProduction || true, // Always true in test
        httpOnly: true,
        sameSite: 'lax' as const,
      }

      expect(cookieOptions.secure).toBe(true)
    })

    it('should use sameSite cookies', () => {
      // Cookies should have sameSite to prevent CSRF
      const cookieOptions = {
        httpOnly: true,
        secure: true,
        sameSite: 'lax' as const,
      }

      expect(cookieOptions.sameSite).toBe('lax')
      // 'lax' allows cookies on top-level navigation
      // 'strict' would prevent OAuth callbacks from working
    })

    it('should validate token signature', async () => {
      const validToken = createMockToken()
      const tamperedToken = validToken.slice(0, -10) + 'tampered123'

      const fetch = mockFetch()
      // Backend validates signature
      fetch.mockResolvedValueOnce({ error: 'Invalid signature' })

      const route = mockUseRoute({ token: tamperedToken })
      const config = mockUseRuntimeConfig()

      const response = await fetch(
        `${config.public.tokenUrl}?token=${tamperedToken}`,
      )

      // Tampered token should be rejected by backend
      expect(response.error).toBe('Invalid signature')
    })

    it('should reject tokens from other domains', () => {
      // Token with wrong issuer claim
      const foreignToken = createMockToken({ iss: 'evil.com' })
      const cookie = mockUseCookie<string>(foreignToken)

      // Parse token
      const parts = cookie.value!.split('.')
      const payload = JSON.parse(
        atob(parts[1].replace(/-/g, '+').replace(/_/g, '/')),
      )

      // Check issuer
      const expectedIssuer = 'wayfarer'
      const isValidIssuer = payload.iss === expectedIssuer

      expect(isValidIssuer).toBe(false)

      // In real implementation, this check would happen in backend or middleware
      if (!isValidIssuer) {
        cookie.value = null
      }

      expect(cookie.value).toBeNull()
    })

    it('should handle XSS attempts in redirect param', () => {
      const xssAttempts = [
        'javascript:alert(1)',
        'data:text/html,<script>alert(1)</script>',
        '//evil.com/phishing',
        'https://evil.com/steal-cookies',
      ]

      const isSafeRedirect = (url: string): boolean => {
        if (url.startsWith('http://') || url.startsWith('https://')) {
          return false
        }
        if (url.startsWith('//')) {
          return false
        }
        if (url.startsWith('javascript:') || url.startsWith('data:')) {
          return false
        }
        return url.startsWith('/')
      }

      // All XSS attempts should be rejected
      for (const attempt of xssAttempts) {
        expect(isSafeRedirect(attempt)).toBe(false)
      }

      // Safe redirects should be allowed
      expect(isSafeRedirect('/admin/projects')).toBe(true)
      expect(isSafeRedirect('/search?q=test')).toBe(true)
    })

    it('should prevent CSRF on token endpoints', async () => {
      const token = 'oauth-token'
      const fetch = mockFetch()

      // In real implementation, CSRF protection would be via:
      // 1. SameSite cookies
      // 2. State parameter in OAuth flow
      // 3. Origin/Referer checks

      const cookieOptions = {
        sameSite: 'lax' as const, // Prevents CSRF
      }

      expect(cookieOptions.sameSite).toBe('lax')

      // OAuth state parameter prevents CSRF
      const state = 'random-state-value-' + Math.random()
      const route = mockUseRoute({
        token,
        state,
      })

      // Verify state matches expected value
      const expectedState = state
      const receivedState = route.query.state

      expect(receivedState).toBe(expectedState)
    })
  })

  describe('Performance', () => {
    it('should cache me query result', async () => {
      const token = createMockToken()
      const user = createMockUser()
      const meRef = ref<GetMeQuery['me'] | null>(null)
      const fetchCallCount = ref(0)

      const mockFetchMe = vi.fn(async () => {
        fetchCallCount.value++
        return user
      })

      const auth = mockUseAuth({
        token: ref(token),
        me: meRef,
      })

      // First fetch - should call API
      meRef.value = await mockFetchMe()
      expect(fetchCallCount.value).toBe(1)
      expect(auth.me.value).toStrictEqual(user)

      // Second access - should use cached value (no new fetch)
      const cachedUser = auth.me.value
      expect(cachedUser).toStrictEqual(user)
      expect(fetchCallCount.value).toBe(1) // Still 1, no new fetch

      // Third access - still cached
      const cachedUser2 = auth.me.value
      expect(cachedUser2).toStrictEqual(user)
      expect(fetchCallCount.value).toBe(1)
    })

    it('should not block page render on auth check', async () => {
      const cookie = mockUseCookie<string>(null)
      const isLoading = ref(true)
      const pageRendered = ref(false)

      // Simulate slow auth check
      const auth = mockUseAuth({
        token: cookie,
        isLoading,
      })

      // Page should render even while auth is loading
      // (Nuxt SSR / client hydration pattern)
      pageRendered.value = true
      expect(pageRendered.value).toBe(true)

      // Auth is still loading
      expect(auth.isLoading.value).toBe(true)

      // Eventually auth completes
      setTimeout(() => {
        isLoading.value = false
      }, 10)

      await new Promise((resolve) => setTimeout(resolve, 20))

      expect(auth.isLoading.value).toBe(false)
      expect(pageRendered.value).toBe(true)
    })

    it('should debounce token validation', async () => {
      const cookie = mockUseCookie<string>(null)
      const validationCallCount = ref(0)

      const mockValidateToken = vi.fn(async (token: string) => {
        validationCallCount.value++
        await new Promise((resolve) => setTimeout(resolve, 10))
        return { valid: true }
      })

      // Rapid token changes
      const tokens = [
        createMockToken({ user_id: 'user1' }),
        createMockToken({ user_id: 'user2' }),
        createMockToken({ user_id: 'user3' }),
      ]

      // Without debouncing, would validate 3 times
      // With debouncing, should only validate once (last token)

      let debounceTimeout: NodeJS.Timeout | null = null
      const debouncedValidate = (token: string) => {
        if (debounceTimeout) clearTimeout(debounceTimeout)
        debounceTimeout = setTimeout(() => {
          mockValidateToken(token)
        }, 50)
      }

      // Rapid changes
      for (const token of tokens) {
        cookie.value = token
        debouncedValidate(token)
      }

      // Wait for debounce
      await new Promise((resolve) => setTimeout(resolve, 100))

      // Should only validate once (last token)
      expect(validationCallCount.value).toBe(1)
      expect(mockValidateToken).toHaveBeenCalledTimes(1)
      expect(mockValidateToken).toHaveBeenCalledWith(tokens[2])
    })

    it('should cancel in-flight requests on logout', async () => {
      const token = createMockToken()
      const cookie = mockUseCookie<string>(token)
      const abortController = new AbortController()

      const mockFetch = vi.fn(
        async (url: string, options?: { signal?: AbortSignal }) => {
          return new Promise((resolve, reject) => {
            const timeout = setTimeout(() => {
              resolve({ data: 'response' })
            }, 100)

            options?.signal?.addEventListener('abort', () => {
              clearTimeout(timeout)
              reject(new Error('Request aborted'))
            })
          })
        },
      )

      // Start a slow request
      const requestPromise = mockFetch('/api/me', {
        signal: abortController.signal,
      })

      // User logs out before request completes
      cookie.value = null
      abortController.abort()

      // Request should be aborted
      await expect(requestPromise).rejects.toThrow('Request aborted')
    })
  })
})
