/**
 * Composable that provides a reactive flag indicating whether auth is ready for GraphQL queries.
 * Use this to pause queries until we have a valid token and are not on the callback page.
 */
export function useAuthReady(providedRoute?: { path: string }) {
  const token = useLocalStorage<string>('token', () => null)
  // `useRouter().currentRoute` rather than `useRoute()`: this also runs inside
  // route middleware (via useAuth), where `useRoute()` warns that it returns
  // the route being navigated *away* from. That is the value we want either
  // way — the flag only pauses queries while the app sits on a callback page —
  // and the ref stays reactive, so the computed updates once navigation lands.
  const router = useRouter()

  const isAuthReady = computed(() => {
    const path = providedRoute?.path ?? router.currentRoute.value.path
    const isAuthPage = path === '/auth0-callback' || path === '/logout-callback'
    return !!token.value && !isAuthPage
  })

  return { isAuthReady }
}
