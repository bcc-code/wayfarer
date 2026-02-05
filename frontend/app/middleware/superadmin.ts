/**
 * Middleware that restricts access to superadmins only.
 * Use this for sensitive pages like role management.
 *
 * Usage in page:
 * ```ts
 * definePageMeta({
 *   middleware: ['admin', 'superadmin'],
 * })
 * ```
 */
export default defineNuxtRouteMiddleware(async () => {
  const me = useState<any>('me', () => null)

  // If we have user data, check if superadmin
  if (me.value) {
    const isSuperAdmin = me.value?.roles.some(
      (role: any) => role.role === 'SUPERADMIN',
    )

    if (!isSuperAdmin) {
      throw createError({
        statusCode: 403,
        statusMessage: 'Forbidden',
        message: 'This page requires superadmin permissions',
      })
    }
  }

  // If we don't have user data yet, let it through
  // The admin middleware runs first and handles auth,
  // then the page will check permissions after data loads
})
