export default defineNuxtRouteMiddleware(async (to) => {
  // Skip auth check for callback page
  if (to.path === '/callback') {
    return
  }

  const token = useLocalStorage<string>('token', () => null)
  const config = useRuntimeConfig()

  if (!token.value) {
    return navigateTo(`${config.public.loginUrl}?redirect=${to.path}`, {
      external: true,
    })
  }
})
