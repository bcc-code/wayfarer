<script setup lang="ts">
definePageMeta({
  permission: 'users:view',
  layout: 'admin',
})

gql(`
	query AdminUserPageCurrentProject {
		currentProject {
			id
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
			pointsByProject {
				projectId
				projectName
				points
			}
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
		# A "what just happened" window; the full log is on the project's own
		# score page, filtered to this user.
		adminScoreJournal(filter: { userId: $id }, last: 5) {
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
		# 10: the panel's own notice says "Viser 10", and it links to the rest.
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

// After `data` exists. `useAdminPage` defers the read, so order is not
// load-bearing — but reading it in order is clearer.
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
  <!-- Capped: full panel width left each row's delete button ~1500px from
       the label it deletes. -->
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
          :user-id="route.params.userId"
        />

        <AdminUserScoreJournal
          :points-by-project="data.user.pointsByProject"
          :recent="scoreEntries"
          :total-count="scoreTotalCount"
          :user-id="route.params.userId"
          :current-project-id="currentProjectId"
        />
      </div>
    </AdminQueryState>
  </div>
</template>
