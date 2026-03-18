<script setup lang="ts">
import {
  VueDraggable,
  type DraggableEvent,
  type SortableEvent,
} from 'vue-draggable-plus'

interface Member {
  id: string
  name: string
  age?: number | null
  gender?: string
  isTeamLead?: boolean
  isRemoving?: boolean
  isAdding?: boolean
}

interface UserItem {
  id: string
  label: string
  user: {
    id: string
    name: string
    age?: number | null
    gender: string
    teams: { id: string; name: string }[]
  }
}

interface DraggableUser {
  id: string
  name: string
  age?: number | null
  gender: string
  teams: { id: string; name: string }[]
}

const props = defineProps<{
  unit: {
    id: string
    name: string
    leaderboardExcluded?: boolean
    averageAge?: number | null
    joinCode?: string
    members: {
      id: string
      name: string
      user: { id: string; age?: number | null }
    }[]
  }
  members: Member[]
  userItems: UserItem[]
  expandAll: boolean
  loading?: boolean
}>()

const emit = defineEmits<{
  addMember: [
    userId: string,
    teamId: string,
    teamName: string,
    user: UserItem['user'],
  ]
  removeMember: [userId: string, teamId: string]
  deleteUnit: [unitId: string, unitName: string]
  dropMember: [user: DraggableUser, teamId: string, teamName: string]
  assignLeader: [userId: string, teamId: string]
  renameUnit: [unitId: string, newName: string]
  'update:expanded': [expanded: boolean]
}>()

const isOpen = ref(props.expandAll)

// When expandAll changes, sync the local state
watch(
  () => props.expandAll,
  (newValue) => {
    isOpen.value = newValue
  },
)

const isExpanded = computed(() => isOpen.value)

const searchValue = ref<UserItem | undefined>()

// Inline editing state
const isEditing = ref(false)
const editName = ref('')
const nameInputRef = ref<HTMLInputElement | null>(null)

function toggle() {
  if (isEditing.value) return
  isOpen.value = !isOpen.value
  emit('update:expanded', isOpen.value)
}

function startEditing() {
  editName.value = props.unit.name
  isEditing.value = true
  nextTick(() => {
    nameInputRef.value?.focus()
    nameInputRef.value?.select()
  })
}

function cancelEditing() {
  isEditing.value = false
  editName.value = ''
}

function saveEditing() {
  const trimmedName = editName.value.trim()
  if (trimmedName && trimmedName !== props.unit.name) {
    emit('renameUnit', props.unit.id, trimmedName)
  }
  isEditing.value = false
  editName.value = ''
}

function handleEditKeydown(event: KeyboardEvent) {
  if (event.key === 'Enter') {
    saveEditing()
  } else if (event.key === 'Escape') {
    cancelEditing()
  }
}

function handleUserSelect(item: UserItem | undefined) {
  if (item) {
    emit('addMember', item.user.id, props.unit.id, props.unit.name, item.user)
    nextTick(() => {
      searchValue.value = undefined
    })
  }
}

// Local copy of members for draggable (doesn't mutate, just for display)
const localMembers = computed(() => [...props.members])

const averageAge = computed(() => {
  const mem = localMembers.value.filter((m) => typeof m.age === 'number')
  if (mem.length === 0) {
    return 0
  }
  const sum = mem.reduce((acc, m) => acc + m.age!, 0)

  if (sum / mem.length != props.unit.averageAge) {
    return sum / mem.length
  }
  return props.unit.averageAge
})

// Check if any member is a team lead
const hasLeader = computed(() => props.members.some((m) => m.isTeamLead))

// Handle when a user is dropped into this unit
function handleDrop(event: SortableEvent) {
  const draggableEvent = event as DraggableEvent<DraggableUser>
  const user = draggableEvent.data
  if (user) {
    emit('dropMember', user, props.unit.id, props.unit.name)
  }
}
</script>

<template>
  <div
    class="rounded-xl group border border-default bg-elevated/50 overflow-hidden"
  >
    <!-- Header -->
    <div
      class="w-full p-3 flex items-center justify-between hover:bg-elevated transition-colors cursor-pointer"
      @click="toggle"
    >
      <div class="flex items-center gap-2">
        <input
          v-if="isEditing"
          ref="nameInputRef"
          v-model="editName"
          type="text"
          class="font-medium bg-transparent border border-default rounded px-2 py-0.5 focus:outline-none focus:ring-2 focus:ring-primary"
          @click.stop
          @keydown="handleEditKeydown"
          @blur="saveEditing"
        >
        <button
          v-else
          type="button"
          class="font-medium px-2 py-0.5 rounded border border-transparent group-hover:border-accented hover:bg-elevated/50 transition-colors text-left"
          @click.stop="startEditing"
        >
          {{ unit.name }}
        </button>
        <UBadge
          v-if="unit.members.length && !hasLeader"
          color="warning"
          variant="soft"
          icon="lucide:triangle-alert"
          :label="$t('admin.unit.noLeaderSelected')"
        />
        <UBadge
          v-if="unit.leaderboardExcluded"
          color="warning"
          variant="subtle"
          icon="lucide:eye-off"
          :label="$t('admin.unit.excludedFromLeaderboard')"
        />
      </div>
      <div class="flex items-center gap-2">
        <UBadge variant="soft" color="neutral">
          {{
            $t(
              'admin.unit.members',
              { count: unit.members.length },
              unit.members.length,
            )
          }}
        </UBadge>
        <UBadge v-if="unit.joinCode" variant="subtle" color="neutral">
          {{ unit.joinCode }}
        </UBadge>
        <UButton
          icon="lucide:trash-2"
          size="xs"
          variant="ghost"
          color="error"
          @click.stop="emit('deleteUnit', unit.id, unit.name)"
        />
        <Icon
          :name="isExpanded ? 'lucide:chevron-up' : 'lucide:chevron-down'"
          class="size-4"
        />
      </div>
    </div>

    <!-- Expanded content -->
    <div v-if="isExpanded" class="border-t border-default p-4">
      <UAlert
        v-if="averageAge >= 36"
        color="warning"
        variant="soft"
        icon="lucide:triangle-alert"
        :description="
          $t('admin.units.ageWarning', { email: 'support@bcc.media' })
        "
        class="mb-2"
      />
      <!-- Members list (droppable) -->
      <VueDraggable
        :model-value="localMembers"
        :group="{ name: 'users', pull: 'clone' }"
        ghost-class="opacity-50"
        :animation="200"
        class="mb-4 min-h-10"
        @add="handleDrop"
      >
        <div
          v-for="member in localMembers"
          :key="member.id"
          class="group flex items-center justify-between py-2 border-b border-default last:border-0 cursor-grab active:cursor-grabbing"
          :class="{
            'opacity-50': member.isRemoving || member.isAdding,
          }"
        >
          <div class="flex items-center gap-2">
            <Icon name="lucide:grip-vertical" class="size-4 text-dimmed" />
            <span>{{ member.name }}</span>
            <Icon
              v-if="member.isTeamLead"
              name="lucide:crown"
              class="size-4 m-1 text-yellow-500"
            />
            <UTooltip
              v-else
              :text="$t('admin.unit.makeLead')"
              :delay-duration="200"
              class="invisible group-hover:visible"
            >
              <UButton
                icon="lucide:crown"
                size="xs"
                variant="ghost"
                color="neutral"
                class="opacity-30"
                @click.stop="emit('assignLeader', member.id, unit.id)"
              />
            </UTooltip>
          </div>
          <div class="flex items-center gap-3">
            <div class="flex items-center gap-2 text-dimmed text-sm">
              <Icon
                v-if="member.gender === 'MALE'"
                name="tabler:gender-male"
                class="size-4 bg-blue-500 rounded-full"
              />
              <Icon
                v-else-if="member.gender === 'FEMALE'"
                name="tabler:gender-female"
                class="size-4 bg-pink-500 rounded-full"
              />
              <span>
                {{ $t('admin.units.years', { years: member.age ?? '-' }) }}
              </span>
            </div>
            <Icon
              v-if="member.isRemoving || member.isAdding"
              name="svg-spinners:bars-rotate-fade"
              class="size-4 m-1"
            />
            <template v-else>
              <UButton
                icon="lucide:x"
                size="xs"
                variant="ghost"
                color="neutral"
                @click.stop="emit('removeMember', member.id, unit.id)"
              />
            </template>
          </div>
        </div>
      </VueDraggable>
      <p
        v-if="unit.members.length === 0"
        class="text-dimmed text-sm text-center py-2 -mt-8 mb-8"
      >
        {{ $t('admin.unit.dragPeopleHere') }}
      </p>

      <!-- Add person search -->
      <UInputMenu
        v-model="searchValue"
        :items="userItems"
        :placeholder="$t('admin.unit.addPersonPlaceholder')"
        icon="lucide:search"
        class="w-full"
        virtualize
        :ui="{ base: 'cursor-pointer' }"
        @update:model-value="handleUserSelect"
      />
    </div>
  </div>
</template>
