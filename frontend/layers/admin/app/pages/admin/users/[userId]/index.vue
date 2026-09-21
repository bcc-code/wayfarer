<script setup lang="ts">
definePageMeta({
  permission: 'users:view',
  layout: 'admin',
})

gql(`
	query AdminUserPageCurrentProject {
		currentProject {
			id
			name
		}
	}
`)

gql(`
	query AdminUserPage($id: ID!, $projectId: ID!) {
		user(id: $id) {
			id
      personUuid
      createdAt
			name
			email
			membersId
			age
			image
			language
			churchLockedUntil
			points(projectId: $projectId)
			church {
				id
				name
			}
			teams {
				id
				name
				parentProject {
					id
					name
				}
			}
			roles {
				id
				role
				scope {
					id
					type
					# RoleScope resolves these itself, so the card can name the
					# scope instead of printing its ULID.
					church {
						id
						name
					}
					project {
						id
						name
					}
					team {
						id
						name
					}
				}
			}
			consentStatus {
				acceptedConsents {
					id
					action
					actionDate
					consent {
						id
						key
						title
						version
						managementType
					}
				}
				rejectedConsents {
					id
					action
					actionDate
					consent {
						id
						key
						title
						version
					}
				}
				pendingConsents {
					id
					key
					title
					version
				}
			}
		}
		adminScoreJournal(filter: { userId: $id }, last: 100) {
			totalCount
			edges {
				node {
					id
					points
					sourceType
					reason
					createdAt
					project {
						id
						name
					}
					awardedBy {
						id
						name
					}
				}
			}
		}
		# 10, matching the "Viser 10 av N" notice this panel renders. It asked for
		# 100 and rendered all of them unsliced, so the notice was simply untrue —
		# and 100 entries inline is a wall in a panel that has a "Vis alle" link.
		feedback(filter: { userId: $id }, first: 10) {
			totalCount
			edges {
				node {
					id
					message
					canContactMe
					userAgent
					platform
					screenWidth
					screenHeight
					appVersion
					createdAt
				}
			}
		}
	}
`)

const route = useRoute('admin-users-userId')

const { isAuthReady } = useAuthReady()

// First query to get current project ID
const { data: currentProjectData } = useAdminUserPageCurrentProjectQuery({
  pause: computed(() => !isAuthReady.value),
})

const currentProjectId = computed(
  () => currentProjectData.value?.currentProject.id,
)
/** Names the project the points panel is scoped to. */
const currentProjectName = computed(
  () => currentProjectData.value?.currentProject.name,
)

// Main query that depends on having the project ID
const {
  data,
  fetching,
  error,
  executeQuery: refetch,
} = useAdminUserPageQuery({
  variables: computed(() => ({
    id: route.params.userId,
    projectId: currentProjectId.value ?? '',
  })),
  pause: computed(() => !isAuthReady.value || !currentProjectId.value),
})

// Supplies the trailing breadcrumb crumb and the navbar title; everything above
// it is derived from the route. Must come after `data` exists — see the
// `flush: 'post'` note in useAdminPage, which makes this safe either way.
useAdminPage(() => data.value?.user.name)

const { canAssignRoles, canCheckAchievements } = usePermissions()

// Score journal helpers
const scoreEntries = computed(
  () => data.value?.adminScoreJournal.edges.map((edge) => edge.node) ?? [],
)

const scoreTotalCount = computed(
  () => data.value?.adminScoreJournal.totalCount ?? 0,
)

// Feedback helpers
const feedbackEntries = computed(
  () => data.value?.feedback.edges.map((edge) => edge.node) ?? [],
)

const feedbackTotalCount = computed(() => data.value?.feedback.totalCount ?? 0)
</script>

<template>
  <!--
    Capped at the same `max-w-6xl` as the home dashboard, for the same reason.
    Full panel width stretched every card to ~1640px while its content sat in
    the left third — and it left each role row's delete button orphaned about
    1500px from the label it deletes, so you had to track across empty space to
    see what you were removing.
  -->
  <div class="max-w-6xl">
    <AdminQueryState :fetching :error>
      <div v-if="data" class="space-y-10">
        <AdminUserIdentity
          :user="data.user"
          :can-manage="canAssignRoles"
          :can-check-achievements="canCheckAchievements"
          @changed="refetch({ requestPolicy: 'network-only' })"
        />

        <AdminUserTeams :teams="data.user.teams" />

        <AdminUserRoles
          :user-id="route.params.userId"
          :roles="data.user.roles"
          :can-manage="canAssignRoles"
          @changed="refetch({ requestPolicy: 'network-only' })"
        />

        <AdminUserConsents
          :user-id="route.params.userId"
          :consent-status="data.user.consentStatus"
          @changed="refetch({ requestPolicy: 'network-only' })"
        />

        <AdminUserFeedbackPanel
          :entries="feedbackEntries"
          :total-count="feedbackTotalCount"
        />

        <AdminUserScoreJournal
          :entries="scoreEntries"
          :total-count="scoreTotalCount"
          :points="data.user.points"
          :project-name="currentProjectName"
        />
      </div>
    </AdminQueryState>
  </div>
</template>
