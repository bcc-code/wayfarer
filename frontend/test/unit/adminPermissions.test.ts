import { describe, it, expect } from 'vitest'
import { computed } from 'vue'
import { RoleType } from '../../app/api/generated'
import {
  checkAdminPermission,
  isChurchAdminOnly,
  type AdminPermission,
} from '../../app/utils/adminPermissions'

/**
 * The admin permission matrix, pinned per role.
 *
 * These expectations are reconciled against the server's `@requireRole`
 * directives in `gql/*.graphqls`, which are the real boundary. The most
 * surprising entry is `project_admin`: it appears in exactly one directive
 * (`updateProject`), so a project admin may edit a project they own and nothing
 * else. Anywhere the UI granted them more, the server was returning 403.
 *
 * If a row here changes, the server rule should have changed first.
 */

type Perms = Parameters<typeof checkAdminPermission>[1]

type Role = 'superadmin' | 'admin' | 'projectAdmin' | 'churchAdmin' | 'user'

/**
 * Rebuilds the `canAccess*` flags the real `usePermissions` derives, from a
 * single role. Kept deliberately small: the point is to exercise
 * `checkAdminPermission`'s mapping, not to re-test usePermissions.
 */
function permissionsFor(role: Role, ownsProject = true): Perms {
  const isSuperAdmin = role === 'superadmin'
  const isAdmin = role === 'admin'
  const isProjectAdmin = role === 'projectAdmin'
  const isChurchAdmin = role === 'churchAdmin'
  const full = isSuperAdmin || isAdmin

  return {
    canAccessAdmin: computed(() => full || isProjectAdmin || isChurchAdmin),
    canAccessProjects: computed(() => full || isProjectAdmin),
    canAccessTeams: computed(() => full),
    canAccessUsers: computed(() => isSuperAdmin),
    canAccessScores: computed(() => full),
    canAccessConsents: computed(() => isSuperAdmin),
    canAccessFeedback: computed(() => full),
    canAccessMaintenance: computed(() => isSuperAdmin),
    canManageChurchAdmins: computed(() => full || isChurchAdmin),
    canEditProject: () => full || (isProjectAdmin && ownsProject),
  } as unknown as Perms
}

const PERMISSIONS: AdminPermission[] = [
  'admin',
  'projects:view',
  'project:edit',
  'teams:view',
  'users:view',
  'scores:view',
  'consents:view',
  'feedback:view',
  'maintenance:view',
  'church:manage',
]

/** Expected allow-list per role. Anything not listed must be denied. */
const ALLOWED: Record<Role, AdminPermission[]> = {
  superadmin: [...PERMISSIONS],
  admin: [
    'admin',
    'projects:view',
    'project:edit',
    'teams:view',
    'scores:view',
    'feedback:view',
    'church:manage',
  ],
  // The whole of a project admin's reach: get into the panel, see projects,
  // edit one they own.
  projectAdmin: ['admin', 'projects:view', 'project:edit'],
  churchAdmin: ['admin', 'church:manage'],
  user: [],
}

describe('admin permission matrix', () => {
  for (const role of Object.keys(ALLOWED) as Role[]) {
    describe(role, () => {
      for (const permission of PERMISSIONS) {
        const expected = ALLOWED[role].includes(permission)

        it(`${expected ? 'allows' : 'denies'} ${permission}`, () => {
          const actual = checkAdminPermission(
            permission,
            permissionsFor(role),
            { projectId: 'PR01' },
          )
          expect(actual).toBe(expected)
        })
      }
    })
  }

  describe('project:edit is scoped to the project in the route', () => {
    it('denies a project admin on a project they do not administer', () => {
      const perms = permissionsFor('projectAdmin', false)
      expect(
        checkAdminPermission('project:edit', perms, { projectId: 'PR99' }),
      ).toBe(false)
    })

    it('allows a full admin regardless of ownership', () => {
      const perms = permissionsFor('admin', false)
      expect(
        checkAdminPermission('project:edit', perms, { projectId: 'PR99' }),
      ).toBe(true)
    })

    // Guards against a project route losing its param and silently granting
    // edit rights to anyone who passes the role check.
    it('denies when the route carries no projectId', () => {
      expect(
        checkAdminPermission('project:edit', permissionsFor('superadmin'), {}),
      ).toBe(false)
    })
  })
})

describe('isChurchAdminOnly', () => {
  const role = (r: RoleType) => ({ role: r })

  it('is true for a church admin with no other admin role', () => {
    expect(isChurchAdminOnly([role(RoleType.ChurchAdmin)])).toBe(true)
  })

  it.each([[RoleType.Admin], [RoleType.Superadmin]])(
    'is false when they also hold %s',
    (other) => {
      expect(isChurchAdminOnly([role(RoleType.ChurchAdmin), role(other)])).toBe(
        false,
      )
    },
  )

  // A project admin is not confined: they have projects to administer, and
  // confining them to my-church would lock them out of that work.
  it('is false for a project admin who is also a church admin', () => {
    expect(
      isChurchAdminOnly([
        role(RoleType.ChurchAdmin),
        role(RoleType.ProjectAdmin),
      ]),
    ).toBe(false)
  })

  it.each([[[] as { role: RoleType }[]], [undefined], [null]])(
    'is false for %s roles',
    (roles) => {
      expect(isChurchAdminOnly(roles)).toBe(false)
    },
  )
})
