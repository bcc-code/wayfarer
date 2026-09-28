import { useAuth0 } from '@auth0/auth0-vue'
import { until } from '@vueuse/core'

export default defineNuxtRouteMiddleware(async (to) => {
  // Skip auth check for auth pages
  if (
    to.path === '/login' ||
    to.path === '/auth0-callback' ||
    to.path === '/logout-callback'
  ) {
    return
  }

  const auth0 = useAuth0()

  // Wait for Auth0 to initialize, but never indefinitely — if the silent-auth
  // iframe hangs (blocked cookies, offline resume), fall through to the token
  // checks instead of blocking navigation forever.
  try {
    await until(auth0.isLoading).toBe(false, {
      timeout: 10_000,
      throwOnTimeout: true,
    })
  } catch {
    // Auth0 init stalled; token checks below decide where to go
  }

  // Check if user has a valid Wayfarer token (already exchanged)
  const wayfarerToken = useLocalStorage<string>('token', () => null)

  // An expired token is as good as no token — clear it so we re-exchange
  // (Auth0 session alive) or land on /login, instead of rendering the page
  // and letting every query 401.
  if (wayfarerToken.value && isTokenExpired(wayfarerToken.value)) {
    wayfarerToken.value = null
  }

  // Redirect loop guard - prevent infinite redirects to callback
  const redirectAttempts = useLocalStorage('auth_redirect_attempts', 0)
  const lastRedirectTime = useLocalStorage('auth_redirect_time', 0)

  if (!wayfarerToken.value) {
    // No Wayfarer token - check if authenticated with Auth0
    if (auth0.isAuthenticated.value) {
      // Check for redirect loop
      const now = Date.now()
      if (now - lastRedirectTime.value < 5000) {
        redirectAttempts.value++
        if (redirectAttempts.value > 3) {
          // Too many redirects in short time - clear state and go to login
          redirectAttempts.value = 0
          lastRedirectTime.value = 0
          await auth0.logout({
            logoutParams: {
              returnTo: `${window.location.origin}/login`,
            },
          })
          return
        }
      } else {
        redirectAttempts.value = 1
      }
      lastRedirectTime.value = now

      // Authenticated with Auth0 but no Wayfarer token
      // This can happen on page refresh - redirect to callback to exchange token
      return navigateTo('/auth0-callback', { replace: true })
    } else {
      // Not authenticated - redirect to login page
      // Clear redirect attempts on successful login flow
      redirectAttempts.value = 0
      lastRedirectTime.value = 0
      const redirect = to.fullPath === '/' ? undefined : to.fullPath
      return navigateTo(
        { path: '/login', query: { redirect } },
        { replace: true },
      )
    }
  } else {
    // Have valid token - clear redirect attempts
    redirectAttempts.value = 0
    lastRedirectTime.value = 0
  }
})
