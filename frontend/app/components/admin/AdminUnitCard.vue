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
}>()

const isOpen = ref(false)

// When expandAll changes to false, collapse this card
watch(
  () => props.expandAll,
  (newValue) => {
    if (!newValue) {
      isOpen.value = false
    }
  },
)

const isExpanded = computed(() => props.expandAll || isOpen.value)

const searchValue = ref<UserItem | undefined>()

function toggle() {
  isOpen.value = !isOpen.value
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
  <div class="rounded-lg border border-default bg-elevated/50 overflow-hidden">
    <!-- Header -->
    <div
      class="w-full p-3 flex items-center justify-between hover:bg-elevated transition-colors cursor-pointer"
      @click="toggle"
    >
      <div class="flex items-center gap-2">
        <span class="font-medium">{{ unit.name }}</span>
        <UTooltip
          v-if="!hasLeader && !loading"
          text="Ingen leder valgt"
          :delay-duration="200"
        >
          <Icon name="lucide:alert-triangle" class="size-4 text-warning" />
        </UTooltip>
        <Icon
          v-if="loading"
          name="svg-spinners:bars-rotate-fade"
          class="size-4 text-dimmed"
        />
      </div>
      <div class="flex items-center gap-2">
        <UBadge variant="soft" size="sm">
          {{ unit.members.length }} personer
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
          class="flex items-center justify-between py-2 border-b border-default last:border-0 cursor-grab active:cursor-grabbing"
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
            <UTooltip v-else text="Gjør til leder" :delay-duration="200">
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
              <span>{{ member.age ?? '-' }} år</span>
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
        Ingen medlemmer - dra en person hit
      </p>

      <!-- Add person search -->
      <UInputMenu
        v-model="searchValue"
        :items="userItems"
        placeholder="Legg til person..."
        icon="lucide:search"
        class="w-full"
        virtualize
        :ui="{ base: 'cursor-pointer' }"
        @update:model-value="handleUserSelect"
      />
    </div>
  </div>
</template>
