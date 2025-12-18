import { RoleType } from '~/api/generated'

/**
 * Centralized permissions composable for the admin panel.
 * Use this to check permissions for UI elements and actions.
 *
 * Example usage:
 * ```vue
 * <script setup>
 * const { canManageScores, canAssignRoles } = usePermissions()
 * </script>
 * <template>
 *   <UButton v-if="canAssignRoles">Add Role</UButton>
 * </template>
 * ```
 */
export function usePermissions() {
  const { me, isSuperAdmin, isAdmin } = useAuth()

  // ============================================
  // Scoped Role Helpers
  // ============================================

  /**
   * Check if user is a project admin for a specific project
   */
  const hasProjectAdminFor = (projectId: string) => {
    return me.value?.roles.some(
      (role) =>
        role.role === RoleType.ProjectAdmin &&
        role.scope?.project?.id === projectId,
    )
  }

  /**
   * Check if user is a church admin for a specific church
   */
  const hasChurchAdminFor = (churchId: string) => {
    return me.value?.roles.some(
      (role) =>
        role.role === RoleType.ChurchAdmin &&
        role.scope?.church?.id === churchId,
    )
  }

  /**
   * Check if user is a team lead for a specific team
   */
  const hasTeamLeadFor = (teamId: string) => {
    return me.value?.roles.some(
      (role) =>
        role.role === RoleType.TeamLead && role.scope?.team?.id === teamId,
    )
  }

  // ============================================
  // Page Access Permissions
  // ============================================

  /**
   * Can access the admin panel at all
   */
  const canAccessAdmin = computed(() => {
    return isSuperAdmin.value || isAdmin.value
  })

  /**
   * Can access the scores/score journal page
   */
  const canAccessScores = computed(() => {
    return isSuperAdmin.value || isAdmin.value
  })

  /**
   * Can access the users list page
   */
  const canAccessUsers = computed(() => {
    return isSuperAdmin.value || isAdmin.value
  })

  /**
   * Can access the teams page
   */
  const canAccessTeams = computed(() => {
    return isSuperAdmin.value || isAdmin.value
  })

  /**
   * Can access the consents page
   */
  const canAccessConsents = computed(() => {
    return isSuperAdmin.value || isAdmin.value
  })

  // ============================================
  // Action Permissions
  // ============================================

  /**
   * Can create/delete score adjustments
   */
  const canManageScores = computed(() => {
    return isSuperAdmin.value || isAdmin.value
  })

  /**
   * Can delete score journal entries
   */
  const canDeleteScoreEntry = computed(() => {
    return isSuperAdmin.value || isAdmin.value
  })

  /**
   * Can assign or revoke roles (superadmin only)
   */
  const canAssignRoles = computed(() => {
    return isSuperAdmin.value
  })

  /**
   * Can create new projects
   */
  const canCreateProject = computed(() => {
    return isSuperAdmin.value || isAdmin.value
  })

  /**
   * Can delete projects
   */
  const canDeleteProject = computed(() => {
    return isSuperAdmin.value || isAdmin.value
  })

  /**
   * Can edit a specific project (includes project admins for that project)
   */
  const canEditProject = (projectId: string) => {
    return (
      isSuperAdmin.value || isAdmin.value || hasProjectAdminFor(projectId)
    )
  }

  /**
   * Can manage teams (create, delete, modify)
   */
  const canManageTeams = computed(() => {
    return isSuperAdmin.value || isAdmin.value
  })

  /**
   * Can manage a specific team (includes team leads for that team)
   */
  const canManageTeam = (teamId: string) => {
    return isSuperAdmin.value || isAdmin.value || hasTeamLeadFor(teamId)
  }

  /**
   * Can manage consents (create, update, delete)
   */
  const canManageConsents = computed(() => {
    return isSuperAdmin.value || isAdmin.value
  })

  return {
    // Scoped helpers
    hasProjectAdminFor,
    hasChurchAdminFor,
    hasTeamLeadFor,

    // Page access
    canAccessAdmin,
    canAccessScores,
    canAccessUsers,
    canAccessTeams,
    canAccessConsents,

    // Actions
    canManageScores,
    canDeleteScoreEntry,
    canAssignRoles,
    canCreateProject,
    canDeleteProject,
    canEditProject,
    canManageTeams,
    canManageTeam,
    canManageConsents,
  }
}
