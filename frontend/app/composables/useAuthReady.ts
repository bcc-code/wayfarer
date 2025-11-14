/**
 * Composable that provides a reactive flag indicating whether auth is ready for GraphQL queries.
 * Use this to pause queries until we have a valid token and are not on the callback page.
 */
export function useAuthReady() {
  const token = useCookie('token')
  const route = useRoute()

  const isAuthReady = computed(() => {
    return !!token.value && route.path !== '/callback'
  })

  return { isAuthReady }
}
