import { describe, it, expect, vi, beforeEach, afterEach } from 'vitest'
import { computed, ref, type Ref } from 'vue'
import { RoleType, type GetMeQuery } from '../../app/api/generated'
import { usePermissions } from '../../app/composables/usePermissions'
import { createMockUser } from '../utils/auth-mocks'

/**
 * The admin route guard, end to end.
 *
 * Runs the REAL `middleware/admin-permission.global.ts` against the REAL
 * `usePermissions`, so a change to either the permission matrix or the gates it
 * reads will fail here. Only `useAuth` is stubbed — there is no Auth0 or user
 * query in a unit test.
 *
 * This replaces tests that hand-copied the middleware logic into each case and
 * asserted against the copy, which could not fail when the middleware changed
 * and encoded a rule the app never implemented.
 */

type Me = GetMeQuery['me'] | null

const navigateTo = vi.fn((to: string) => ({ __redirect: to }))

let me: Ref<Me>
let token: Ref<string | null>
let isLoading: Ref<boolean>
let isAuth0Loading: Ref<boolean>

function hasRole(...roles: RoleType[]) {
  return computed(
    () =>
      !!me.value?.roles.some((r: { role: RoleType }) => roles.includes(r.role)),
  )
}

/**
 * Emulates Nuxt's injection context. Nuxt drops it across an `await`, so a
 * `use*` call after one throws at runtime. Flipping this to false once the
 * middleware has returned its promise reproduces that exactly: everything up to
 * the first `await` ran synchronously, anything after it did not.
 */
let contextActive = true

function requireContext() {
  if (!contextActive) {
    throw new Error(
      'use* function must be called within a reactive context (component setup, composable, or effect scope).',
    )
  }
}

function stubGlobals() {
  vi.stubGlobal('defineNuxtRouteMiddleware', (fn: unknown) => fn)
  vi.stubGlobal('navigateTo', navigateTo)
  vi.stubGlobal('computed', computed)
  vi.stubGlobal('useAuth', () => ({
    ...(requireContext() as undefined),
    me,
    token,
    isLoading,
    isAuth0Loading,
    isSuperAdmin: hasRole(RoleType.Superadmin),
    isAdmin: hasRole(RoleType.Admin),
    isChurchAdmin: hasRole(RoleType.ChurchAdmin),
    isProjectAdmin: hasRole(RoleType.ProjectAdmin),
  }))
  // The real implementation, so the matrix is exercised rather than mocked.
  vi.stubGlobal('usePermissions', () => {
    requireContext()
    return usePermissions()
  })
}

async function loadMiddleware() {
  stubGlobals()
  vi.resetModules()
  const mod = await import('../../app/middleware/02.admin-permission.global')
  return mod.default as (to: {
    path: string
    meta?: Record<string, unknown>
    params?: Record<string, string>
  }) => Promise<unknown>
}

function signIn(...roles: RoleType[]) {
  me.value = createMockUser({
    roles: roles.map((role) => ({
      role,
      // Project admins are scoped to a project; usePermissions reads the scope.
      scope:
        role === RoleType.ProjectAdmin ? { project: { id: 'PR01' } } : null,
    })),
  }) as Me
}

/** Shorthand for a route: path, optional permission meta and params. */
function route(
  path: string,
  permission?: string,
  params: Record<string, string> = {},
) {
  return { path, meta: permission ? { permission } : {}, params }
}

beforeEach(() => {
  contextActive = true
  me = ref(null)
  token = ref('a-token')
  isLoading = ref(false)
  isAuth0Loading = ref(false)
  navigateTo.mockClear()
})

afterEach(() => {
  vi.unstubAllGlobals()
})

describe('middleware/admin-permission (global)', () => {
  it('ignores routes outside /admin', async () => {
    const middleware = await loadMiddleware()
    me.value = null

    await middleware(route('/challenges'))

    expect(navigateTo).not.toHaveBeenCalled()
  })

  it('sends a signed-out visitor to the front page', async () => {
    const middleware = await loadMiddleware()
    me.value = null
    token.value = null

    await middleware(route('/admin'))

    expect(navigateTo).toHaveBeenCalledWith('/')
  })

  it('sends a user with no admin role to the front page', async () => {
    const middleware = await loadMiddleware()
    signIn(RoleType.User)

    await middleware(route('/admin'))

    expect(navigateTo).toHaveBeenCalledWith('/')
  })

  it('waits for the user query before deciding on a cold deep link', async () => {
    const middleware = await loadMiddleware()
    // Token present, user not resolved yet — the old middleware waved this
    // through and left the decision to a layout watcher that no longer exists.
    token.value = 'a-token'
    me.value = null
    isLoading.value = true

    const navigation = middleware(route('/admin/users', 'users:view'))
    signIn(RoleType.Superadmin)
    isLoading.value = false
    await navigation

    expect(navigateTo).not.toHaveBeenCalled()
  })

  // Regression: `useAuth` flips `isLoading` to false immediately when the `me`
  // query is paused (`watch(fetching, …, { immediate: true })`), so waiting on
  // the loading flags returns at once with `me` still null and the guard
  // redirected a signed-in superadmin to the front page.
  it('waits for the user, not merely for the loading flags to clear', async () => {
    const middleware = await loadMiddleware()
    token.value = 'a-token'
    me.value = null
    // Already false, as it is on a real cold load.
    isLoading.value = false
    isAuth0Loading.value = false

    const navigation = middleware(route('/admin'))
    // The user query resolves after a real delay, as a network call does. A
    // microtask would let the assertion pass on timing luck.
    await new Promise((resolve) => setTimeout(resolve, 20))
    signIn(RoleType.Superadmin)
    await navigation

    expect(navigateTo).not.toHaveBeenCalled()
  })

  // Regression: `usePermissions()` sat after the `until()` await, so on a cold
  // deep link it ran without a reactive context and the page 500'd. Unit tests
  // with plain stubs cannot see that, hence the emulation above.
  it('resolves every composable before awaiting auth', async () => {
    const middleware = await loadMiddleware()
    token.value = 'a-token'
    me.value = null
    isLoading.value = true

    const navigation = middleware(route('/admin/users', 'users:view'))
    // Past this point Nuxt would have dropped the injection context.
    contextActive = false
    signIn(RoleType.Superadmin)
    isLoading.value = false

    await expect(navigation).resolves.not.toThrow()
  })

  describe('permission enforcement', () => {
    it('defaults to any admin role when a page declares none', async () => {
      const middleware = await loadMiddleware()
      signIn(RoleType.ProjectAdmin)

      await middleware(route('/admin'))

      expect(navigateTo).not.toHaveBeenCalled()
    })

    it('admits a superadmin to a superadmin-only page', async () => {
      const middleware = await loadMiddleware()
      signIn(RoleType.Superadmin)

      await middleware(route('/admin/maintenance', 'maintenance:view'))

      expect(navigateTo).not.toHaveBeenCalled()
    })

    it('bounces a plain admin off a superadmin-only page', async () => {
      const middleware = await loadMiddleware()
      signIn(RoleType.Admin)

      await middleware(route('/admin/maintenance', 'maintenance:view'))

      expect(navigateTo).toHaveBeenCalledWith('/admin')
    })

    // The bug this whole change exists to fix: the nav offered "Lag" to project
    // admins and the page's `superadmin` middleware bounced them.
    it('no longer offers project admins a teams page they cannot use', async () => {
      const middleware = await loadMiddleware()
      signIn(RoleType.ProjectAdmin)

      await middleware(route('/admin/teams', 'teams:view'))

      expect(navigateTo).toHaveBeenCalledWith('/admin')
    })

    it('bounces a project admin off the score journal', async () => {
      const middleware = await loadMiddleware()
      signIn(RoleType.ProjectAdmin)

      await middleware(route('/admin/scores', 'scores:view'))

      expect(navigateTo).toHaveBeenCalledWith('/admin')
    })
  })

  describe('project:edit is scoped to the project in the route', () => {
    it('admits a project admin to a project they administer', async () => {
      const middleware = await loadMiddleware()
      signIn(RoleType.ProjectAdmin)

      await middleware(
        route('/admin/projects/PR01/edit', 'project:edit', {
          projectId: 'PR01',
        }),
      )

      expect(navigateTo).not.toHaveBeenCalled()
    })

    it('bounces a project admin from someone else’s project', async () => {
      const middleware = await loadMiddleware()
      signIn(RoleType.ProjectAdmin)

      await middleware(
        route('/admin/projects/PR99/edit', 'project:edit', {
          projectId: 'PR99',
        }),
      )

      expect(navigateTo).toHaveBeenCalledWith('/admin')
    })
  })

  describe('church-admin confinement', () => {
    it('redirects a church-admin-only user away from general admin routes', async () => {
      const middleware = await loadMiddleware()
      signIn(RoleType.ChurchAdmin)

      await middleware(route('/admin/projects', 'projects:view'))

      expect(navigateTo).toHaveBeenCalledWith('/admin/my-church')
    })

    it('leaves them alone inside /admin/my-church', async () => {
      const middleware = await loadMiddleware()
      signIn(RoleType.ChurchAdmin)

      await middleware(route('/admin/my-church/units', 'church:manage'))

      expect(navigateTo).not.toHaveBeenCalled()
    })

    it('does not confine a church admin who is also a full admin', async () => {
      const middleware = await loadMiddleware()
      signIn(RoleType.ChurchAdmin, RoleType.Admin)

      await middleware(route('/admin/projects', 'projects:view'))

      expect(navigateTo).not.toHaveBeenCalled()
    })

    // Previously confined, which locked them out of the projects they run.
    it('does not confine a church admin who is also a project admin', async () => {
      const middleware = await loadMiddleware()
      signIn(RoleType.ChurchAdmin, RoleType.ProjectAdmin)

      await middleware(route('/admin/projects', 'projects:view'))

      expect(navigateTo).not.toHaveBeenCalled()
    })
  })
})
