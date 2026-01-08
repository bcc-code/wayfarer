import type { RoleType } from '~/api/generated'

/**
 * Pure functions for permission and role checking.
 * These functions are stateless and can be easily unit tested.
 */

export interface RoleLike {
  role: RoleType
  scope?: {
    project?: { id: string } | null
    church?: { id: string } | null
  } | null
}

/**
 * Get all project IDs where the user has project admin role
 */
export function getProjectAdminIds(roles: RoleLike[]): string[] {
  return roles
    .filter(
      (role): role is RoleLike & { scope: { project: { id: string } } } =>
        role.role === 'PROJECT_ADMIN' && !!role.scope?.project?.id,
    )
    .map((role) => role.scope.project.id)
}

/**
 * Get all church IDs where the user has church admin role
 */
export function getChurchAdminIds(roles: RoleLike[]): string[] {
  return roles
    .filter(
      (role): role is RoleLike & { scope: { church: { id: string } } } =>
        role.role === 'CHURCH_ADMIN' && !!role.scope?.church?.id,
    )
    .map((role) => role.scope.church.id)
}

/**
 * Check if user has project admin role for a specific project
 */
export function hasProjectAdminFor(
  roles: RoleLike[],
  projectId: string,
): boolean {
  return roles.some(
    (role) =>
      role.role === 'PROJECT_ADMIN' && role.scope?.project?.id === projectId,
  )
}

/**
 * Check if user has church admin role for a specific church
 */
export function hasChurchAdminFor(
  roles: RoleLike[],
  churchId: string,
): boolean {
  return roles.some(
    (role) =>
      role.role === 'CHURCH_ADMIN' && role.scope?.church?.id === churchId,
  )
}

/**
 * Check if user has a specific role type
 */
export function hasRoleType(roles: RoleLike[], roleType: RoleType): boolean {
  return roles.some((role) => role.role === roleType)
}

/**
 * Find a specific project admin role
 */
export function findProjectAdminRole(
  roles: RoleLike[],
  projectId: string,
): RoleLike | undefined {
  return roles.find(
    (role) =>
      role.role === 'PROJECT_ADMIN' && role.scope?.project?.id === projectId,
  )
}

/**
 * Find a specific church admin role
 */
export function findChurchAdminRole(
  roles: RoleLike[],
  churchId: string,
): RoleLike | undefined {
  return roles.find(
    (role) =>
      role.role === 'CHURCH_ADMIN' && role.scope?.church?.id === churchId,
  )
}
