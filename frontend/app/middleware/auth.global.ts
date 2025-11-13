export default defineNuxtRouteMiddleware(async (to) => {
  // Skip auth check for callback page
  if (to.path.includes('/callback')) {
    return
  }

  const { token, loginWithRedirect } = useAuth()

  if (!token.value) {
    await loginWithRedirect()
  }
})
