import {
  getProjectAdminIds,
  getChurchAdminIds,
  hasProjectAdminFor as hasProjectAdminForPure,
  hasChurchAdminFor as hasChurchAdminForPure,
} from '~/utils/permissions'

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
  const { me, isSuperAdmin, isAdmin, isChurchAdmin, isProjectAdmin } = useAuth()

  // ============================================
  // Scoped Role Helpers
  // ============================================

  /**
   * Check if user is a project admin for a specific project
   */
  const hasProjectAdminFor = (projectId: string) => {
    return hasProjectAdminForPure(me.value?.roles ?? [], projectId)
  }

  /**
   * Check if user is a church admin for a specific church
   */
  const hasChurchAdminFor = (churchId: string) => {
    return hasChurchAdminForPure(me.value?.roles ?? [], churchId)
  }

  /**
   * Get the project IDs the user is a project admin for
   */
  const projectAdminProjectIds = computed(() => {
    return getProjectAdminIds(me.value?.roles ?? [])
  })

  /**
   * Get the church IDs the user is a church admin for
   */
  const churchAdminChurchIds = computed(() => {
    return getChurchAdminIds(me.value?.roles ?? [])
  })

  // ============================================
  // Page Access Permissions
  // ============================================

  /**
   * Can access the admin panel at all
   * - Superadmins and Admins have full access
   * - Project admins can access for their projects
   * - Church admins can access for their church members
   */
  const canAccessAdmin = computed(() => {
    return (
      isSuperAdmin.value ||
      isAdmin.value ||
      isProjectAdmin.value ||
      isChurchAdmin.value
    )
  })

  /**
   * Can access the projects list page
   * - Superadmins and Admins see all projects
   * - Project admins can see the projects they manage
   */
  const canAccessProjects = computed(() => {
    return isSuperAdmin.value || isAdmin.value || isProjectAdmin.value
  })

  /**
   * Can access the scores/score journal page
   */
  const canAccessScores = computed(() => {
    return isSuperAdmin.value || isAdmin.value
  })

  /**
   * Can access the users list page
   * - Superadmins and Admins see all users
   * - Church admins can see users from their church
   */
  const canAccessUsers = computed(() => {
    return isSuperAdmin.value || isAdmin.value || isChurchAdmin.value
  })

  /**
   * Can access the teams page
   * - Superadmins and Admins see all teams
   * - Project admins can see teams in their projects
   */
  const canAccessTeams = computed(() => {
    return isSuperAdmin.value || isAdmin.value || isProjectAdmin.value
  })

  /**
   * Can access the consents page
   */
  const canAccessConsents = computed(() => {
    return isSuperAdmin.value || isAdmin.value
  })

  /**
   * Can access the feedback page
   */
  const canAccessFeedback = computed(() => {
    return isSuperAdmin.value || isAdmin.value
  })

  /**
   * Can delete feedback entries
   */
  const canDeleteFeedback = computed(() => {
    return isSuperAdmin.value || isAdmin.value
  })

  // ============================================
  // Action Permissions
  // ============================================

  /**
   * Can create/delete score adjustments
   * - Project admins can manage scores for their projects
   */
  const canManageScores = computed(() => {
    return isSuperAdmin.value || isAdmin.value || isProjectAdmin.value
  })

  /**
   * Can manage scores for a specific project
   */
  const canManageScoresFor = (projectId: string) => {
    return isSuperAdmin.value || isAdmin.value || hasProjectAdminFor(projectId)
  }

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
    return isSuperAdmin.value || isAdmin.value || hasProjectAdminFor(projectId)
  }

  /**
   * Can view a specific project
   */
  const canViewProject = (projectId: string) => {
    return isSuperAdmin.value || isAdmin.value || hasProjectAdminFor(projectId)
  }

  /**
   * Can manage teams (create, delete, modify)
   */
  const canManageTeams = computed(() => {
    return isSuperAdmin.value || isAdmin.value
  })

  /**
   * Can create teams for a project
   */
  const canCreateTeamFor = (projectId: string) => {
    return isSuperAdmin.value || isAdmin.value || hasProjectAdminFor(projectId)
  }

  /**
   * Can manage a specific team
   */
  const canManageTeam = (_teamId?: string) => {
    // TODO: Add project-based permission check when team's project is available
    return isSuperAdmin.value || isAdmin.value
  }

  /**
   * Can view a specific user
   * - Church admins can view users from their church
   */
  const canViewUser = (userChurchId?: string) => {
    if (isSuperAdmin.value || isAdmin.value) return true
    if (isChurchAdmin.value && userChurchId) {
      return hasChurchAdminFor(userChurchId)
    }
    return false
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

    // Scoped IDs (useful for filtering)
    projectAdminProjectIds,
    churchAdminChurchIds,

    // Page access
    canAccessAdmin,
    canAccessProjects,
    canAccessScores,
    canAccessUsers,
    canAccessTeams,
    canAccessConsents,
    canAccessFeedback,
    canDeleteFeedback,

    // Actions
    canManageScores,
    canManageScoresFor,
    canDeleteScoreEntry,
    canAssignRoles,
    canCreateProject,
    canDeleteProject,
    canEditProject,
    canViewProject,
    canManageTeams,
    canCreateTeamFor,
    canManageTeam,
    canViewUser,
    canManageConsents,
  }
}
