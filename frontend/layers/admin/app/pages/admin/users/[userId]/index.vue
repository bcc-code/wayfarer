<script setup lang="ts">
import {
  ConsentAction,
  ConsentManagementType,
  RoleType,
  ScopeType,
} from '~/api/generated'

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

gql(`
	mutation AdminSetUserConsent($userId: ID!, $consentId: ID!, $action: ConsentAction!) {
		adminSetUserConsent(userId: $userId, consentId: $consentId, action: $action) {
			id
			action
		}
	}
`)

gql(`
	mutation SyncUser($userId: ID!) {
		syncUser(userId: $userId) {
			user {
				id
				name
				personUuid
				churchLockedUntil
				church {
					id
					name
				}
			}
			contentEventsProcessed
			churchUpdated
			churchLockSkipped
			personUuidUpdated
		}
	}
`)

gql(`
	mutation LockUserChurch($userId: ID!) {
		lockUserChurch(userId: $userId) {
			id
			churchLockedUntil
			church {
				id
				name
			}
		}
	}
`)

gql(`
	mutation UnlockUserChurch($userId: ID!) {
		unlockUserChurch(userId: $userId) {
			id
			churchLockedUntil
			church {
				id
				name
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
type UserRole = NonNullable<AdminUserPageQuery['user']>['roles'][number]

/**
 * What a role is scoped to, by name.
 *
 * `RoleScope` resolves `church`/`project`/`team` server-side, so the card can
 * say "Østfold" where it used to print `CH01K9VZ865699692N7FVTXYR4AQ`. The id
 * is the fallback rather than the default: a scope pointing at something
 * deleted still needs to render, and then the raw id is the only honest thing
 * left to show.
 */
function scopeLabel(scope: UserRole['scope']): string | undefined {
  if (!scope) return undefined
  return (
    scope.church?.name ?? scope.project?.name ?? scope.team?.name ?? scope.id
  )
}

/**
 * A readable language, not a bare code. The DB stores `no` where the app uses
 * `nb`, so it goes through the existing mapping before being named.
 */
const LANGUAGE_NAMES = new Intl.DisplayNames(['nb'], { type: 'language' })
const languageLabel = computed(() => {
  const code = data.value?.user.language
  if (!code) return undefined
  const locale = dbLanguageToLocale(code)
  try {
    return LANGUAGE_NAMES.of(locale) ?? locale
  } catch {
    // Intl throws on a malformed tag; the raw code is better than nothing.
    return locale
  }
})

/**
 * The three consent lists as one, grouped by sorting rather than by heading.
 *
 * Each row already carries a status badge, so the "Ventende / Akseptert /
 * Avvist" sub-headings said the same word twice. Sorting on status keeps the
 * grouping — pending first, because it is the only one that wants an admin to
 * do something — while dropping the duplication, and it normalises the two
 * different shapes the API returns: `pendingConsents` are bare `Consent`s,
 * while accepted and rejected are `UserConsent`s wrapping one.
 */
type ConsentRowStatus = 'pending' | 'accepted' | 'rejected'

interface ConsentRow {
  /** Unique across the three source lists, which can share ids. */
  rowKey: string
  status: ConsentRowStatus
  title: string
  version: number
  consentKey: string
  /** Only accepted/rejected rows have a decision date. */
  actionDate?: string
  /** Only locally-managed accepted consents can be withdrawn here. */
  removableConsentId?: string
}

const CONSENT_STATUS_ORDER: Record<ConsentRowStatus, number> = {
  pending: 0,
  accepted: 1,
  rejected: 2,
}

const CONSENT_STATUS_LABELS: Record<ConsentRowStatus, string> = {
  pending: 'Ventende',
  accepted: 'Akseptert',
  rejected: 'Avvist',
}

const CONSENT_STATUS_COLORS: Record<
  ConsentRowStatus,
  'warning' | 'success' | 'error'
> = {
  pending: 'warning',
  accepted: 'success',
  rejected: 'error',
}

const consentRows = computed<ConsentRow[]>(() => {
  const status = data.value?.user.consentStatus
  if (!status) return []

  const rows: ConsentRow[] = [
    ...status.pendingConsents.map((consent) => ({
      rowKey: `pending-${consent.id}`,
      status: 'pending' as const,
      title: consent.title,
      version: consent.version,
      consentKey: consent.key,
    })),
    ...status.acceptedConsents.map((item) => ({
      rowKey: `accepted-${item.id}`,
      status: 'accepted' as const,
      title: item.consent.title,
      version: item.consent.version,
      consentKey: item.consent.key,
      actionDate: item.actionDate,
      removableConsentId:
        item.consent.managementType === ConsentManagementType.Local
          ? item.consent.id
          : undefined,
    })),
    ...status.rejectedConsents.map((item) => ({
      rowKey: `rejected-${item.id}`,
      status: 'rejected' as const,
      title: item.consent.title,
      version: item.consent.version,
      consentKey: item.consent.key,
      actionDate: item.actionDate,
    })),
  ]

  return rows.sort(
    (a, b) =>
      CONSENT_STATUS_ORDER[a.status] - CONSENT_STATUS_ORDER[b.status] ||
      a.title.localeCompare(b.title, 'nb'),
  )
})

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

const { executeMutation: assignRole } = useAssignRoleMutation()
const { executeMutation: revokeRole } = useRevokeRoleMutation()
const { executeMutation: adminSetUserConsent } =
  useAdminSetUserConsentMutation()
const { executeMutation: syncUserMutation } = useSyncUserMutation()
const { executeMutation: lockUserChurchMutation } = useLockUserChurchMutation()
const { executeMutation: unlockUserChurchMutation } =
  useUnlockUserChurchMutation()
const toast = useToast()
const syncing = ref(false)
const locking = ref(false)
const unlocking = ref(false)

const isChurchLocked = computed(() => {
  const lockedUntil = data.value?.user.churchLockedUntil
  if (!lockedUntil) return false
  return new Date(lockedUntil) > new Date()
})

// Permissions
const { canAssignRoles, canCheckAchievements } = usePermissions()

const roleLabels: Record<RoleType, string> = {
  [RoleType.User]: 'Bruker',
  [RoleType.Admin]: 'Admin',
  [RoleType.Superadmin]: 'Superadmin',
  [RoleType.ChurchAdmin]: 'Menighetsadmin',
  [RoleType.ProjectAdmin]: 'Prosjektadmin',
  [RoleType.TeamLead]: 'Lagleder',
  [RoleType.M2M]: 'M2M',
}

const roleOptions = Object.entries(roleLabels).map(([value, label]) => ({
  label,
  value: value as RoleType,
}))

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

const showRemoveConsentModal = ref(false)
const consentToRemove = ref<{ id: string; title: string } | null>(null)

function openRemoveConsentModal(consentId: string, consentTitle: string) {
  consentToRemove.value = { id: consentId, title: consentTitle }
  showRemoveConsentModal.value = true
}

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

async function handleRemoveConsent() {
  if (!consentToRemove.value) return

  const { id, title } = consentToRemove.value
  const result = await adminSetUserConsent({
    userId: route.params.userId,
    consentId: id,
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
    description: `Fjernet samtykke for "${title}"`,
    color: 'success',
  })

  showRemoveConsentModal.value = false
  consentToRemove.value = null
  refetch()
}

async function handleSyncUser() {
  syncing.value = true
  const result = await syncUserMutation({
    userId: route.params.userId,
  })
  syncing.value = false

  if (result.error) {
    toast.add({
      title: 'Synkronisering feilet',
      description: result.error.message,
      color: 'error',
    })
    return
  }

  const syncResult = result.data?.syncUser
  const details: string[] = []
  if (syncResult) {
    if (syncResult.contentEventsProcessed > 0)
      details.push(`${syncResult.contentEventsProcessed} innholdseventer`)
    if (syncResult.churchUpdated) details.push('menighet oppdatert')
    if (syncResult.churchLockSkipped)
      details.push('menighet hoppet over (last)')
    if (syncResult.personUuidUpdated) details.push('person-UUID oppdatert')
  }

  toast.add({
    title: 'Synkronisering fullført',
    description: details.length > 0 ? details.join(', ') : 'Ingen endringer',
    color: 'success',
  })

  refetch({ requestPolicy: 'network-only' })
}

async function handleLockChurch() {
  locking.value = true
  const result = await lockUserChurchMutation({
    userId: route.params.userId,
  })
  locking.value = false

  if (result.error) {
    toast.add({
      title: 'Kunne ikke låse menighet',
      description: result.error.message,
      color: 'error',
    })
    return
  }

  toast.add({
    title: 'Menighet låst',
    description: 'Menigheten er låst i 6 måneder',
    color: 'success',
  })

  refetch({ requestPolicy: 'network-only' })
}

async function handleUnlockChurch() {
  unlocking.value = true
  const result = await unlockUserChurchMutation({
    userId: route.params.userId,
  })
  unlocking.value = false

  if (result.error) {
    toast.add({
      title: 'Kunne ikke låse opp menighet',
      description: result.error.message,
      color: 'error',
    })
    return
  }

  toast.add({
    title: 'Menighet låst opp',
    description: 'Menighetslåsing er fjernet',
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
    <div>
      <AdminQueryState :fetching :error>
        <div v-if="data" class="space-y-6">
          <!--
            Identity card rather than a bare name. The header used to carry the
            least information on the page — everything identifying (church,
            roles, teams, points) sat below the fold, while the one prominent
            element held a single string. The avatar was already being fetched
            and thrown away.
          -->
          <div class="flex flex-wrap items-start justify-between gap-4">
            <div class="flex items-start gap-4">
              <UAvatar
                :src="data.user.image ?? undefined"
                :alt="data.user.name"
                size="3xl"
              />
              <div class="min-w-0">
                <h1 class="text-3xl font-bold">{{ data.user.name }}</h1>
                <p v-if="data.user.email" class="text-muted text-sm">
                  {{ data.user.email }}
                </p>
                <!--
                  Separated and spelled out: "Østfold 25 år nb" ran together as
                  one string, and a bare language code says nothing to a reader
                  who does not already know it.
                -->
                <div
                  class="text-muted mt-1 flex flex-wrap items-center gap-x-2 gap-y-1 text-sm"
                >
                  <NuxtLink
                    :to="{
                      name: 'admin-churches-churchId',
                      params: { churchId: data.user.church.id },
                    }"
                    class="hover:underline"
                  >
                    {{ data.user.church.name }}
                  </NuxtLink>
                  <template v-if="data.user.age">
                    <span aria-hidden="true">·</span>
                    <span>{{ data.user.age }} år</span>
                  </template>
                  <template v-if="data.user.language">
                    <span aria-hidden="true">·</span>
                    <span>{{ languageLabel }}</span>
                  </template>
                </div>
              </div>
            </div>
            <div class="flex flex-wrap gap-2">
              <UButton
                v-if="canCheckAchievements"
                icon="i-lucide-trophy"
                variant="soft"
                :to="{
                  name: 'admin-users-userId-achievements',
                  params: { userId: route.params.userId },
                }"
              >
                Sjekk prestasjoner
              </UButton>
              <UButton
                icon="i-lucide-refresh-cw"
                variant="soft"
                :loading="syncing"
                @click="handleSyncUser"
              >
                Synkroniser
              </UButton>
            </div>
          </div>

          <!--
            Sync-lock and the technical identifiers share one quiet row. The
            lock was buried inside a definition list about the church, which
            hid a consequential action; promoting it to a full-width card then
            overstated it — it is the rarest thing anyone does here. A labelled
            row with the state spelled out is the middle ground.
          -->
          <div
            class="border-default flex flex-wrap items-center gap-x-6 gap-y-2 border-y py-2 text-sm"
          >
            <UCollapsible>
              <UButton
                variant="link"
                color="neutral"
                size="sm"
                class="px-0"
                trailing-icon="i-lucide-chevron-down"
                label="Tekniske detaljer"
              />
              <template #content>
                <dl
                  class="divide-default mt-2 grid grid-cols-[auto_1fr] gap-x-6 divide-y text-sm"
                >
                  <div class="col-span-full grid grid-cols-subgrid py-2">
                    <dt class="text-muted w-36 shrink-0">Bruker-ID</dt>
                    <dd class="font-mono">{{ data.user.id }}</dd>
                  </div>
                  <div class="col-span-full grid grid-cols-subgrid py-2">
                    <dt class="text-muted w-36 shrink-0">Members-ID</dt>
                    <dd class="font-mono">{{ data.user.membersId }}</dd>
                  </div>
                  <div class="col-span-full grid grid-cols-subgrid py-2">
                    <dt class="text-muted w-36 shrink-0">Members-UUID</dt>
                    <dd class="font-mono">{{ data.user.personUuid }}</dd>
                  </div>
                  <div class="col-span-full grid grid-cols-subgrid py-2">
                    <dt class="text-muted w-36 shrink-0">Menighets-ID</dt>
                    <dd class="font-mono">{{ data.user.church.id }}</dd>
                  </div>
                  <div class="col-span-full grid grid-cols-subgrid py-2">
                    <dt class="text-muted w-36 shrink-0">Bruker opprettet</dt>
                    <dd>{{ formatDateTime(data.user.createdAt) }}</dd>
                  </div>
                </dl>
              </template>
            </UCollapsible>

            <div v-if="canAssignRoles" class="flex items-center gap-2">
              <span class="text-muted">Menighetslås:</span>
              <template v-if="isChurchLocked">
                <UBadge color="warning" variant="soft">
                  Låst til {{ formatDateTime(data.user.churchLockedUntil!) }}
                </UBadge>
                <UButton
                  size="xs"
                  variant="soft"
                  color="neutral"
                  :loading="unlocking"
                  @click="handleUnlockChurch"
                >
                  Lås opp
                </UButton>
              </template>
              <template v-else>
                <span class="text-dimmed">Ikke låst</span>
                <UButton
                  size="xs"
                  variant="soft"
                  color="neutral"
                  :loading="locking"
                  @click="handleLockChurch"
                >
                  Lås i 6 måneder
                </UButton>
              </template>
            </div>
          </div>

          <!-- Teams Card -->
          <UCard>
            <template #header>
              <h2 class="text-xl font-semibold">
                Lag
                <span
                  v-if="data.user.teams.length > 0"
                  class="text-dimmed text-sm font-normal"
                >
                  ({{ data.user.teams.length }})
                </span>
              </h2>
            </template>

            <div v-if="data.user.teams.length > 0" class="space-y-2">
              <NuxtLink
                v-for="team in data.user.teams"
                :key="team.id"
                :to="{
                  name: 'admin-projects-projectId-teams-teamId',
                  params: {
                    projectId: team.parentProject.id,
                    teamId: team.id,
                  },
                }"
                class="border-default flex items-center justify-between rounded-md border p-3 hover:bg-elevated transition-colors"
              >
                <div>
                  <span class="font-medium">{{ team.name }}</span>
                  <div class="text-muted text-xs">
                    {{ team.parentProject.name }}
                  </div>
                </div>
                <Icon name="lucide:chevron-right" class="size-4 text-dimmed" />
              </NuxtLink>
            </div>
            <div v-else class="text-dimmed">Ikke med i noen lag</div>
          </UCard>

          <!-- Roles Card -->
          <UCard>
            <template #header>
              <div class="flex items-center justify-between">
                <h2 class="text-xl font-semibold">Roller og tillatelser</h2>
                <UButton
                  v-if="canAssignRoles"
                  icon="i-lucide-plus"
                  size="sm"
                  @click="
                    () => {
                      showAddRoleModal = true
                    }
                  "
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
                    {{ roleLabels[role.role] ?? role.role }}
                  </UBadge>
                  <div v-if="role.scope" class="text-sm">
                    <span class="text-dimmed">{{
                      capitalizeFirst(role.scope.type)
                    }}</span>
                    <span class="ml-2 font-medium">
                      {{ scopeLabel(role.scope) }}
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
              <div class="flex flex-wrap items-center justify-between gap-2">
                <h2 class="text-xl font-semibold">
                  Samtykker
                  <span
                    v-if="consentRows.length"
                    class="text-dimmed text-sm font-normal"
                  >
                    ({{ consentRows.length }})
                  </span>
                </h2>
              </div>
            </template>

            <!--
              One list, grouped by sorting: pending first, then accepted, then
              rejected, alphabetical within each. The badge carries the status,
              so the old sub-headings repeated it.
            -->
            <div v-if="consentRows.length" class="space-y-2">
              <div
                v-for="row in consentRows"
                :key="row.rowKey"
                class="border-default flex flex-wrap items-center justify-between gap-3 rounded-md border p-3"
              >
                <div class="flex min-w-0 items-center gap-3">
                  <UBadge
                    variant="soft"
                    :color="CONSENT_STATUS_COLORS[row.status]"
                  >
                    {{ CONSENT_STATUS_LABELS[row.status] }}
                  </UBadge>
                  <div class="min-w-0">
                    <span class="font-medium">{{ row.title }}</span>
                    <span class="text-dimmed ml-2 text-xs">
                      v{{ row.version }}
                    </span>
                  </div>
                </div>

                <div class="ms-auto flex items-center gap-3">
                  <UButton
                    v-if="row.removableConsentId"
                    color="neutral"
                    variant="soft"
                    size="sm"
                    @click="
                      openRemoveConsentModal(row.removableConsentId, row.title)
                    "
                  >
                    Fjern samtykke
                  </UButton>
                  <div class="text-right">
                    <code class="text-dimmed text-xs">
                      {{ row.consentKey }}
                    </code>
                    <div v-if="row.actionDate" class="text-dimmed text-xs">
                      {{ formatDateTime(row.actionDate) }}
                    </div>
                  </div>
                </div>
              </div>
            </div>
            <div v-else class="text-dimmed">Ingen samtykkeaktivitet</div>
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
                    {{
                      feedbackTotalCount === 1 ? 'oppføring' : 'oppføringer'
                    }})
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

          <!--
            Both the total and the journal are filtered to the *current*
            project (`points(projectId:)`, `adminScoreJournal(filter:)`), but
            the panel said only "Poenglogg" and "N poeng" — which reads as the
            user's lifetime total. Naming the project is the whole fix.
          -->
          <UCard>
            <template #header>
              <div class="flex flex-wrap items-center justify-between gap-2">
                <h2 class="text-xl font-semibold">
                  Poenglogg
                  <span
                    v-if="currentProjectName"
                    class="text-muted text-sm font-normal"
                  >
                    i {{ currentProjectName }}
                  </span>
                  <UBadge color="neutral" variant="soft">
                    {{ data.user.points }} poeng
                  </UBadge>
                  <span
                    v-if="scoreTotalCount > 0"
                    class="text-dimmed text-sm font-normal"
                  >
                    ({{ scoreTotalCount }} oppføringer)
                  </span>
                </h2>
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
                v-if="scoreTotalCount > 100"
                class="text-dimmed pt-2 text-center text-sm"
              >
                Viser 100 av {{ scoreTotalCount }} oppføringer
              </div>
            </div>
            <div v-else class="text-dimmed">Ingen poengoppføringer</div>
          </UCard>
        </div>
      </AdminQueryState>
    </div>

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

    <!-- Remove Consent Confirmation Modal -->
    <UModal v-model:open="showRemoveConsentModal">
      <template #header>
        <h3 class="text-lg font-semibold">Fjern samtykke</h3>
      </template>

      <template #body>
        <p>
          Er du sikker på at du vil fjerne samtykke for
          <strong>{{ consentToRemove?.title }}</strong
          >?
        </p>
        <p class="text-dimmed mt-2 text-sm">
          Samtykket vil bli markert som avvist.
        </p>
      </template>

      <template #footer>
        <div class="flex justify-end gap-3 w-full">
          <UButton
            variant="ghost"
            color="neutral"
            @click="
              () => {
                showRemoveConsentModal = false
                consentToRemove = null
              }
            "
          >
            Avbryt
          </UButton>
          <UButton color="error" @click="handleRemoveConsent">
            Fjern samtykke
          </UButton>
        </div>
      </template>
    </UModal>
  </div>
</template>
