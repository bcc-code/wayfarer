export default defineNuxtRouteMiddleware(async (to) => {
  // Skip auth check for callback page
  if (to.path === '/callback') {
    return
  }

  const token = useCookie('token')
  const config = useRuntimeConfig()

  if (!token.value) {
    return navigateTo(
      `${config.public.loginUrl}?redirect=${to.path}`,
      {
        external: true,
      },
    )
  }
})
