import { describe, it, expect, vi, beforeEach, afterEach } from 'vitest'
import { createMockUser } from '../utils/auth-mocks'
import { RoleType, type GetMeQuery } from '../../app/api/generated'

/**
 * Admin route middleware tests.
 *
 * These import and run the REAL middleware from `app/middleware/`. An earlier
 * version of these tests (in `auth.test.ts`) hand-copied the middleware logic
 * into each test body and asserted against the copy, so the suite stayed green
 * no matter what the middleware actually did — and the copy encoded the wrong
 * rule (it rejected everyone but Admin/Superadmin, while the real middleware
 * also admits ProjectAdmin and ChurchAdmin).
 *
 * The middleware relies entirely on Nuxt auto-imports, so the globals are
 * stubbed before the module is dynamically imported.
 */

type Me = GetMeQuery['me'] | null

const navigateTo = vi.fn((to: string) => ({ __redirect: to }))
const createError = vi.fn((opts: unknown) => ({ __error: opts }))

/** Current value of the `me` state, swapped per test. */
let meValue: Me = null

function stubNuxtGlobals() {
  vi.stubGlobal('defineNuxtRouteMiddleware', (fn: unknown) => fn)
  vi.stubGlobal('useState', () => ({
    get value() {
      return meValue
    },
  }))
  vi.stubGlobal('navigateTo', navigateTo)
  vi.stubGlobal('createError', createError)
  vi.stubGlobal('RoleType', RoleType)
}

/**
 * Load a middleware fresh, with globals in place before module evaluation.
 * The import paths are spelled out rather than interpolated so Vite can
 * statically analyse them.
 */
async function loadMiddleware(name: 'admin' | 'superadmin') {
  stubNuxtGlobals()
  vi.resetModules()
  const mod =
    name === 'admin'
      ? await import('../../app/middleware/admin')
      : await import('../../app/middleware/superadmin')
  return mod.default as (to: { path: string }) => Promise<unknown> | unknown
}

function userWithRoles(...roles: RoleType[]): Me {
  return createMockUser({ roles: roles.map((role) => ({ role, scope: null })) })
}

beforeEach(() => {
  meValue = null
  navigateTo.mockClear()
  createError.mockClear()
})

afterEach(() => {
  vi.unstubAllGlobals()
})

describe('middleware/admin', () => {
  it('ignores routes outside /admin', async () => {
    const middleware = await loadMiddleware('admin')
    meValue = userWithRoles(RoleType.User)

    await middleware({ path: '/challenges' })

    expect(createError).not.toHaveBeenCalled()
    expect(navigateTo).not.toHaveBeenCalled()
  })

  it('lets the request through when user data has not loaded yet', async () => {
    const middleware = await loadMiddleware('admin')
    meValue = null

    await middleware({ path: '/admin/projects' })

    // The page renders and the layout re-checks once auth resolves.
    expect(createError).not.toHaveBeenCalled()
    expect(navigateTo).not.toHaveBeenCalled()
  })

  // This is the case the old hand-copied tests got wrong: the real middleware
  // admits four roles, not two.
  it.each([
    [RoleType.Superadmin],
    [RoleType.Admin],
    [RoleType.ProjectAdmin],
    [RoleType.ChurchAdmin],
  ])('admits %s to the admin panel', async (role) => {
    const middleware = await loadMiddleware('admin')
    meValue = userWithRoles(role)

    // ChurchAdmin is confined to /admin/my-church, so probe it on its own path.
    const path =
      role === RoleType.ChurchAdmin ? '/admin/my-church' : '/admin/projects'
    await middleware({ path })

    expect(createError).not.toHaveBeenCalled()
    expect(navigateTo).not.toHaveBeenCalled()
  })

  it('rejects a plain user with 403', async () => {
    const middleware = await loadMiddleware('admin')
    meValue = userWithRoles(RoleType.User)

    await middleware({ path: '/admin/projects' })

    expect(createError).toHaveBeenCalledWith(
      expect.objectContaining({ statusCode: 403 }),
    )
    expect(navigateTo).not.toHaveBeenCalled()
  })

  describe('church-admin confinement', () => {
    it('redirects a church-admin-only user away from general admin routes', async () => {
      const middleware = await loadMiddleware('admin')
      meValue = userWithRoles(RoleType.ChurchAdmin)

      await middleware({ path: '/admin/projects' })

      expect(navigateTo).toHaveBeenCalledWith('/admin/my-church')
      expect(createError).not.toHaveBeenCalled()
    })

    it('leaves a church-admin-only user alone inside /admin/my-church', async () => {
      const middleware = await loadMiddleware('admin')
      meValue = userWithRoles(RoleType.ChurchAdmin)

      await middleware({ path: '/admin/my-church/units' })

      expect(navigateTo).not.toHaveBeenCalled()
      expect(createError).not.toHaveBeenCalled()
    })

    it('does not confine a church admin who also holds a full admin role', async () => {
      const middleware = await loadMiddleware('admin')
      meValue = userWithRoles(RoleType.ChurchAdmin, RoleType.Admin)

      await middleware({ path: '/admin/projects' })

      expect(navigateTo).not.toHaveBeenCalled()
      expect(createError).not.toHaveBeenCalled()
    })

    it('does not confine a project admin', async () => {
      const middleware = await loadMiddleware('admin')
      meValue = userWithRoles(RoleType.ProjectAdmin)

      await middleware({ path: '/admin/projects' })

      expect(navigateTo).not.toHaveBeenCalled()
      expect(createError).not.toHaveBeenCalled()
    })
  })
})

describe('middleware/superadmin', () => {
  it('lets the request through when user data has not loaded yet', async () => {
    const middleware = await loadMiddleware('superadmin')
    meValue = null

    await middleware({ path: '/admin/users' })

    expect(navigateTo).not.toHaveBeenCalled()
  })

  it('admits a superadmin', async () => {
    const middleware = await loadMiddleware('superadmin')
    meValue = userWithRoles(RoleType.Superadmin)

    await middleware({ path: '/admin/users' })

    expect(navigateTo).not.toHaveBeenCalled()
  })

  it.each([[RoleType.Admin], [RoleType.ProjectAdmin], [RoleType.ChurchAdmin]])(
    'redirects %s to /admin',
    async (role) => {
      const middleware = await loadMiddleware('superadmin')
      meValue = userWithRoles(role)

      await middleware({ path: '/admin/users' })

      expect(navigateTo).toHaveBeenCalledWith('/admin')
    },
  )

  // Documents a live inconsistency rather than asserting it is correct:
  // `canAccessTeams` in usePermissions is `isSuperAdmin || isProjectAdmin`, but
  // the teams pages guard with this middleware. A project admin therefore sees
  // the "Lag" nav entry and is bounced when they click it. See the restructure
  // plan — the permission matrix is reconciled before teams move under a project.
  it('bounces a project admin from the teams pages despite canAccessTeams allowing them', async () => {
    const middleware = await loadMiddleware('superadmin')
    meValue = userWithRoles(RoleType.ProjectAdmin)

    await middleware({ path: '/admin/teams' })

    expect(navigateTo).toHaveBeenCalledWith('/admin')
  })
})
