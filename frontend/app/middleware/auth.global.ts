export default defineNuxtRouteMiddleware(async () => {
  const { token, loginWithRedirect } = useAuth()

  if (!token.value) {
    await loginWithRedirect()
  }
})
