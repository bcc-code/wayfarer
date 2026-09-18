import { RoleType } from '~/api/generated'
import type { usePermissions } from '~/composables/usePermissions'

type Permissions = ReturnType<typeof usePermissions>

/**
 * Page-level permissions for the admin panel.
 *
 * Pages declare one via `definePageMeta({ permission: '...' })` and the global
 * `admin-permission` middleware enforces it. This replaces three copies of the
 * same rules that disagreed with each other: named `admin`/`superadmin` route
 * middleware, a `routePermissions` watcher in `layouts/admin.vue`, and the
 * `canAccess*` flags the navigation reads.
 *
 * The matrix is reconciled against the server's `@requireRole` directives in
 * `gql/*.graphqls`, which are the real boundary — this guard is navigation
 * hygiene, not security. Worth knowing when reading it: `project_admin` is
 * accepted by exactly one operation (`updateProject`), so project admins can
 * edit a project they own and do nothing else.
 */
export type AdminPermission =
  /** Any admin-panel role at all. The default when a page declares none. */
  | 'admin'
  | 'projects:view'
  | 'project:edit'
  | 'teams:view'
  | 'users:view'
  | 'scores:view'
  | 'consents:view'
  | 'churches:view'
  | 'feedback:view'
  | 'maintenance:view'
  | 'church:manage'

export interface AdminPermissionContext {
  /** Present while the route is inside a project. */
  projectId?: string
}

/**
 * Whether the current user satisfies `permission`.
 *
 * Pure apart from reading the already-computed flags off `permissions`, so the
 * whole matrix is testable without mounting anything or touching a router.
 */
export function checkAdminPermission(
  permission: AdminPermission,
  permissions: Permissions,
  context: AdminPermissionContext = {},
): boolean {
  switch (permission) {
    case 'admin':
      return !!permissions.canAccessAdmin.value
    case 'projects:view':
      return !!permissions.canAccessProjects.value
    // The one thing a project admin can actually do, and only for their own
    // projects — `updateProject` accepts `project_admin`.
    case 'project:edit':
      return (
        !!context.projectId && permissions.canEditProject(context.projectId)
      )
    case 'teams:view':
      return !!permissions.canAccessTeams.value
    case 'users:view':
      return !!permissions.canAccessUsers.value
    case 'scores:view':
      return !!permissions.canAccessScores.value
    case 'consents:view':
      return !!permissions.canAccessConsents.value
    case 'churches:view':
      return !!permissions.canAccessChurches.value
    case 'feedback:view':
      return !!permissions.canAccessFeedback.value
    case 'maintenance:view':
      return !!permissions.canAccessMaintenance.value
    case 'church:manage':
      return !!permissions.canManageChurchAdmins.value
  }
}

/**
 * Church admins who hold no other admin role are confined to `/admin/my-church`.
 * Anything else under `/admin` would be a wall of things they cannot use.
 */
export const CHURCH_ADMIN_HOME = '/admin/my-church'

export function isChurchAdminOnly(
  roles: { role: RoleType }[] | undefined | null,
): boolean {
  if (!roles?.length) return false
  // ProjectAdmin counts as "another role" here. The previous rule exempted only
  // Admin and Superadmin, so someone who was both a project admin and a church
  // admin was confined to my-church and locked out of the projects they
  // administer.
  const hasOtherAdminRole = roles.some((r) =>
    [RoleType.Admin, RoleType.Superadmin, RoleType.ProjectAdmin].includes(
      r.role,
    ),
  )
  return (
    !hasOtherAdminRole && roles.some((r) => r.role === RoleType.ChurchAdmin)
  )
}
