import { until } from '@vueuse/core'
import {
  CHURCH_ADMIN_HOME,
  checkAdminPermission,
  isChurchAdminOnly,
} from '~/utils/adminPermissions'

/**
 * The single admin route guard.
 *
 * Replaces `middleware/admin.ts`, `middleware/superadmin.ts` and the
 * `routePermissions` watcher that used to live in `layouts/admin.vue` — three
 * copies of the same rules that disagreed with each other.
 *
 * Global rather than named so that no page under `/admin` can be left
 * unguarded by forgetting to declare it — previously every page had to remember
 * to list one, and several disagreed with both the nav and the server about
 * what they required.
 *
 * Pages opt into something stricter with `definePageMeta({ permission })`;
 * with none declared the default is `admin`, i.e. any admin-panel role.
 */
export default defineNuxtRouteMiddleware(async (to) => {
  if (!to.path.startsWith('/admin')) return

  // Every composable must be resolved before the first `await`: Nuxt drops the
  // injection context across one, and a `use*` call after it throws
  // "must be called within a reactive context". The values below are refs and
  // computeds, so they still observe anything that changes while we wait.
  const { me, token } = useAuth()
  const permissions = usePermissions()

  // A cold deep link arrives before the user query resolves, so wait for the
  // user rather than waving the navigation through — the old middleware
  // returned early on a missing `me` and left the decision to a layout watcher
  // that no longer exists.
  //
  // Waiting on the loading flags is not enough: `useAuth` pauses the `me` query
  // until a token is present and runs `watch(fetching, …, { immediate: true })`,
  // so `isLoading` is already false on the first tick while `me` is still null.
  // Keying off `me` itself is the only condition that means what it says.
  //
  // `!token.value` terminates the wait for the unauthenticated case. It is
  // reliable because `01.auth.global.ts` runs first and has already cleared an
  // expired token and redirected to /login by this point.
  if (token.value && !me.value) {
    await until(() => !!me.value || !token.value).toBe(true, {
      timeout: 10_000,
    })
  }

  if (!me.value || !permissions.canAccessAdmin.value) {
    return navigateTo('/')
  }

  if (
    isChurchAdminOnly(me.value.roles) &&
    !to.path.startsWith(CHURCH_ADMIN_HOME)
  ) {
    return navigateTo(CHURCH_ADMIN_HOME)
  }

  const required = to.meta.permission ?? 'admin'
  // `to.params` is a union across every route under typedPages, so the key has
  // to be probed rather than read directly.
  const projectId =
    'projectId' in to.params && typeof to.params.projectId === 'string'
      ? to.params.projectId
      : undefined

  if (!checkAdminPermission(required, permissions, { projectId })) {
    return navigateTo('/admin')
  }
})
