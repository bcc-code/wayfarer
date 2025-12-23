import { useAuth0 } from '@auth0/auth0-vue'

export default defineNuxtRouteMiddleware(async (to) => {
  // Skip auth check for auth pages
  if (to.path === '/auth0-callback' || to.path === '/logout-callback') {
    return
  }

  const auth0 = useAuth0()

  // Wait for Auth0 to initialize
  while (auth0.isLoading.value) {
    await new Promise((resolve) => setTimeout(resolve, 10))
  }

  // Check if user has a valid Wayfarer token (already exchanged)
  const wayfarerToken = useLocalStorage<string>('token', () => null)

  if (!wayfarerToken.value) {
    // No Wayfarer token - check if authenticated with Auth0
    if (auth0.isAuthenticated.value) {
      // Authenticated with Auth0 but no Wayfarer token
      // This can happen on page refresh - redirect to callback to exchange token
      return navigateTo('/auth0-callback', { replace: true })
    } else {
      // Not authenticated - redirect to Auth0 login
      await auth0.loginWithRedirect({
        appState: {
          targetUrl: to.path,
        },
      })
      // Return false to halt navigation (we're redirecting)
      return false
    }
  }
})
