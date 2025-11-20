export default defineNuxtRouteMiddleware(async (to) => {
  if (!to.path.startsWith('/admin')) {
    return
  }

  const token = useCookie('token')
  const config = useRuntimeConfig()

  // If no token, redirect to login
  if (!token.value) {
    return navigateTo(
      `${config.public.loginUrl}?redirect=${to.path}`,
      {
        external: true,
      },
    )
  }

  // Check if we already have user data
  const me = useState<any>('me', () => null)

  // If we have user data, check roles
  if (me.value) {
    const isAdmin = me.value?.roles.some((role: any) => role.role === 'ADMIN')
    const isSuperAdmin = me.value?.roles.some(
      (role: any) => role.role === 'SUPERADMIN',
    )

    if (!isSuperAdmin && !isAdmin) {
      return createError({
        statusCode: 403,
        statusMessage: 'Forbidden',
        message: 'You do not have permission to access this page',
      })
    }
  }

  // If we don't have user data yet, let it through and the page will handle loading
  // The useAuth() composable will be called in the layout/page and will populate the data
})
