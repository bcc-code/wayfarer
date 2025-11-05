export default defineNuxtRouteMiddleware(async (to) => {
  if (!to.path.startsWith('/admin')) {
    return
  }

  const { me, isLoading } = useAuth()

  if (isLoading.value) {
    await new Promise((resolve) => setTimeout(resolve, 10))
  }

  if (!me.value?.roles.some((role) => role.role === RoleType.Superadmin)) {
    return createError({
      statusCode: 403,
      statusMessage: 'Unauthorized',
      message: 'You do not have permission to access this page',
    })
  }
})
