import { useAuth0 } from '@auth0/auth0-vue'

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

  // Wait for Auth0 to initialize
  while (auth0.isLoading.value) {
    await new Promise((resolve) => setTimeout(resolve, 10))
  }

  // Check if user has a valid Wayfarer token (already exchanged)
  const wayfarerToken = useLocalStorage<string>('token', () => null)

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
      return navigateTo('/login', { replace: true })
    }
  } else {
    // Have valid token - clear redirect attempts
    redirectAttempts.value = 0
    lastRedirectTime.value = 0
  }
})
