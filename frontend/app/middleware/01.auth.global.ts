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

  // Check if user has a valid Wayfarer token (already exchanged)
  const wayfarerToken = useLocalStorage<string>(ACCESS_TOKEN_KEY, () => null)
  // Redirect loop guard - prevent infinite redirects to callback
  const redirectAttempts = useLocalStorage('auth_redirect_attempts', 0)
  const lastRedirectTime = useLocalStorage('auth_redirect_time', 0)

  // A Wayfarer session renews itself without Auth0: refresh a missing or
  // expired access token from the stored refresh token.
  if (
    (!wayfarerToken.value || isTokenExpired(wayfarerToken.value)) &&
    hasRefreshToken()
  ) {
    await refreshSession(authBaseUrl(useRuntimeConfig().public))
    wayfarerToken.value = getStoredAccessToken()
  }

  // Valid token: no need to wait for Auth0 at all.
  if (wayfarerToken.value && !isTokenExpired(wayfarerToken.value)) {
    redirectAttempts.value = 0
    lastRedirectTime.value = 0
    return
  }

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

  // An expired token is as good as no token — clear it so we re-exchange
  // (Auth0 session alive) or land on /login, instead of rendering the page
  // and letting every query 401.
  if (wayfarerToken.value && isTokenExpired(wayfarerToken.value)) {
    wayfarerToken.value = null
  }

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
      // This can happen on page refresh - redirect to callback to exchange token.
      // The callback has no Auth0 params to read a target from on this path, so
      // the page being asked for travels with it: without that, an admin whose
      // token had expired was dropped on the user app's front page.
      return navigateTo(
        {
          path: '/auth0-callback',
          query: { redirect: safeRedirectPath(to.fullPath) },
        },
        { replace: true },
      )
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
  }
})
