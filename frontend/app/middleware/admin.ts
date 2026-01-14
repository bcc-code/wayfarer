export default defineNuxtRouteMiddleware(async (to) => {
  if (!to.path.startsWith('/admin')) {
    return
  }

  // Check if we already have user data
  const me = useState<any>('me', () => null)

  // If user data not loaded yet, let page render - layout will handle auth check after loading
  if (!me.value) {
    return
  }

  // Check roles
  const hasAdminRole = me.value?.roles.some((role: RoleLike) =>
    [
      RoleType.Admin,
      RoleType.Superadmin,
      RoleType.ProjectAdmin,
      RoleType.ChurchAdmin,
    ].includes(role.role),
  )

  if (!hasAdminRole) {
    return createError({
      statusCode: 403,
      statusMessage: 'Forbidden',
      message: 'You do not have permission to access this page',
    })
  }

  // Check if user has full admin access or is church-admin-only
  const hasFullAdminRole = me.value?.roles.some((role: RoleLike) =>
    [RoleType.Admin, RoleType.Superadmin].includes(role.role),
  )

  const isChurchAdminOnly =
    !hasFullAdminRole &&
    me.value?.roles.some((role: RoleLike) => role.role === RoleType.ChurchAdmin)

  // Restrict church-admin-only users to /admin/my-church routes
  if (isChurchAdminOnly && !to.path.startsWith('/admin/my-church')) {
    return navigateTo('/admin/my-church')
  }
})
