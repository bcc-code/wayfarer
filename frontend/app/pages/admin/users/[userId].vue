<script setup lang="ts">
import { RoleType, ScopeType } from '~/api/generated'

definePageMeta({
  layout: 'admin',
  middleware: 'admin',
})

gql(`
	query AdminUserPage($id: ID!) {
		user(id: $id) {
			id
			name
			email
			membersId
			gender
			birthdate
			age
			image
			church {
				id
				name
			}
			roles {
				id
				role
				scope {
					id
					type
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
		adminScoreJournal(filter: { userId: $id }, first: 20) {
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
const {
  data,
  fetching,
  error,
  executeQuery: refetch,
} = useAdminUserPageQuery({
  variables: {
    id: route.params.userId,
  },
  pause: computed(() => !isAuthReady.value),
})

const { executeMutation: assignRole } = useAssignRoleMutation()
const { executeMutation: revokeRole } = useRevokeRoleMutation()
const toast = useToast()

// Permissions
const { canAssignRoles } = usePermissions()

const roleOptions = [
  { label: 'User', value: RoleType.User },
  { label: 'Admin', value: RoleType.Admin },
  { label: 'Superadmin', value: RoleType.Superadmin },
  { label: 'Church Admin', value: RoleType.ChurchAdmin },
  { label: 'Project Admin', value: RoleType.ProjectAdmin },
  { label: 'Team Lead', value: RoleType.TeamLead },
  { label: 'M2M', value: RoleType.M2M },
]

const scopeTypeOptions = [
  { label: 'None (Global)', value: null },
  { label: 'Church', value: ScopeType.Church },
  { label: 'Project', value: ScopeType.Project },
  { label: 'Team', value: ScopeType.Team },
]

const showAddRoleModal = ref(false)
const newRole = reactive({
  role: RoleType.User as RoleType,
  scopeType: null as ScopeType | null,
  scopeId: '',
})

function resetNewRoleForm() {
  newRole.role = RoleType.User
  newRole.scopeType = null
  newRole.scopeId = ''
}

async function handleAssignRole() {
  const result = await assignRole({
    input: {
      userId: route.params.userId,
      role: newRole.role,
      scopeType: newRole.scopeType,
      scopeId:
        newRole.scopeType && newRole.scopeId ? newRole.scopeId : undefined,
    },
  })

  if (result.error) {
    toast.add({
      title: 'Failed to assign role',
      description: result.error.message,
      color: 'error',
    })
    return
  }

  toast.add({
    title: 'Role assigned',
    description: `Successfully assigned ${newRole.role} role`,
    color: 'success',
  })

  showAddRoleModal.value = false
  resetNewRoleForm()
  refetch({ requestPolicy: 'network-only' })
}

async function handleRevokeRole(
  roleId: string,
  role: RoleType,
  scopeType?: ScopeType | null,
  scopeId?: string | null,
) {
  const result = await revokeRole({
    input: {
      userId: route.params.userId,
      role,
      scopeType: scopeType ?? undefined,
      scopeId: scopeId ?? undefined,
    },
  })

  if (result.error) {
    toast.add({
      title: 'Failed to revoke role',
      description: result.error.message,
      color: 'error',
    })
    return
  }

  toast.add({
    title: 'Role revoked',
    description: `Successfully revoked ${role} role`,
    color: 'success',
  })

  refetch({ requestPolicy: 'network-only' })
}

// Score journal helpers
const scoreEntries = computed(
  () => data.value?.adminScoreJournal.edges.map((edge) => edge.node) ?? [],
)

const scoreTotalCount = computed(
  () => data.value?.adminScoreJournal.totalCount ?? 0,
)

function formatSourceType(type: string) {
  return type.charAt(0) + type.slice(1).toLowerCase()
}

function formatScoreDate(date: string) {
  return new Date(date).toLocaleDateString('en-US', {
    year: 'numeric',
    month: 'short',
    day: 'numeric',
  })
}

// Feedback helpers
const feedbackEntries = computed(
  () => data.value?.feedback.edges.map((edge) => edge.node) ?? [],
)

const feedbackTotalCount = computed(() => data.value?.feedback.totalCount ?? 0)

function formatFeedbackDate(date: string) {
  return new Date(date).toLocaleDateString('en-US', {
    year: 'numeric',
    month: 'short',
    day: 'numeric',
    hour: '2-digit',
    minute: '2-digit',
  })
}
</script>

<template>
  <div>
    <div class="border-default border-b py-2">
      <UContainer>
        <UBreadcrumb
          :items="[
            { label: 'Users', to: { name: 'admin-users' } },
            {
              label: data?.user.name ?? route.params.userId,
              to: {
                name: 'admin-users-userId',
                params: { userId: route.params.userId },
              },
            },
          ]"
        />
      </UContainer>
    </div>
    <UContainer class="py-12">
      <LoadingState v-if="fetching" />
      <ErrorState v-else-if="error" :error />
      <div v-else-if="data" class="space-y-6">
        <!-- User Header -->
        <div>
          <h1 class="text-3xl font-bold">{{ data.user.name }}</h1>
          <p class="text-dimmed text-lg">{{ data.user.email }}</p>
        </div>

        <!-- User Info -->
        <dl class="text-sm">
          <div class="flex gap-6 border-b border-default py-2">
            <dt class="text-muted w-24 shrink-0">Members ID</dt>
            <dd class="font-medium">{{ data.user.membersId }}</dd>
          </div>
          <div class="flex gap-6 border-b border-default py-2">
            <dt class="text-muted w-24 shrink-0">User ID</dt>
            <dd class="font-mono">{{ data.user.id }}</dd>
          </div>
          <div class="flex gap-6 border-b border-default py-2">
            <dt class="text-muted w-24 shrink-0">Gender</dt>
            <dd class="font-medium">{{ capitalizeFirst(data.user.gender) }}</dd>
          </div>
          <div class="flex gap-6 border-b border-default py-2">
            <dt class="text-muted w-24 shrink-0">Age</dt>
            <dd class="font-medium">{{ data.user.age }} years</dd>
          </div>
          <div class="flex gap-6 border-b border-default py-2">
            <dt class="text-muted w-24 shrink-0">Birthdate</dt>
            <dd class="font-medium">{{ formatDate(data.user.birthdate) }}</dd>
          </div>
          <div class="flex gap-6 border-b border-default py-2">
            <dt class="text-muted w-24 shrink-0">Church</dt>
            <dd class="font-medium">{{ data.user.church.name }}</dd>
          </div>
          <div class="flex gap-6 py-2">
            <dt class="text-muted w-24 shrink-0">Church ID</dt>
            <dd class="font-mono">{{ data.user.church.id }}</dd>
          </div>
        </dl>

        <!-- Roles Card -->
        <UCard>
          <template #header>
            <div class="flex items-center justify-between">
              <h2 class="text-xl font-semibold">Roles & Permissions</h2>
              <UButton
                v-if="canAssignRoles"
                icon="i-lucide-plus"
                size="sm"
                @click="showAddRoleModal = true"
              >
                Add Role
              </UButton>
            </div>
          </template>

          <div v-if="data.user.roles.length > 0" class="space-y-3">
            <div
              v-for="role in data.user.roles"
              :key="role.id"
              class="border-default flex items-center justify-between rounded-md border p-3"
            >
              <div class="flex items-center gap-3">
                <UBadge variant="soft" size="lg">
                  {{ role.role }}
                </UBadge>
                <div v-if="role.scope">
                  <span class="text-dimmed text-sm">Scope: </span>
                  <span class="text-sm font-medium">
                    {{ capitalizeFirst(role.scope.type) }}
                  </span>
                  <span class="text-dimmed ml-2 text-xs">
                    ({{ role.scope.id }})
                  </span>
                </div>
              </div>
              <UButton
                v-if="canAssignRoles"
                icon="i-lucide-trash-2"
                color="error"
                variant="ghost"
                size="sm"
                @click="
                  handleRevokeRole(
                    role.id,
                    role.role,
                    role.scope?.type,
                    role.scope?.id,
                  )
                "
              />
            </div>
          </div>
          <div v-else class="text-dimmed">No roles assigned</div>
        </UCard>

        <!-- Consents Card -->
        <UCard>
          <template #header>
            <h2 class="text-xl font-semibold">Consents</h2>
          </template>

          <div class="space-y-4">
            <!-- Pending Consents -->
            <div v-if="data.user.consentStatus.pendingConsents.length > 0">
              <h3 class="text-muted mb-2 text-sm font-medium">Pending</h3>
              <div class="space-y-2">
                <div
                  v-for="consent in data.user.consentStatus.pendingConsents"
                  :key="consent.id"
                  class="border-default flex items-center justify-between rounded-md border p-3"
                >
                  <div class="flex items-center gap-3">
                    <UBadge variant="soft" color="warning">Pending</UBadge>
                    <div>
                      <span class="font-medium">{{ consent.title }}</span>
                      <span class="text-dimmed ml-2 text-xs"
                        >v{{ consent.version }}</span
                      >
                    </div>
                  </div>
                  <code class="text-dimmed text-xs">{{ consent.key }}</code>
                </div>
              </div>
            </div>

            <!-- Accepted Consents -->
            <div v-if="data.user.consentStatus.acceptedConsents.length > 0">
              <h3 class="text-muted mb-2 text-sm font-medium">Accepted</h3>
              <div class="space-y-2">
                <div
                  v-for="item in data.user.consentStatus.acceptedConsents"
                  :key="item.id"
                  class="border-default flex items-center justify-between rounded-md border p-3"
                >
                  <div class="flex items-center gap-3">
                    <UBadge variant="soft" color="success">Accepted</UBadge>
                    <div>
                      <span class="font-medium">{{ item.consent.title }}</span>
                      <span class="text-dimmed ml-2 text-xs"
                        >v{{ item.consent.version }}</span
                      >
                    </div>
                  </div>
                  <div class="text-right">
                    <code class="text-dimmed text-xs">{{
                      item.consent.key
                    }}</code>
                    <div class="text-dimmed text-xs">
                      {{ formatDateTime(item.actionDate) }}
                    </div>
                  </div>
                </div>
              </div>
            </div>

            <!-- Rejected Consents -->
            <div v-if="data.user.consentStatus.rejectedConsents.length > 0">
              <h3 class="text-muted mb-2 text-sm font-medium">Rejected</h3>
              <div class="space-y-2">
                <div
                  v-for="item in data.user.consentStatus.rejectedConsents"
                  :key="item.id"
                  class="border-default flex items-center justify-between rounded-md border p-3"
                >
                  <div class="flex items-center gap-3">
                    <UBadge variant="soft" color="error">Rejected</UBadge>
                    <div>
                      <span class="font-medium">{{ item.consent.title }}</span>
                      <span class="text-dimmed ml-2 text-xs"
                        >v{{ item.consent.version }}</span
                      >
                    </div>
                  </div>
                  <div class="text-right">
                    <code class="text-dimmed text-xs">{{
                      item.consent.key
                    }}</code>
                    <div class="text-dimmed text-xs">
                      {{ formatDateTime(item.actionDate) }}
                    </div>
                  </div>
                </div>
              </div>
            </div>

            <!-- No consents -->
            <div
              v-if="
                data.user.consentStatus.pendingConsents.length === 0 &&
                data.user.consentStatus.acceptedConsents.length === 0 &&
                data.user.consentStatus.rejectedConsents.length === 0
              "
              class="text-dimmed"
            >
              No consent activity
            </div>
          </div>
        </UCard>

        <!-- Score Journal Card -->
        <UCard>
          <template #header>
            <div class="flex items-center justify-between">
              <h2 class="text-xl font-semibold">
                Score Journal
                <span
                  v-if="scoreTotalCount > 0"
                  class="text-dimmed text-sm font-normal"
                >
                  ({{ scoreTotalCount }} entries)
                </span>
              </h2>
              <UButton variant="ghost" size="sm" :to="{ name: 'admin-scores' }">
                View All
              </UButton>
            </div>
          </template>

          <div v-if="scoreEntries.length > 0" class="space-y-2">
            <div
              v-for="entry in scoreEntries"
              :key="entry.id"
              class="border-default flex items-center justify-between rounded-md border p-3"
            >
              <div class="flex items-center gap-3">
                <UBadge
                  :color="entry.points >= 0 ? 'success' : 'error'"
                  variant="soft"
                >
                  {{ entry.points >= 0 ? '+' : ''
                  }}{{ formatNumber(entry.points) }}
                </UBadge>
                <div>
                  <span class="font-medium">{{ entry.project.name }}</span>
                  <UBadge variant="subtle" size="xs" class="ml-2">
                    {{ formatSourceType(entry.sourceType) }}
                  </UBadge>
                </div>
              </div>
              <div class="text-right">
                <div
                  v-if="entry.reason"
                  class="text-dimmed max-w-xs truncate text-sm"
                >
                  {{ entry.reason }}
                </div>
                <div class="text-dimmed text-xs">
                  {{ formatScoreDate(entry.createdAt) }}
                </div>
              </div>
            </div>
            <div
              v-if="scoreTotalCount > 20"
              class="text-dimmed pt-2 text-center text-sm"
            >
              Showing 20 of {{ scoreTotalCount }} entries
            </div>
          </div>
          <div v-else class="text-dimmed">No score entries</div>
        </UCard>

        <!-- Feedback Card -->
        <UCard>
          <template #header>
            <div class="flex items-center justify-between">
              <h2 class="text-xl font-semibold">
                Feedback
                <span
                  v-if="feedbackTotalCount > 0"
                  class="text-dimmed text-sm font-normal"
                >
                  ({{ feedbackTotalCount }}
                  {{ feedbackTotalCount === 1 ? 'entry' : 'entries' }})
                </span>
              </h2>
              <UButton
                variant="ghost"
                size="sm"
                :to="{ name: 'admin-feedback' }"
              >
                View All
              </UButton>
            </div>
          </template>

          <div v-if="feedbackEntries.length > 0" class="space-y-3">
            <div
              v-for="entry in feedbackEntries"
              :key="entry.id"
              class="border-default rounded-md border p-3"
            >
              <div class="flex items-start justify-between gap-4">
                <p class="text-sm whitespace-pre-wrap">{{ entry.message }}</p>
                <UBadge
                  :color="entry.canContactMe ? 'success' : 'neutral'"
                  variant="soft"
                  class="shrink-0"
                >
                  {{ entry.canContactMe ? 'Can contact' : 'No contact' }}
                </UBadge>
              </div>
              <div
                class="text-dimmed mt-2 flex flex-wrap items-center gap-x-4 gap-y-1 text-xs"
              >
                <span>{{ formatFeedbackDate(entry.createdAt) }}</span>
                <span v-if="entry.platform">{{ entry.platform }}</span>
                <span v-if="entry.screenWidth && entry.screenHeight">
                  {{ entry.screenWidth }}x{{ entry.screenHeight }}
                </span>
                <code v-if="entry.appVersion">v{{ entry.appVersion }}</code>
              </div>
            </div>
            <div
              v-if="feedbackTotalCount > 10"
              class="text-dimmed pt-2 text-center text-sm"
            >
              Showing 10 of {{ feedbackTotalCount }} entries
            </div>
          </div>
          <div v-else class="text-dimmed">No feedback submitted</div>
        </UCard>
      </div>
    </UContainer>

    <!-- Add Role Modal -->
    <UModal v-model:open="showAddRoleModal">
      <template #header>
        <h3 class="text-lg font-semibold">Add Role</h3>
      </template>

      <template #body>
        <div class="space-y-4">
          <UFormField label="Role">
            <USelect
              v-model="newRole.role"
              :items="roleOptions"
              value-key="value"
              class="w-full"
            />
          </UFormField>

          <UFormField label="Scope Type">
            <USelect
              v-model="newRole.scopeType"
              :items="scopeTypeOptions"
              value-key="value"
              class="w-full"
            />
          </UFormField>

          <UFormField v-if="newRole.scopeType" label="Scope ID">
            <UInput
              v-model="newRole.scopeId"
              :placeholder="`Enter ${newRole.scopeType.toLowerCase()} ID`"
              class="w-full"
            />
          </UFormField>
        </div>
      </template>

      <template #footer>
        <div class="flex justify-end gap-3">
          <UButton
            variant="ghost"
            @click="
              () => {
                showAddRoleModal = false
                resetNewRoleForm()
              }
            "
          >
            Cancel
          </UButton>
          <UButton @click="handleAssignRole"> Assign Role </UButton>
        </div>
      </template>
    </UModal>
  </div>
</template>
