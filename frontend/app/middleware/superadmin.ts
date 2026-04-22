/**
 * Middleware that restricts access to superadmins only.
 * Use this for sensitive pages like role management.
 *
 * Note: When user data hasn't loaded yet (direct URL navigation),
 * the admin layout watcher handles the redirect after auth completes.
 *
 * Usage in page:
 * ```ts
 * definePageMeta({
 *   middleware: 'superadmin',
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
      return navigateTo('/admin')
    }
  }

  // If we don't have user data yet, let it through
  // The admin layout watcher handles the redirect after auth loads
})
