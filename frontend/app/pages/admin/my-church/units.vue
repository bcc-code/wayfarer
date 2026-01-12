<script setup lang="ts">
import type { TableColumn } from '@nuxt/ui'
import { generateUniqueNames } from '~/utils/unitNameGenerator'

definePageMeta({
  layout: 'church-admin',
  middleware: ['admin'],
})

gql(`
  query MyChurchUnitsPage($churchId: ID!) {
    users(filter: {churchId: $churchId}, first: 500) {
      edges {
        node {
          id
          name
          age
          teams {
            id
            name
          }
        }
      }
    }
    myCurrentProject {
      id
      name
      teams {
        id
        name
        members {
          id
          name
          user {
            id
            age
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
    now: new Date().toISOString(),
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
const { executeMutation: createTeam } = useCreateTeamMutation()
const { executeMutation: deleteTeam } = useDeleteTeamMutation()

// Get all users from church
const allUsers = computed(
  () => data.value?.users.edges.map((edge) => edge.node) ?? [],
)

// Get teams from active project
const projectTeams = computed(() => data.value?.myCurrentProject?.teams ?? [])

// Search and filter state
const unitSearch = ref('')
const personSearch = ref('')
const activeFilter = ref<'all' | 'not-in-unit' | 'in-unit'>('all')

// Expand all state
const expandAll = ref(false)

// Create unit state
const isCreatingUnit = ref(false)
const newUnitName = ref('')
const creatingUnitLoading = ref(false)

// Bulk create state
const isBulkCreating = ref(false)
const bulkCount = ref(5)
const bulkCreatingLoading = ref(false)

// Track optimistic updates
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

// Delete confirmation modal state
const deleteConfirmOpen = ref(false)
const pendingDelete = ref<{
  unitId: string
  unitName: string
} | null>(null)
const deletingLoading = ref(false)

// Filter tabs for people table
const filterTabs = [
  { label: 'Alle', value: 'all' },
  { label: 'Ikke i en unit', value: 'not-in-unit' },
  { label: 'I en unit', value: 'in-unit' },
]

// People table columns
const peopleColumns: TableColumn<(typeof allUsers.value)[number]>[] = [
  { accessorKey: 'name', header: 'Navn' },
  { accessorKey: 'age', header: 'Alder' },
]

// Filtered units based on search
const filteredUnits = computed(() => {
  if (!unitSearch.value) return projectTeams.value
  const search = unitSearch.value.toLowerCase()
  return projectTeams.value.filter((t) => t.name.toLowerCase().includes(search))
})

// Get project team IDs for filtering
const projectTeamIds = computed(
  () => new Set(projectTeams.value.map((t) => t.id)),
)

// Check if user is in a team from the current project
function isInProjectTeam(user: { teams: { id: string }[] }) {
  return user.teams.some((t) => projectTeamIds.value.has(t.id))
}

// Filtered people based on search and filter tab
const filteredPeople = computed(() => {
  let people = allUsers.value

  // Apply search
  if (personSearch.value) {
    const search = personSearch.value.toLowerCase()
    people = people.filter((p) => p.name.toLowerCase().includes(search))
  }

  // Apply filter tab - only consider teams from this project
  if (activeFilter.value === 'not-in-unit') {
    people = people.filter((p) => !isInProjectTeam(p))
  } else if (activeFilter.value === 'in-unit') {
    people = people.filter((p) => isInProjectTeam(p))
  }

  return people
})

// Start creating a new unit
function startCreateUnit() {
  isCreatingUnit.value = true
  newUnitName.value = ''
}

// Cancel creating unit
function cancelCreateUnit() {
  isCreatingUnit.value = false
  newUnitName.value = ''
}

// Save new unit
async function saveNewUnit() {
  if (!newUnitName.value.trim() || !data.value?.myCurrentProject) return

  creatingUnitLoading.value = true
  const result = await createTeam({
    projectId: data.value?.myCurrentProject.id,
    input: {
      name: newUnitName.value.trim(),
      description: '',
    },
  })

  creatingUnitLoading.value = false

  if (result.error) {
    toast.add({
      title: 'Kunne ikke opprette unit',
      description: result.error.message,
      color: 'error',
    })
    return
  }

  toast.add({
    title: 'Unit opprettet',
    color: 'success',
  })

  isCreatingUnit.value = false
  newUnitName.value = ''
  await refetch({ requestPolicy: 'network-only' })
}

// Save bulk units
async function saveBulkUnits() {
  if (!data.value?.myCurrentProject || bulkCount.value < 1) return

  bulkCreatingLoading.value = true
  const names = generateUniqueNames(bulkCount.value)

  let successCount = 0
  for (const name of names) {
    const result = await createTeam({
      projectId: data.value.myCurrentProject.id,
      input: {
        name,
        description: '',
      },
    })

    if (!result.error) {
      successCount++
    }
  }

  bulkCreatingLoading.value = false
  isBulkCreating.value = false

  if (successCount > 0) {
    toast.add({
      title: `${successCount} units opprettet`,
      color: 'success',
    })
    await refetch({ requestPolicy: 'network-only' })
  } else {
    toast.add({
      title: 'Kunne ikke opprette units',
      color: 'error',
    })
  }
}

// Add user to team
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

// Remove user from team
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

// Delete unit handlers
function handleDeleteUnit(unitId: string, unitName: string) {
  pendingDelete.value = { unitId, unitName }
  deleteConfirmOpen.value = true
}

async function confirmDelete() {
  if (!pendingDelete.value) return

  deletingLoading.value = true
  const result = await deleteTeam({ id: pendingDelete.value.unitId })
  deletingLoading.value = false

  if (result.error) {
    toast.add({
      title: 'Kunne ikke slette unit',
      description: result.error.message,
      color: 'error',
    })
  } else {
    toast.add({
      title: 'Unit slettet',
      color: 'success',
    })
    await refetch({ requestPolicy: 'network-only' })
  }

  deleteConfirmOpen.value = false
  pendingDelete.value = null
}

function cancelDelete() {
  deleteConfirmOpen.value = false
  pendingDelete.value = null
}

// Autocomplete items for adding users
const userItems = computed(() =>
  allUsers.value.map((user) => ({
    id: user.id,
    label: user.name,
    user,
  })),
)

// Get member data with optimistic updates
function getTeamMembers(team: (typeof projectTeams.value)[number]) {
  return team.members.map((member) => {
    const key = `${member.user.id}-${team.id}`
    const update = optimisticUpdates.value.get(key)
    return {
      id: member.user.id,
      name: member.name,
      age: member.user.age,
      isRemoving: update && !update.isAdding,
    }
  })
}
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
        Tilbake
      </UButton>

      <LoadingState v-if="fetching && !hasLoadedOnce" />
      <ErrorState v-else-if="error" :error />
      <div v-else-if="data" class="mt-8">
        <!-- No active project message -->
        <div
          v-if="!data.myCurrentProject"
          class="text-dimmed text-center py-12"
        >
          <Icon name="lucide:calendar-x" class="size-12 mb-4 mx-auto" />
          <p class="text-lg">Ingen aktiv konkurranse</p>
        </div>

        <!-- Two-column layout -->
        <div v-else class="grid grid-cols-1 lg:grid-cols-2 gap-8">
          <!-- Left column: Units -->
          <div>
            <h2 class="text-2xl font-semibold mb-4">Units</h2>

            <!-- Search -->
            <UInput
              v-model="unitSearch"
              placeholder="Søk..."
              icon="lucide:search"
              class="mb-4"
            />

            <!-- Expand all checkbox -->
            <UCheckbox
              v-model="expandAll"
              label="Ekspander alle units"
              class="mb-4"
            />

            <!-- Create unit buttons -->
            <div v-if="!isCreatingUnit" class="flex gap-2 mb-4">
              <UButton variant="soft" class="flex-1" @click="startCreateUnit">
                <Icon name="lucide:plus" />
                Ny unit
              </UButton>
              <UPopover v-model:open="isBulkCreating">
                <UButton variant="soft">
                  <Icon name="lucide:copy-plus" />
                  Opprett flere
                </UButton>
                <template #content>
                  <div class="p-4 w-64">
                    <p class="text-sm font-medium mb-3">Antall units</p>
                    <UInput
                      v-model.number="bulkCount"
                      type="number"
                      :min="1"
                      :max="50"
                      class="mb-3"
                    />
                    <p class="text-xs text-dimmed mb-3">
                      Navnene genereres automatisk
                    </p>
                    <div class="flex gap-2 justify-end">
                      <UButton
                        variant="ghost"
                        size="sm"
                        @click="isBulkCreating = false"
                      >
                        Avbryt
                      </UButton>
                      <UButton
                        size="sm"
                        :loading="bulkCreatingLoading"
                        :disabled="bulkCount < 1 || bulkCount > 50"
                        @click="saveBulkUnits"
                      >
                        Opprett
                      </UButton>
                    </div>
                  </div>
                </template>
              </UPopover>
            </div>

            <!-- Create unit form -->
            <div
              v-if="isCreatingUnit"
              class="mb-4 rounded-lg border border-default bg-elevated/50 p-4"
            >
              <UInput
                v-model="newUnitName"
                placeholder="Navn på unit..."
                class="mb-3"
                autofocus
              />
              <div class="flex gap-2 justify-end">
                <UButton variant="ghost" size="sm" @click="cancelCreateUnit">
                  Avbryt
                </UButton>
                <UButton
                  size="sm"
                  :loading="creatingUnitLoading"
                  :disabled="!newUnitName.trim()"
                  @click="saveNewUnit"
                >
                  Lagre
                </UButton>
              </div>
            </div>

            <!-- Units list -->
            <div class="space-y-2">
              <AdminUnitCard
                v-for="unit in filteredUnits"
                :key="unit.id"
                :unit="unit"
                :members="getTeamMembers(unit)"
                :user-items="userItems"
                :expand-all="expandAll"
                @add-member="
                  (
                    _userId: string,
                    _teamId: string,
                    _teamName: string,
                    user: {
                      id: string
                      name: string
                      teams: { id: string; name: string }[]
                    },
                  ) => handleUserSelect(user, unit.id, unit.name)
                "
                @remove-member="handleRemoveFromTeam"
                @delete-unit="handleDeleteUnit"
              />

              <p v-if="filteredUnits.length === 0" class="text-dimmed text-sm">
                Ingen units funnet
              </p>
            </div>
          </div>

          <!-- Right column: Personer -->
          <div>
            <h2 class="text-2xl font-semibold mb-4">Personer</h2>

            <!-- Search -->
            <UInput
              v-model="personSearch"
              placeholder="Søk..."
              icon="lucide:search"
              class="mb-4"
            />

            <!-- Filter tabs -->
            <UTabs
              v-model="activeFilter"
              :items="filterTabs"
              variant="pill"
              color="neutral"
            />

            <!-- People table -->
            <UTable :data="filteredPeople" :columns="peopleColumns">
              <template #name-cell="{ row }">
                <span>{{ row.original.name }}</span>
              </template>
              <template #age-cell="{ row }">
                <span class="text-dimmed">{{ row.original.age ?? '-' }}</span>
              </template>
              <template #empty>
                <p class="text-sm text-dimmed">Ingen personer funnet</p>
              </template>
            </UTable>
          </div>
        </div>
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

    <!-- Delete confirmation modal -->
    <UModal v-model:open="deleteConfirmOpen">
      <template #content>
        <div class="p-6">
          <h3 class="mb-4 text-lg font-semibold">Slett unit?</h3>
          <p class="text-dimmed mb-6">
            Er du sikker på at du vil slette
            <strong>{{ pendingDelete?.unitName }}</strong
            >? Alle medlemmer vil bli fjernet fra uniten.
          </p>
          <div class="flex justify-end gap-3">
            <UButton variant="ghost" @click="cancelDelete"> Avbryt </UButton>
            <UButton
              color="error"
              :loading="deletingLoading"
              @click="confirmDelete"
            >
              Slett
            </UButton>
          </div>
        </div>
      </template>
    </UModal>
  </div>
</template>
