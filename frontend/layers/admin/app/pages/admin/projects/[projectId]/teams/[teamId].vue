<script setup lang="ts">
import { rankTeamMembers } from '../../../../../utils/teamMembers'

definePageMeta({
  permission: 'teams:view',
  layout: 'admin',
})

gql(`
  query AdminTeamPage($id: ID!) {
    team(id: $id) {
      id
      name
      description
      joinCode
      leaderboardExcluded
      averageAge
      members {
        id
        name
        isTeamLead
        joinedAt
        user {
          id
          image
        }
        church {
          id
          name
        }
      }
      superTeam {
        id
        name
      }
      # LeaderboardEntry.id is the user id, and score is that member's points
      # in this team's project.
      memberLeaderboard {
        id
        score
        rank
      }
    }
  }
`)

const route = useRoute('admin-projects-projectId-teams-teamId')

const { canManageTeam } = usePermissions()
// Role-only: no team mutation accepts `project_admin`, so there is nothing
// team-specific left to check.
const canEdit = computed(() => canManageTeam())

const { isAuthReady } = useAuthReady()
const {
  data,
  fetching,
  error,
  executeQuery: refetch,
} = useAdminTeamPageQuery({
  variables: {
    id: route.params.teamId,
  },
  pause: computed(() => !isAuthReady.value),
})

// Supplies the trailing breadcrumb crumb and the navbar title; everything
// above it is derived from the route.
useAdminPage(() => data.value?.team.name)

// `averageAge` is nullable and codegen also makes it optional, so a bare
// `!== null` in the template does not narrow it.
const averageAge = computed(() => data.value?.team.averageAge ?? null)

const members = computed(() =>
  data.value
    ? rankTeamMembers(
        data.value.team.members,
        data.value.team.memberLeaderboard,
      )
    : [],
)

const { executeMutation: updateTeam } = useUpdateTeamMutation()
const { executeMutation: removeTeamMembers } = useRemoveTeamMembersMutation()
const { executeMutation: regenerateJoinCode } = useRegenerateJoinCodeMutation()
const { executeMutation: assignTeamLead } = useAssignTeamLeadMutation()
const { executeMutation: deleteTeam } = useDeleteTeamMutation()
const toast = useToast()
const { confirm } = useConfirm()

// Edit mode state
const isEditing = ref(false)
const editState = reactive({
  name: '',
  description: '',
})

function startEditing() {
  if (data.value) {
    editState.name = data.value.team.name
    editState.description = data.value.team.description
    isEditing.value = true
  }
}

function cancelEditing() {
  isEditing.value = false
}

async function saveChanges() {
  const result = await updateTeam({
    id: route.params.teamId,
    input: {
      name: editState.name,
      description: editState.description,
    },
  })

  if (result.error) {
    toast.add({
      title: 'Kunne ikke oppdatere lag',
      description: result.error.message,
      color: 'error',
    })
    return
  }

  toast.add({
    title: 'Lag oppdatert',
    color: 'success',
  })

  isEditing.value = false
  refetch({ requestPolicy: 'network-only' })
}

async function handleRemoveMember(userId: string, userName: string) {
  const confirmed = await confirm({
    title: `Fjerne ${userName} fra laget?`,
    description:
      'Medlemmet mister lagtilhørigheten, men beholder poengene sine i prosjektet.',
    confirmLabel: 'Fjern',
  })
  if (!confirmed) return

  const result = await removeTeamMembers({
    teamId: route.params.teamId,
    userIds: [userId],
  })

  if (result.error) {
    toast.add({
      title: 'Kunne ikke fjerne medlem',
      description: result.error.message,
      color: 'error',
    })
    return
  }

  toast.add({
    title: 'Medlem fjernet',
    description: `${userName} har blitt fjernet fra laget`,
    color: 'success',
  })

  refetch({ requestPolicy: 'network-only' })
}

async function handleRegenerateJoinCode() {
  const confirmed = await confirm({
    title: 'Lage ny invitasjonskode?',
    description:
      'Den gamle koden slutter å virke umiddelbart. Alle som har fått den, må få den nye.',
    confirmLabel: 'Lag ny kode',
    color: 'primary',
  })
  if (!confirmed) return

  const result = await regenerateJoinCode({
    teamId: route.params.teamId,
  })

  if (result.error) {
    toast.add({
      title: 'Kunne ikke generere ny invitasjonskode',
      description: result.error.message,
      color: 'error',
    })
    return
  }

  toast.add({
    title: 'Invitasjonskode generert',
    color: 'success',
  })

  refetch({ requestPolicy: 'network-only' })
}

async function handleAssignTeamLead(userId: string, userName: string) {
  const result = await assignTeamLead({
    teamId: route.params.teamId,
    userId,
  })

  if (result.error) {
    toast.add({
      title: 'Kunne ikke tildele lagleder',
      description: result.error.message,
      color: 'error',
    })
    return
  }

  toast.add({
    title: 'Lagleder tildelt',
    description: `${userName} er nå lagleder`,
    color: 'success',
  })

  await refetch({ requestPolicy: 'network-only' })
}

async function handleDeleteTeam() {
  const confirmed = await confirm({
    title: `Slette ${data.value?.team.name ?? 'laget'}?`,
    description: `Laget og medlemskapene til ${data.value?.team.members.length ?? 0} medlemmer slettes. Dette kan ikke angres.`,
  })
  if (!confirmed) return

  const result = await deleteTeam({
    id: route.params.teamId,
  })

  if (result.error) {
    toast.add({
      title: 'Kunne ikke slette lag',
      description: result.error.message,
      color: 'error',
    })
    return
  }

  toast.add({
    title: 'Lag slettet',
    color: 'success',
  })

  navigateTo({
    name: 'admin-projects-projectId-teams',
    params: { projectId: route.params.projectId },
  })
}

function copyJoinCode() {
  if (data.value) {
    navigator.clipboard.writeText(data.value.team.joinCode)
    toast.add({
      title: 'Invitasjonskode kopiert',
      color: 'success',
    })
  }
}

async function handleToggleLeaderboardExclusion(excluded: boolean) {
  const result = await updateTeam({
    id: route.params.teamId,
    input: { leaderboardExcluded: excluded },
  })

  if (result.error) {
    toast.add({
      title: 'Kunne ikke oppdatere toppliste-innstilling',
      description: result.error.message,
      color: 'error',
    })
    return
  }

  toast.add({
    title: excluded
      ? 'Laget er nå skjult fra topplisten'
      : 'Laget vises nå på topplisten',
    color: 'success',
  })

  refetch({ requestPolicy: 'network-only' })
}
</script>

<template>
  <!-- Capped: at full panel width each row's delete button sat far from the
       member it deletes. -->
  <div class="max-w-6xl">
    <AdminQueryState :fetching :error>
      <div v-if="data" class="space-y-8">
        <div class="flex items-start justify-between gap-4">
          <div class="min-w-0">
            <h1 class="text-3xl font-bold">{{ data.team.name }}</h1>
            <p v-if="data.team.description" class="text-dimmed">
              {{ data.team.description }}
            </p>
          </div>
          <div v-if="canEdit" class="flex shrink-0 gap-2">
            <UButton v-if="!isEditing" variant="soft" @click="startEditing">
              Rediger
            </UButton>
            <UButton variant="soft" color="error" @click="handleDeleteTeam">
              Slett
            </UButton>
          </div>
        </div>

        <AdminSection v-if="isEditing" title="Rediger lag">
          <div class="space-y-4">
            <UFormField label="Navn">
              <UInput v-model="editState.name" class="w-full" />
            </UFormField>
            <UFormField label="Beskrivelse">
              <UTextarea
                v-model="editState.description"
                class="w-full"
                autoresize
              />
            </UFormField>
          </div>
          <!-- `AdminSection` has no footer slot; the actions sit at the end
               of the section instead. -->
          <div class="flex justify-end gap-3 pt-4">
            <UButton variant="ghost" @click="cancelEditing">Avbryt</UButton>
            <UButton @click="saveChanges">Lagre endringer</UButton>
          </div>
        </AdminSection>

        <AdminSection title="Detaljer">
          <!-- No project row: the breadcrumb names it, and this page is only
               reachable through it. No member count either — the Medlemmer
               heading below carries it. -->
          <dl class="divide-default divide-y text-sm">
            <div class="flex gap-6 py-2">
              <dt class="text-muted w-36 shrink-0">Superlag</dt>
              <dd v-if="data.team.superTeam">
                <NuxtLink
                  :to="{
                    name: 'admin-projects-projectId-superteams-superTeamId',
                    params: {
                      projectId: route.params.projectId,
                      superTeamId: data.team.superTeam.id,
                    },
                  }"
                  class="hover:underline"
                >
                  {{ data.team.superTeam.name }}
                </NuxtLink>
              </dd>
              <dd v-else class="text-dimmed">Ingen</dd>
            </div>
            <div class="flex items-center gap-6 py-2">
              <dt class="text-muted w-36 shrink-0">Invitasjonskode</dt>
              <dd class="flex items-center gap-2">
                <code class="bg-elevated rounded px-2 py-1">
                  {{ data.team.joinCode }}
                </code>
                <UButton
                  variant="ghost"
                  size="xs"
                  icon="i-lucide-copy"
                  @click="copyJoinCode"
                />
                <UButton
                  v-if="canEdit"
                  variant="ghost"
                  size="xs"
                  icon="i-lucide-refresh-cw"
                  @click="handleRegenerateJoinCode"
                />
              </dd>
            </div>
            <div class="flex gap-6 py-2">
              <dt class="text-muted w-36 shrink-0">Gjennomsnittsalder</dt>
              <dd v-if="averageAge !== null">{{ averageAge.toFixed(1) }} år</dd>
              <dd v-else class="text-dimmed">Ukjent</dd>
            </div>
            <div class="flex items-center gap-6 py-2">
              <dt class="text-muted w-36 shrink-0">Skjul fra toppliste</dt>
              <dd class="flex items-center gap-2">
                <USwitch
                  :model-value="data.team.leaderboardExcluded"
                  :disabled="!canEdit"
                  @update:model-value="handleToggleLeaderboardExclusion"
                />
                <UTooltip
                  text="Når aktivert vil dette laget ikke vises på topplisten"
                  :delay-duration="200"
                >
                  <Icon name="lucide:info" class="text-dimmed size-4" />
                </UTooltip>
              </dd>
            </div>
            <!-- Last, matching the user, church and consent pages. -->
            <div class="flex gap-6 py-2">
              <dt class="text-muted w-36 shrink-0">Lag-ID</dt>
              <dd class="font-mono">{{ data.team.id }}</dd>
            </div>
          </dl>
        </AdminSection>

        <AdminSection title="Medlemmer" :count="members.length">
          <!-- Rows carry no border of their own: they sit on the section's
               surface, not in boxes within it. -->
          <div v-if="members.length" class="divide-default divide-y">
            <div
              v-for="member in members"
              :key="member.id"
              class="flex items-center gap-3 py-2"
            >
              <UAvatar
                :src="member.user.image ?? undefined"
                :alt="member.name"
                size="sm"
              />
              <div class="min-w-0 grow">
                <div class="flex items-center gap-2">
                  <NuxtLink
                    :to="{
                      name: 'admin-users-userId',
                      params: { userId: member.user.id },
                    }"
                    class="truncate font-medium hover:underline"
                  >
                    {{ member.name }}
                  </NuxtLink>
                  <UBadge
                    v-if="member.isTeamLead"
                    variant="soft"
                    size="xs"
                    color="primary"
                  >
                    Leder
                  </UBadge>
                </div>
                <p class="text-muted truncate text-xs">
                  {{ member.church.name }}
                  <template v-if="member.joinedAt">
                    &middot; Ble med {{ formatDate(member.joinedAt) }}
                  </template>
                </p>
              </div>
              <div class="shrink-0 text-right">
                <p class="text-sm font-medium tabular-nums">
                  {{ formatNumber(member.score) }}
                </p>
                <p class="text-dimmed text-xs">poeng</p>
              </div>
              <div v-if="canEdit" class="flex shrink-0 items-center gap-2">
                <UButton
                  v-if="!member.isTeamLead"
                  variant="ghost"
                  size="xs"
                  @click="handleAssignTeamLead(member.user.id, member.name)"
                >
                  Gjør til leder
                </UButton>
                <UButton
                  icon="i-lucide-trash-2"
                  color="error"
                  variant="ghost"
                  size="sm"
                  @click="handleRemoveMember(member.user.id, member.name)"
                />
              </div>
            </div>
          </div>
          <p v-else class="text-dimmed text-sm">Ingen medlemmer</p>
        </AdminSection>
      </div>
    </AdminQueryState>
  </div>
</template>
