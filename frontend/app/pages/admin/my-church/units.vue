<script setup lang="ts">
definePageMeta({
  layout: 'church-admin',
  middleware: ['admin'],
})

gql(`
  query MyChurchUnitsPage($churchId: ID!) {
    users(filter: { churchId: $churchId }, first: 100) {
      edges {
        node {
          id
          name
          email
          teams {
            id
            name
          }
        }
      }
    }
  }
`)

const { me } = useAuth()
const { isAuthReady } = useAuthReady()
const toast = useToast()

const {
  data,
  fetching,
  error,
  executeQuery: refetch,
} = useMyChurchUnitsPageQuery({
  variables: computed(() => ({
    churchId: me.value?.church.id ?? '',
  })),
  pause: computed(() => !isAuthReady.value || !me.value?.church.id),
})

// Track initial load to avoid showing loading state during refetches
const hasLoadedOnce = ref(false)
watch(data, (newData) => {
  if (newData) hasLoadedOnce.value = true
})

const { executeMutation: addTeamMembers } = useAddTeamMembersMutation()
const { executeMutation: removeTeamMembers } = useRemoveTeamMembersMutation()

// Track optimistic assignments: Map<visibleKey, { userId, teamId, isAdding }>
const optimisticUpdates = ref<
  Map<string, { userId: string; teamId: string; isAdding: boolean }>
>(new Map())

// Move confirmation modal state
const moveConfirmOpen = ref(false)
const pendingMove = ref<{
  userId: string
  userName: string
  fromTeamName: string
  toTeamId: string
  toTeamName: string
} | null>(null)

async function handleAddToTeam(userId: string, teamId: string) {
  const key = `${userId}-${teamId}`
  optimisticUpdates.value.set(key, { userId, teamId, isAdding: true })

  const result = await addTeamMembers({
    teamId,
    userIds: [userId],
    force: true,
  })

  if (result.error) {
    optimisticUpdates.value.delete(key)
    toast.add({
      title: 'Kunne ikke legge til medlem',
      description: result.error.message,
      color: 'error',
    })
    return
  }

  await refetch({ requestPolicy: 'network-only' })
  optimisticUpdates.value.delete(key)
}

async function handleRemoveFromTeam(userId: string, teamId: string) {
  const key = `${userId}-${teamId}`
  optimisticUpdates.value.set(key, { userId, teamId, isAdding: false })

  const result = await removeTeamMembers({
    teamId,
    userIds: [userId],
  })

  if (result.error) {
    optimisticUpdates.value.delete(key)
    toast.add({
      title: 'Kunne ikke fjerne medlem',
      description: result.error.message,
      color: 'error',
    })
    return
  }

  await refetch({ requestPolicy: 'network-only' })
  optimisticUpdates.value.delete(key)
}

// Handle user selection from autocomplete
function handleUserSelect(
  user: { id: string; name: string; teams: { id: string; name: string }[] },
  targetTeamId: string,
  targetTeamName: string,
) {
  // Check if user is already in a team
  const currentTeam = user.teams[0]
  if (currentTeam) {
    if (currentTeam.id === targetTeamId) {
      // Already in this team, do nothing
      return
    }
    // Show confirmation modal
    pendingMove.value = {
      userId: user.id,
      userName: user.name,
      fromTeamName: currentTeam.name,
      toTeamId: targetTeamId,
      toTeamName: targetTeamName,
    }
    moveConfirmOpen.value = true
  } else {
    // Not in any team, add directly
    handleAddToTeam(user.id, targetTeamId)
  }
}

function confirmMove() {
  if (pendingMove.value) {
    handleAddToTeam(pendingMove.value.userId, pendingMove.value.toTeamId)
  }
  moveConfirmOpen.value = false
  pendingMove.value = null
}

function cancelMove() {
  moveConfirmOpen.value = false
  pendingMove.value = null
}

const allUsers = computed(
  () => data.value?.users.edges.map((edge) => edge.node) ?? [],
)

// Build unique teams with their members
const teams = computed(() => {
  const teamMap = new Map<
    string,
    {
      id: string
      name: string
      members: {
        id: string
        name: string
        isOptimistic?: boolean
        isRemoving?: boolean
      }[]
    }
  >()

  for (const user of allUsers.value) {
    for (const team of user.teams) {
      if (!teamMap.has(team.id)) {
        teamMap.set(team.id, { id: team.id, name: team.name, members: [] })
      }
      const key = `${user.id}-${team.id}`
      const update = optimisticUpdates.value.get(key)
      const isRemoving = update && !update.isAdding

      teamMap.get(team.id)!.members.push({
        id: user.id,
        name: user.name,
        isRemoving,
      })
    }

    // Add optimistically assigned users
    for (const [key, update] of optimisticUpdates.value) {
      if (update.userId === user.id && update.isAdding) {
        if (!teamMap.has(update.teamId)) {
          const teamName =
            allUsers.value
              .flatMap((u) => u.teams)
              .find((t) => t.id === update.teamId)?.name ?? 'Team'
          teamMap.set(update.teamId, {
            id: update.teamId,
            name: teamName,
            members: [],
          })
        }
        // Only add if not already in the team
        const teamData = teamMap.get(update.teamId)!
        if (!teamData.members.some((m) => m.id === user.id)) {
          teamData.members.push({
            id: user.id,
            name: user.name,
            isOptimistic: true,
          })
        }
      }
    }
  }

  return Array.from(teamMap.values()).sort((a, b) =>
    a.name.localeCompare(b.name),
  )
})

// Autocomplete items for each team (all users from church)
const userItems = computed(() =>
  allUsers.value.map((user) => ({
    id: user.id,
    label: user.name,
    user,
  })),
)
</script>

<template>
  <div>
    <div class="border-default border-b py-2">
      <UContainer>
        <UBreadcrumb
          :items="[
            {
              label: 'Forside',
              to: { name: 'admin-my-church' },
            },
            {
              label: 'Units',
            },
          ]"
        />
      </UContainer>
    </div>
    <UContainer class="py-6">
      <UButton
        color="neutral"
        variant="soft"
        size="lg"
        :to="{ name: 'admin-my-church' }"
      >
        <Icon name="lucide:arrow-left" />
        Tilbake til forsiden
      </UButton>

      <h1 class="mt-12 mb-6 text-4xl">Units</h1>

      <LoadingState v-if="fetching && !hasLoadedOnce" />
      <ErrorState v-else-if="error" :error />
      <div v-else-if="data">
        <div v-if="teams.length > 0" class="space-y-3">
          <div
            v-for="team in teams"
            :key="team.id"
            class="flex items-center gap-4 rounded-lg border border-default bg-elevated/50 p-3"
          >
            <!-- Team name -->
            <div class="w-40 shrink-0 font-medium">
              {{ team.name }}
            </div>

            <!-- Members -->
            <div class="flex flex-1 flex-wrap items-center gap-2">
              <UBadge
                v-for="member in team.members"
                :key="member.id"
                variant="subtle"
                size="lg"
                :class="{
                  'opacity-50': member.isOptimistic || member.isRemoving,
                }"
              >
                <Icon
                  v-if="member.isOptimistic || member.isRemoving"
                  name="lucide:loader-2"
                  class="mr-1 size-3 animate-spin"
                />
                {{ member.name }}
                <UButton
                  v-if="!member.isOptimistic && !member.isRemoving"
                  icon="lucide:x"
                  size="xs"
                  variant="link"
                  color="neutral"
                  class="-mr-1 ml-1"
                  @click="handleRemoveFromTeam(member.id, team.id)"
                />
              </UBadge>

              <!-- Add user autocomplete -->
              <UInputMenu
                :items="userItems"
                placeholder="Legg til..."
                class="w-40"
                virtualize
                :ui="{ base: 'cursor-pointer' }"
                @update:model-value="
                  (
                    item:
                      | {
                          user: {
                            id: string
                            name: string
                            teams: { id: string; name: string }[]
                          }
                        }
                      | undefined,
                  ) => item && handleUserSelect(item.user, team.id, team.name)
                "
              />
            </div>
          </div>
        </div>
        <p v-else class="text-dimmed">Ingen teams funnet</p>
      </div>
    </UContainer>

    <!-- Move confirmation modal -->
    <UModal v-model:open="moveConfirmOpen">
      <template #content>
        <div class="p-6">
          <h3 class="mb-4 text-lg font-semibold">Flytt bruker?</h3>
          <p class="text-dimmed mb-6">
            {{ pendingMove?.userName }} er allerede medlem av
            <strong>{{ pendingMove?.fromTeamName }}</strong
            >. Vil du flytte brukeren til
            <strong>{{ pendingMove?.toTeamName }}</strong
            >?
          </p>
          <div class="flex justify-end gap-3">
            <UButton variant="ghost" @click="cancelMove"> Avbryt </UButton>
            <UButton color="primary" @click="confirmMove"> Flytt </UButton>
          </div>
        </div>
      </template>
    </UModal>
  </div>
</template>
