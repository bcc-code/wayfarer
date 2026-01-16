<script setup lang="ts">
import {
  ConsentAction,
  ConsentManagementType,
  RoleType,
  ScopeType,
} from '~/api/generated'

definePageMeta({
  layout: 'admin',
  middleware: 'admin',
})

gql(`
	query AdminUserPage($id: ID!) {
		user(id: $id) {
			id
      personUuid
      createdAt
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

gql(`
	mutation AdminSetUserConsent($userId: ID!, $consentId: ID!, $action: ConsentAction!) {
		adminSetUserConsent(userId: $userId, consentId: $consentId, action: $action) {
			id
			action
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
const { executeMutation: adminSetUserConsent } =
  useAdminSetUserConsentMutation()
const toast = useToast()

// Permissions
const { canAssignRoles } = usePermissions()

const roleOptions = [
  { label: 'Bruker', value: RoleType.User },
  { label: 'Admin', value: RoleType.Admin },
  { label: 'Superadmin', value: RoleType.Superadmin },
  { label: 'Menighetsadmin', value: RoleType.ChurchAdmin },
  { label: 'Prosjektadmin', value: RoleType.ProjectAdmin },
  { label: 'Lagleder', value: RoleType.TeamLead },
  { label: 'M2M', value: RoleType.M2M },
]

const scopeTypeOptions = [
  { label: 'Ingen (Global)', value: null },
  { label: 'Menighet', value: ScopeType.Church },
  { label: 'Prosjekt', value: ScopeType.Project },
  { label: 'Lag', value: ScopeType.Team },
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
      title: 'Kunne ikke tildele rolle',
      description: result.error.message,
      color: 'error',
    })
    return
  }

  toast.add({
    title: 'Rolle tildelt',
    description: `Tildelte rollen ${newRole.role}`,
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
      title: 'Kunne ikke fjerne rolle',
      description: result.error.message,
      color: 'error',
    })
    return
  }

  toast.add({
    title: 'Rolle fjernet',
    description: `Fjernet rollen ${role}`,
    color: 'success',
  })

  refetch({ requestPolicy: 'network-only' })
}

async function handleRemoveConsent(consentId: string, consentTitle: string) {
  const result = await adminSetUserConsent({
    userId: route.params.userId,
    consentId,
    action: ConsentAction.Rejected,
  })

  if (result.error) {
    toast.add({
      title: 'Kunne ikke fjerne samtykke',
      description: result.error.message,
      color: 'error',
    })
    return
  }

  toast.add({
    title: 'Samtykke fjernet',
    description: `Fjernet samtykke for "${consentTitle}"`,
    color: 'success',
  })

  refetch()
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

// Feedback helpers
const feedbackEntries = computed(
  () => data.value?.feedback.edges.map((edge) => edge.node) ?? [],
)

const feedbackTotalCount = computed(() => data.value?.feedback.totalCount ?? 0)
</script>

<template>
  <div>
    <div class="border-default border-b py-2">
      <UContainer>
        <UBreadcrumb
          :items="[
            { label: 'Brukere', to: { name: 'admin-users' } },
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
        <div class="space-y-6">
          <!-- Identity -->
          <div>
            <h3 class="mb-2 text-xs font-medium uppercase tracking-wide">
              Identitet
            </h3>
            <dl
              class="text-sm grid grid-cols-[auto_1fr] gap-x-6 divide-y divide-default"
            >
              <div class="py-2 grid grid-cols-subgrid col-span-full">
                <dt class="text-muted w-36 shrink-0">ID</dt>
                <dd class="font-mono">{{ data.user.id }}</dd>
              </div>
              <div class="py-2 grid grid-cols-subgrid col-span-full">
                <dt class="text-muted w-36 shrink-0">Members-ID</dt>
                <dd class="font-medium">{{ data.user.membersId }}</dd>
              </div>
              <div class="py-2 grid grid-cols-subgrid col-span-full">
                <dt class="text-muted w-36 shrink-0">Members-UUID</dt>
                <dd class="font-medium">{{ data.user.personUuid }}</dd>
              </div>
              <div class="py-2 grid grid-cols-subgrid col-span-full">
                <dt class="text-muted w-36 shrink-0">Bruker opprettet</dt>
                <dd class="font-medium">
                  {{ formatDateTime(data.user.createdAt) }}
                </dd>
              </div>
            </dl>
          </div>

          <!-- Personal -->
          <div>
            <h3 class="mb-2 text-xs font-medium uppercase tracking-wide">
              Personlig
            </h3>
            <dl
              class="text-sm grid grid-cols-[auto_1fr] gap-x-6 divide-y divide-default"
            >
              <div class="py-2 grid grid-cols-subgrid col-span-full">
                <dt class="text-muted w-36 shrink-0">Kjønn</dt>
                <dd class="font-medium">
                  {{ capitalizeFirst(data.user.gender) }}
                </dd>
              </div>
              <div class="py-2 grid grid-cols-subgrid col-span-full">
                <dt class="text-muted w-36 shrink-0">Alder</dt>
                <dd class="font-medium">{{ data.user.age }} år</dd>
              </div>
              <div class="py-2 grid grid-cols-subgrid col-span-full">
                <dt class="text-muted w-36 shrink-0">Fødselsdato</dt>
                <dd class="font-medium">
                  {{ formatDate(data.user.birthdate) }}
                </dd>
              </div>
            </dl>
          </div>

          <!-- Church -->
          <div>
            <h3 class="mb-2 text-xs font-medium uppercase tracking-wide">
              Menighet
            </h3>
            <dl
              class="text-sm grid grid-cols-[auto_1fr] gap-x-6 divide-y divide-default"
            >
              <div class="py-2 grid grid-cols-subgrid col-span-full">
                <dt class="text-muted w-36 shrink-0">Navn</dt>
                <dd class="font-medium">{{ data.user.church.name }}</dd>
              </div>
              <div class="py-2 grid grid-cols-subgrid col-span-full">
                <dt class="text-muted w-36 shrink-0">ID</dt>
                <dd class="font-mono">{{ data.user.church.id }}</dd>
              </div>
            </dl>
          </div>
        </div>

        <!-- Roles Card -->
        <UCard>
          <template #header>
            <div class="flex items-center justify-between">
              <h2 class="text-xl font-semibold">Roller og tillatelser</h2>
              <UButton
                v-if="canAssignRoles"
                icon="i-lucide-plus"
                size="sm"
                @click="showAddRoleModal = true"
              >
                Legg til rolle
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
                  <span class="text-dimmed text-sm">Omfang: </span>
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
          <div v-else class="text-dimmed">Ingen roller tildelt</div>
        </UCard>

        <!-- Consents Card -->
        <UCard>
          <template #header>
            <h2 class="text-xl font-semibold">Samtykker</h2>
          </template>

          <div class="space-y-4">
            <!-- Pending Consents -->
            <div v-if="data.user.consentStatus.pendingConsents.length > 0">
              <h3 class="text-muted mb-2 text-sm font-medium">Ventende</h3>
              <div class="space-y-2">
                <div
                  v-for="consent in data.user.consentStatus.pendingConsents"
                  :key="consent.id"
                  class="border-default flex items-center justify-between rounded-md border p-3"
                >
                  <div class="flex items-center gap-3">
                    <UBadge variant="soft" color="warning">Ventende</UBadge>
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
              <h3 class="text-muted mb-2 text-sm font-medium">Akseptert</h3>
              <div class="space-y-2">
                <div
                  v-for="item in data.user.consentStatus.acceptedConsents"
                  :key="item.id"
                  class="border-default flex items-center justify-between gap-4 rounded-md border p-3"
                >
                  <div class="flex items-center gap-3">
                    <UBadge variant="soft" color="success">Akseptert</UBadge>
                    <div>
                      <span class="font-medium">{{ item.consent.title }}</span>
                      <span class="text-dimmed ml-2 text-xs">
                        v{{ item.consent.version }}
                      </span>
                    </div>
                  </div>
                  <UButton
                    v-if="
                      item.consent.managementType ===
                      ConsentManagementType.Local
                    "
                    color="neutral"
                    variant="soft"
                    size="sm"
                    class="ml-auto"
                    @click="
                      handleRemoveConsent(item.consent.id, item.consent.title)
                    "
                  >
                    Fjern samtykke
                  </UButton>
                  <div class="text-right">
                    <code class="text-dimmed text-xs">
                      {{ item.consent.key }}
                    </code>
                    <div class="text-dimmed text-xs">
                      {{ formatDateTime(item.actionDate) }}
                    </div>
                  </div>
                </div>
              </div>
            </div>

            <!-- Rejected Consents -->
            <div v-if="data.user.consentStatus.rejectedConsents.length > 0">
              <h3 class="text-muted mb-2 text-sm font-medium">Avvist</h3>
              <div class="space-y-2">
                <div
                  v-for="item in data.user.consentStatus.rejectedConsents"
                  :key="item.id"
                  class="border-default flex items-center justify-between rounded-md border p-3"
                >
                  <div class="flex items-center gap-3">
                    <UBadge variant="soft" color="error">Avvist</UBadge>
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
              Ingen samtykkeaktivitet
            </div>
          </div>
        </UCard>

        <!-- Score Journal Card -->
        <UCard>
          <template #header>
            <div class="flex items-center justify-between">
              <h2 class="text-xl font-semibold">
                Poenglogg
                <span
                  v-if="scoreTotalCount > 0"
                  class="text-dimmed text-sm font-normal"
                >
                  ({{ scoreTotalCount }} oppføringer)
                </span>
              </h2>
              <UButton variant="ghost" size="sm" :to="{ name: 'admin-scores' }">
                Vis alle
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
                  {{ formatDateTime(entry.createdAt) }}
                </div>
              </div>
            </div>
            <div
              v-if="scoreTotalCount > 20"
              class="text-dimmed pt-2 text-center text-sm"
            >
              Viser 20 av {{ scoreTotalCount }} oppføringer
            </div>
          </div>
          <div v-else class="text-dimmed">Ingen poengoppføringer</div>
        </UCard>

        <!-- Feedback Card -->
        <UCard>
          <template #header>
            <div class="flex items-center justify-between">
              <h2 class="text-xl font-semibold">
                Tilbakemeldinger
                <span
                  v-if="feedbackTotalCount > 0"
                  class="text-dimmed text-sm font-normal"
                >
                  ({{ feedbackTotalCount }}
                  {{ feedbackTotalCount === 1 ? 'oppføring' : 'oppføringer' }})
                </span>
              </h2>
              <UButton
                variant="ghost"
                size="sm"
                :to="{ name: 'admin-feedback' }"
              >
                Vis alle
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
                  {{ entry.canContactMe ? 'Kan kontaktes' : 'Ikke kontakt' }}
                </UBadge>
              </div>
              <div
                class="text-dimmed mt-2 flex flex-wrap items-center gap-x-4 gap-y-1 text-xs"
              >
                <span>{{ formatDateTime(entry.createdAt) }}</span>
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
              Viser 10 av {{ feedbackTotalCount }} oppføringer
            </div>
          </div>
          <div v-else class="text-dimmed">Ingen tilbakemeldinger</div>
        </UCard>
      </div>
    </UContainer>

    <!-- Add Role Modal -->
    <UModal v-model:open="showAddRoleModal">
      <template #header>
        <h3 class="text-lg font-semibold">Legg til rolle</h3>
      </template>

      <template #body>
        <div class="space-y-4">
          <UFormField label="Rolle">
            <USelect
              v-model="newRole.role"
              :items="roleOptions"
              value-key="value"
              class="w-full"
            />
          </UFormField>

          <UFormField label="Omfangstype">
            <USelect
              v-model="newRole.scopeType"
              :items="scopeTypeOptions"
              value-key="value"
              class="w-full"
            />
          </UFormField>

          <UFormField v-if="newRole.scopeType" label="Omfangs-ID">
            <UInput
              v-model="newRole.scopeId"
              :placeholder="`Skriv inn ${newRole.scopeType.toLowerCase()}-ID`"
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
            Avbryt
          </UButton>
          <UButton @click="handleAssignRole"> Tildel rolle </UButton>
        </div>
      </template>
    </UModal>
  </div>
</template>
