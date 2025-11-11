export default defineNuxtRouteMiddleware(async (to) => {
  if (!to.path.startsWith('/admin')) {
    return
  }

  const { me, isLoading, token, loginWithRedirect, isAdmin, isSuperAdmin } =
    useAuth()

  // If no token, redirect to login
  if (!token.value) {
    return loginWithRedirect()
  }

  // Wait for auth to complete (with timeout)
  let attempts = 0
  while (isLoading.value && attempts < 100) {
    // Max 1 second
    await new Promise((resolve) => setTimeout(resolve, 10))
    attempts++
  }

  // If still loading or no user data, something went wrong
  if (!me.value) {
    return loginWithRedirect()
  }

  // Check for superadmin role
  if (!isSuperAdmin.value && !isAdmin.value) {
    return createError({
      statusCode: 403,
      statusMessage: 'Forbidden',
      message: 'You do not have permission to access this page',
    })
  }
})
