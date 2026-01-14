<script setup lang="ts">
import type { TabsItem } from '@nuxt/ui'
import { VueDraggable } from 'vue-draggable-plus'
import { generateUniqueNames } from '~/utils/unitNameGenerator'

definePageMeta({
  layout: 'church-admin',
  middleware: ['admin'],
})

gql(`
  query MyChurchUnitsPage($filter: UserFilter) {
    users(filter: $filter, first: 500) {
      edges {
        node {
          id
          name
          age
          gender
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
        leaderboardExcluded
        averageAge
        members {
          id
          name
          isTeamLead
          user {
            id
            age
            gender
          }
        }
      }
    }
  }
`)

const { me } = useAuth()
const { isAuthReady } = useAuthReady()
const toast = useToast()
const { t } = useI18n()

// Filters
const filters = reactive({
  showO36: false,
})

const {
  data,
  fetching,
  error,
  executeQuery: refetch,
} = useMyChurchUnitsPageQuery({
  variables: computed(() => ({
    filter: {
      churchId: me.value?.church.id,
      minAge: 13,
      maxAge: filters.showO36 ? undefined : 36,
    },
  })),
  pause: computed(() => !isAuthReady.value || !me.value?.church.id),
})

// Expand all state
const expandAll = ref(true)

// Track initial load to avoid showing loading state during refetches
const hasLoadedOnce = ref(false)
watch(data, (newData) => {
  if (!newData) return
  hasLoadedOnce.value = true

  if (newData.myCurrentProject.teams.length > 10) {
    expandAll.value = false
  }
})

const { executeMutation: addTeamMembers } = useAddTeamMembersMutation()
const { executeMutation: removeTeamMembers } = useRemoveTeamMembersMutation()
const { executeMutation: createTeam } = useCreateTeamMutation()
const { executeMutation: updateTeam } = useUpdateTeamMutation()
const { executeMutation: deleteTeam } = useDeleteTeamMutation()
const { executeMutation: assignTeamLead } = useAssignTeamLeadMutation()

// Get all users from church
const allUsers = computed(
  () => data.value?.users.edges.map((edge) => edge.node) ?? [],
)

// Get teams from active project
const projectTeams = computed(() => data.value?.myCurrentProject?.teams ?? [])

// Search and filter state
const unitSearch = ref('')
const personSearch = ref('')
const activeFilter = ref<'all' | 'not-in-unit' | 'in-unit'>('not-in-unit')

// Create unit state
const isCreatingUnit = ref(false)
const newUnitName = ref('')
const creatingUnitLoading = ref(false)

// Bulk create state
const isBulkCreating = ref(false)
const bulkCount = ref(5)
const bulkCreatingLoading = ref(false)

// Track optimistic updates for add/remove operations
const optimisticUpdates = ref<
  Map<string, { userId: string; teamId: string; isAdding: boolean }>
>(new Map())

// Track optimistic moves with full user data for immediate UI updates
const optimisticMoves = ref<
  Map<
    string,
    {
      userId: string
      userName: string
      userAge: number | null
      userGender: string
      userTeams: { id: string; name: string }[]
      fromTeamId: string | null
      toTeamId: string
    }
  >
>(new Map())

// Delete confirmation modal state
const deleteConfirmOpen = ref(false)
const pendingDelete = ref<{
  unitId: string
  unitName: string
} | null>(null)
const deletingLoading = ref(false)

// Bulk selection state
const selectedUnitIds = ref<Set<string>>(new Set())
const bulkDeleteConfirmOpen = ref(false)
const bulkDeletingLoading = ref(false)

const hasSelection = computed(() => selectedUnitIds.value.size > 0)
const allSelected = computed(
  () =>
    filteredUnits.value.length > 0 &&
    filteredUnits.value.every((u) => selectedUnitIds.value.has(u.id)),
)

function toggleUnitSelection(unitId: string) {
  const newSet = new Set(selectedUnitIds.value)
  if (newSet.has(unitId)) {
    newSet.delete(unitId)
  } else {
    newSet.add(unitId)
  }
  selectedUnitIds.value = newSet
}

function toggleSelectAll() {
  if (allSelected.value) {
    selectedUnitIds.value = new Set()
  } else {
    selectedUnitIds.value = new Set(filteredUnits.value.map((u) => u.id))
  }
}

function clearSelection() {
  selectedUnitIds.value = new Set()
}

// Filter tabs for people table with counts
const filterTabs = computed<TabsItem[]>(() => {
  const total = allUsers.value.length
  const inUnit = allUsers.value.filter((u) => isInProjectTeam(u)).length
  const notInUnit = total - inUnit

  return [
    {
      label: t('admin.units.filter.all'),
      value: 'all',
      badge: { label: total, variant: 'subtle' },
    },
    {
      label: t('admin.units.filter.notInUnit'),
      value: 'not-in-unit',
      badge: { label: notInUnit, variant: 'subtle' },
    },
    {
      label: t('admin.units.filter.inUnit'),
      value: 'in-unit',
      badge: { label: inUnit, variant: 'subtle' },
    },
  ]
})

// Filtered and sorted units based on search
const filteredUnits = computed(() => {
  let units = projectTeams.value

  // Filter by search
  if (unitSearch.value) {
    const search = unitSearch.value.toLowerCase()
    units = units.filter(
      (t) =>
        t.name.toLowerCase().includes(search) ||
        t.members.some((m) => m.name.toLowerCase().includes(search)),
    )
  }

  // Sort naturally (so "Unit 2" comes before "Unit 10")
  return [...units].sort((a, b) =>
    a.name.localeCompare(b.name, undefined, { numeric: true }),
  )
})

// Get project team IDs for filtering
const projectTeamIds = computed(
  () => new Set(projectTeams.value.map((t) => t.id)),
)

// Check if user is in a team from the current project (accounting for optimistic moves)
function isInProjectTeam(user: { id: string; teams: { id: string }[] }) {
  // Check if user has an optimistic move
  const move = optimisticMoves.value.get(user.id)
  if (move) {
    // User is optimistically in a project team
    return projectTeamIds.value.has(move.toTeamId)
  }
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
  if (!data.value?.myCurrentProject || !me.value?.church) return

  // Auto-generate name if empty
  const existingNames = projectTeams.value.map((t) => t.name)
  const name =
    newUnitName.value.trim() ||
    generateUniqueNames(1, me.value.church.name, existingNames)[0] ||
    `${me.value.church.name} 1`

  creatingUnitLoading.value = true
  const result = await createTeam({
    projectId: data.value?.myCurrentProject.id,
    input: {
      name,
      description: '',
    },
  })

  creatingUnitLoading.value = false

  if (result.error) {
    toast.add({
      title: t('admin.units.errors.createFailed'),
      description: result.error.message,
      color: 'error',
    })
    return
  }

  toast.add({
    title: t('admin.units.success.created'),
    color: 'success',
  })

  isCreatingUnit.value = false
  newUnitName.value = ''
  await refetch({ requestPolicy: 'network-only' })
}

// Save bulk units
async function saveBulkUnits() {
  if (!data.value?.myCurrentProject || !me.value?.church || bulkCount.value < 1)
    return

  const projectId = data.value.myCurrentProject.id

  bulkCreatingLoading.value = true
  const existingNames = projectTeams.value.map((t) => t.name)
  const names = generateUniqueNames(
    bulkCount.value,
    me.value.church.name,
    existingNames,
  )

  const results = await Promise.allSettled(
    names.map((name) =>
      createTeam({
        projectId,
        input: {
          name,
          description: '',
        },
      }),
    ),
  )
  const successCount = results.filter(
    (r) => r.status === 'fulfilled' && !r.value.error,
  ).length

  bulkCreatingLoading.value = false
  isBulkCreating.value = false

  if (successCount > 0) {
    toast.add({
      title: t('admin.units.success.bulkCreated', { count: successCount }),
      color: 'success',
    })
    await refetch({ requestPolicy: 'network-only' })
  } else {
    toast.add({
      title: t('admin.units.errors.bulkCreateFailed'),
      color: 'error',
    })
  }
}

// Add user to team with optimistic UI update
async function handleAddToTeam(
  user: {
    id: string
    name: string
    age?: number | null
    gender: string
    teams: { id: string; name: string }[]
  },
  teamId: string,
) {
  // Find if user is currently in a project team (for move operations)
  const currentProjectTeam = user.teams.find((t) =>
    projectTeamIds.value.has(t.id),
  )
  const fromTeamId = currentProjectTeam?.id ?? null

  // Set up optimistic move - this immediately updates the UI
  optimisticMoves.value.set(user.id, {
    userId: user.id,
    userName: user.name,
    userAge: user.age ?? null,
    userGender: user.gender,
    userTeams: user.teams,
    fromTeamId,
    toTeamId: teamId,
  })

  try {
    const result = await addTeamMembers({
      teamId,
      userIds: [user.id],
      force: true,
    })

    if (result.error) {
      toast.add({
        title: t('admin.units.errors.addMemberFailed'),
        description: result.error.message,
        color: 'error',
      })
      return
    }

    // Sync with server - keep optimistic state until refetch completes
    await refetch({ requestPolicy: 'network-only' })
  } finally {
    // Always clear ALL optimistic state after refetch
    optimisticMoves.value.clear()
  }
}

// Remove user from team
async function handleRemoveFromTeam(userId: string, teamId: string) {
  const key = `${userId}-${teamId}`
  optimisticUpdates.value.set(key, { userId, teamId, isAdding: false })

  try {
    const result = await removeTeamMembers({
      teamId,
      userIds: [userId],
    })

    if (result.error) {
      toast.add({
        title: t('admin.units.errors.removeMemberFailed'),
        description: result.error.message,
        color: 'error',
      })
      return
    }

    await refetch({ requestPolicy: 'network-only' })
  } finally {
    // Always clear ALL optimistic state after refetch
    optimisticUpdates.value.clear()
  }
}

// Assign team lead
async function handleAssignLeader(userId: string, teamId: string) {
  const result = await assignTeamLead({ teamId, userId })

  if (result.error) {
    toast.add({
      title: t('admin.units.errors.assignLeaderFailed'),
      description: result.error.message,
      color: 'error',
    })
    return
  }

  await refetch({ requestPolicy: 'network-only' })
}

// Rename unit
async function handleRenameUnit(unitId: string, newName: string) {
  const result = await updateTeam({ id: unitId, input: { name: newName } })

  if (result.error) {
    toast.add({
      title: t('admin.units.errors.renameFailed'),
      description: result.error.message,
      color: 'error',
    })
    return false
  }

  await refetch({ requestPolicy: 'network-only' })
  return true
}

// Handle user selection from autocomplete or drag-drop
function handleUserSelect(
  user: {
    id: string
    name: string
    age?: number | null
    gender: string
    teams: { id: string; name: string }[]
  },
  targetTeamId: string,
) {
  // Check if user is already in this team
  const currentProjectTeam = user.teams.find((t) =>
    projectTeamIds.value.has(t.id),
  )
  if (currentProjectTeam?.id === targetTeamId) {
    // Already in this team, do nothing
    return
  }

  // Move directly without confirmation (optimistic UI handles the visual feedback)
  handleAddToTeam(user, targetTeamId)
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
      title: t('admin.units.errors.deleteFailed'),
      description: result.error.message,
      color: 'error',
    })
  } else {
    toast.add({
      title: t('admin.units.success.deleted'),
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

// Bulk delete handlers
function handleBulkDelete() {
  if (selectedUnitIds.value.size === 0) return
  bulkDeleteConfirmOpen.value = true
}

async function confirmBulkDelete() {
  if (selectedUnitIds.value.size === 0) return

  bulkDeletingLoading.value = true

  const results = await Promise.allSettled(
    [...selectedUnitIds.value].map((unitId) => deleteTeam({ id: unitId })),
  )
  const successCount = results.filter(
    (r) => r.status === 'fulfilled' && !r.value.error,
  ).length

  bulkDeletingLoading.value = false
  bulkDeleteConfirmOpen.value = false

  if (successCount > 0) {
    toast.add({
      title: t('admin.units.success.bulkDeleted', { count: successCount }),
      color: 'success',
    })
    clearSelection()
    await refetch({ requestPolicy: 'network-only' })
  } else {
    toast.add({
      title: t('admin.units.errors.bulkDeleteFailed'),
      color: 'error',
    })
  }
}

function cancelBulkDelete() {
  bulkDeleteConfirmOpen.value = false
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
  const members: {
    id: string
    name: string
    age: number | null
    gender: string
    isTeamLead: boolean
    teams: { id: string; name: string }[]
    isRemoving: boolean
    isAdding: boolean
  }[] = []

  // Add existing members (excluding those optimistically moved away)
  for (const member of team.members) {
    const move = optimisticMoves.value.get(member.user.id)
    // Skip if user is being moved away from this team
    if (move && move.fromTeamId === team.id) {
      continue
    }

    const key = `${member.user.id}-${team.id}`
    const update = optimisticUpdates.value.get(key)
    const fullUser = allUsers.value.find((u) => u.id === member.user.id)
    members.push({
      id: member.user.id,
      name: member.name,
      age: member.user.age ?? null,
      gender: member.user.gender,
      isTeamLead: member.isTeamLead,
      teams: fullUser?.teams ?? [],
      isRemoving: update ? !update.isAdding : false,
      isAdding: false,
    })
  }

  // Add users optimistically moved to this team
  for (const move of optimisticMoves.value.values()) {
    if (move.toTeamId === team.id) {
      // Check if not already in members (could happen if server responded before we cleared)
      if (!members.some((m) => m.id === move.userId)) {
        members.push({
          id: move.userId,
          name: move.userName,
          age: move.userAge,
          gender: move.userGender,
          isTeamLead: false,
          teams: move.userTeams,
          isRemoving: false,
          isAdding: true,
        })
      }
    }
  }

  return members
}

// Check if a team has any pending optimistic operations
function isTeamLoading(teamId: string) {
  for (const move of optimisticMoves.value.values()) {
    if (move.toTeamId === teamId || move.fromTeamId === teamId) {
      return true
    }
  }
  return false
}

// Handle drop from draggable
function handleDropMember(
  user: {
    id: string
    name: string
    age?: number | null
    gender: string
    teams: { id: string; name: string }[]
  },
  teamId: string,
) {
  handleUserSelect(user, teamId)
}
</script>

<template>
  <div>
    <div class="border-default border-b py-2">
      <UContainer>
        <UBreadcrumb
          :items="[
            {
              label: $t('admin.breadcrumb.home'),
              to: { name: 'admin-my-church' },
            },
            {
              label: $t('admin.units.title'),
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
        {{ $t('admin.common.back') }}
      </UButton>

      <LoadingState v-if="fetching && !hasLoadedOnce" />
      <ErrorState v-else-if="error" :error />
      <div v-else-if="data" class="mt-12">
        <!-- No active project message -->
        <div
          v-if="!data.myCurrentProject"
          class="text-dimmed text-center py-12"
        >
          <Icon name="lucide:calendar-x" class="size-12 mb-4 mx-auto" />
          <p class="text-lg">{{ $t('admin.units.noActiveProject') }}</p>
        </div>

        <!-- Two-column layout -->
        <div v-else class="grid grid-cols-1 lg:grid-cols-2 gap-8">
          <!-- Left column: Units -->
          <div class="relative">
            <h2 class="text-2xl font-semibold mb-4">
              {{ $t('admin.units.title') }}
            </h2>

            <!-- Search -->
            <UInput
              v-model="unitSearch"
              :placeholder="$t('admin.units.searchPlaceholder')"
              icon="lucide:search"
              class="mb-4"
            />

            <!-- Expand all checkbox -->
            <UCheckbox
              v-model="expandAll"
              :label="$t('admin.units.expandAll')"
              class="mb-4"
            />

            <!-- Create unit buttons -->
            <div v-if="!isCreatingUnit" class="grid grid-cols-2 gap-2 mb-4">
              <UButton size="lg" block class="flex-1" @click="startCreateUnit">
                <Icon name="lucide:plus" />
                {{ $t('admin.units.createOneUnit') }}
              </UButton>
              <UPopover v-model:open="isBulkCreating">
                <UButton size="lg" block>
                  <Icon name="lucide:plus" />
                  {{ $t('admin.units.createMultipleUnits') }}
                </UButton>
                <template #content>
                  <div class="p-4 w-64">
                    <p class="text-sm font-medium mb-3">
                      {{ $t('admin.units.bulkCreate.countLabel') }}
                    </p>
                    <UInput
                      v-model.number="bulkCount"
                      type="number"
                      :min="1"
                      :max="50"
                      class="mb-3"
                    />
                    <p class="text-xs text-dimmed mb-3">
                      {{ $t('admin.units.bulkCreate.autoGenerateInfo') }}
                    </p>
                    <div class="flex gap-2 justify-end">
                      <UButton
                        variant="ghost"
                        size="sm"
                        @click="isBulkCreating = false"
                      >
                        {{ $t('admin.common.cancel') }}
                      </UButton>
                      <UButton
                        size="sm"
                        :loading="bulkCreatingLoading"
                        :disabled="bulkCount < 1 || bulkCount > 50"
                        @click="saveBulkUnits"
                      >
                        {{ $t('admin.common.create') }}
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
                :placeholder="$t('admin.units.namePlaceholder')"
                class="mb-3"
                autofocus
              />
              <div class="flex gap-2 justify-end">
                <UButton variant="ghost" size="sm" @click="cancelCreateUnit">
                  {{ $t('admin.common.cancel') }}
                </UButton>
                <UButton
                  size="sm"
                  :loading="creatingUnitLoading"
                  @click="saveNewUnit"
                >
                  {{ $t('admin.common.save') }}
                </UButton>
              </div>
            </div>

            <!-- Select all checkbox -->
            <div
              v-if="filteredUnits.length > 0"
              class="mb-2 flex items-center gap-2"
            >
              <UCheckbox
                :model-value="allSelected"
                @update:model-value="toggleSelectAll"
              />
              <span class="text-sm text-dimmed">
                {{ $t('admin.units.selectAll') }}
              </span>
            </div>

            <!-- Units list -->
            <TransitionGroup
              tag="div"
              class="space-y-2"
              enter-active-class="transition duration-300 ease-out"
              enter-from-class="scale-95 opacity-0"
              enter-to-class="scale-100 opacity-100"
              leave-active-class="transition duration-300 ease-out absolute left-0 right-0"
              leave-from-class="scale-100 opacity-100"
              leave-to-class="scale-95 opacity-0"
              move-class="transition duration-300 ease-out"
            >
              <div
                v-for="unit in filteredUnits"
                :key="unit.id"
                class="flex items-start gap-2"
              >
                <UCheckbox
                  :model-value="selectedUnitIds.has(unit.id)"
                  class="mt-3.5"
                  @update:model-value="toggleUnitSelection(unit.id)"
                />
                <AdminUnitCard
                  class="flex-1"
                  :unit="unit"
                  :members="getTeamMembers(unit)"
                  :user-items="userItems"
                  :expand-all="expandAll"
                  :loading="isTeamLoading(unit.id)"
                  @add-member="
                    (
                      _userId: string,
                      _teamId: string,
                      _teamName: string,
                      user: {
                        id: string
                        name: string
                        age?: number | null
                        gender: string
                        teams: { id: string; name: string }[]
                      },
                    ) => handleUserSelect(user, unit.id)
                  "
                  @remove-member="handleRemoveFromTeam"
                  @delete-unit="handleDeleteUnit"
                  @drop-member="handleDropMember"
                  @assign-leader="handleAssignLeader"
                  @rename-unit="handleRenameUnit"
                />
              </div>

              <p v-if="filteredUnits.length === 0" class="text-dimmed text-sm">
                {{ $t('admin.units.noUnitsFound') }}
              </p>
            </TransitionGroup>
          </div>

          <!-- Right column: Personer -->
          <div>
            <h2 class="text-2xl font-semibold mb-4 flex gap-3 items-center">
              {{ $t('admin.units.people') }}
              <Icon v-if="fetching" name="svg-spinners:bars-rotate-fade" />
            </h2>

            <!-- Search -->
            <div class="flex gap-4 justify-between items-center mb-4">
              <UInput
                v-model="personSearch"
                :placeholder="$t('admin.units.searchPeoplePlaceholder')"
                icon="lucide:search"
              />
              <UPopover :content="{ align: 'end' }">
                <UChip color="neutral" size="lg" :show="filters.showO36">
                  <UButton variant="outline" color="neutral">
                    <Icon name="lucide:filter" />
                    {{ $t('admin.units.filter.title') }}
                  </UButton>
                </UChip>
                <template #content>
                  <div class="p-4">
                    <UCheckbox
                      v-model="filters.showO36"
                      :label="$t('admin.units.filter.showO36')"
                      :description="$t('admin.units.filter.showO36Description')"
                      :ui="{ description: 'max-w-[200px]' }"
                    />
                  </div>
                </template>
              </UPopover>
            </div>

            <!-- Filter tabs -->
            <UTabs
              v-model="activeFilter"
              :items="filterTabs"
              variant="pill"
              color="neutral"
            />

            <!-- People list (draggable) -->
            <VueDraggable
              :model-value="filteredPeople"
              :group="{ name: 'users', pull: 'clone', put: false }"
              ghost-class="opacity-50"
              :animation="200"
              :sort="false"
              class="mt-4 space-y-1"
            >
              <div
                v-for="person in filteredPeople"
                :key="person.id"
                class="flex items-center justify-between p-2 rounded-lg border border-transparent hover:border-default hover:bg-elevated/50 cursor-grab active:cursor-grabbing"
              >
                <div class="flex items-center gap-2">
                  <Icon
                    name="lucide:grip-vertical"
                    class="size-4 text-dimmed"
                  />
                  <span>{{ person.name }}</span>
                </div>
                <div class="flex items-center gap-1.5 text-dimmed text-sm">
                  <Icon
                    v-if="person.gender === 'MALE'"
                    name="tabler:gender-male"
                    class="size-4 bg-blue-500 rounded-full"
                  />
                  <Icon
                    v-else-if="person.gender === 'FEMALE'"
                    name="tabler:gender-female"
                    class="size-4 bg-pink-500 rounded-full"
                  />
                  <span>
                    {{ $t('admin.units.years', { years: person.age ?? '-' }) }}
                  </span>
                </div>
              </div>
            </VueDraggable>
            <p
              v-if="filteredPeople.length === 0"
              class="text-sm text-dimmed text-center py-4"
            >
              {{ $t('admin.units.noPeopleFound') }}
            </p>
          </div>
        </div>
      </div>
    </UContainer>

    <!-- Floating bulk action bar -->
    <Transition
      enter-active-class="transition duration-200 ease-out"
      enter-from-class="translate-y-full opacity-0"
      enter-to-class="translate-y-0 opacity-100"
      leave-active-class="transition duration-150 ease-in"
      leave-from-class="translate-y-0 opacity-100"
      leave-to-class="translate-y-full opacity-0"
    >
      <div
        v-if="hasSelection"
        class="fixed bottom-6 left-1/2 -translate-x-1/2 z-50 px-4 py-3 rounded-lg border border-default bg-elevated shadow-lg flex items-center gap-4"
      >
        <span class="text-sm">
          {{
            $t(
              'admin.units.selectedCount',
              { count: selectedUnitIds.size },
              selectedUnitIds.size,
            )
          }}
        </span>
        <div class="flex gap-2">
          <UButton variant="ghost" size="sm" @click="clearSelection">
            {{ $t('admin.common.cancel') }}
          </UButton>
          <UButton
            color="error"
            variant="soft"
            size="sm"
            @click="handleBulkDelete"
          >
            <Icon name="lucide:trash-2" />
            {{ $t('admin.units.deleteSelected') }}
          </UButton>
        </div>
      </div>
    </Transition>

    <!-- Delete confirmation modal -->
    <UModal v-model:open="deleteConfirmOpen">
      <template #content>
        <div class="p-6">
          <h3 class="mb-4 text-lg font-semibold">
            {{ $t('admin.units.confirmDelete.title') }}
          </h3>
          <p class="text-dimmed mb-6">
            {{
              $t('admin.units.confirmDelete.message', {
                name: pendingDelete?.unitName,
              })
            }}
          </p>
          <div class="flex justify-end gap-3">
            <UButton variant="ghost" @click="cancelDelete">{{
              $t('admin.common.cancel')
            }}</UButton>
            <UButton
              color="error"
              :loading="deletingLoading"
              @click="confirmDelete"
            >
              {{ $t('admin.common.delete') }}
            </UButton>
          </div>
        </div>
      </template>
    </UModal>

    <!-- Bulk delete confirmation modal -->
    <UModal v-model:open="bulkDeleteConfirmOpen">
      <template #content>
        <div class="p-6">
          <h3 class="mb-4 text-lg font-semibold">
            {{ $t('admin.units.confirmBulkDelete.title') }}
          </h3>
          <p class="text-dimmed mb-6">
            {{
              $t('admin.units.confirmBulkDelete.message', {
                count: selectedUnitIds.size,
              })
            }}
          </p>
          <div class="flex justify-end gap-3">
            <UButton variant="ghost" @click="cancelBulkDelete">
              {{ $t('admin.common.cancel') }}
            </UButton>
            <UButton
              color="error"
              :loading="bulkDeletingLoading"
              @click="confirmBulkDelete"
            >
              {{ $t('admin.common.deleteAll') }}
            </UButton>
          </div>
        </div>
      </template>
    </UModal>
  </div>
</template>
