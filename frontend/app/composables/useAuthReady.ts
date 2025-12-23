/**
 * Composable that provides a reactive flag indicating whether auth is ready for GraphQL queries.
 * Use this to pause queries until we have a valid token and are not on the callback page.
 */
export function useAuthReady(providedRoute?: { path: string }) {
  const token = useLocalStorage<string>('token', () => null)
  const currentRoute = providedRoute || useRoute()

  const isAuthReady = computed(() => {
    const isAuthPage =
      currentRoute.path === '/auth0-callback' ||
      currentRoute.path === '/logout-callback'
    return !!token.value && !isAuthPage
  })

  return { isAuthReady }
}
