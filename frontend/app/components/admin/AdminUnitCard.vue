<script setup lang="ts">
import { VueDraggable, type SortableEvent } from 'vue-draggable-plus'

interface Member {
  id: string
  name: string
  age?: number | null
  isRemoving?: boolean
}

interface UserItem {
  id: string
  label: string
  user: {
    id: string
    name: string
    teams: { id: string; name: string }[]
  }
}

interface DraggableUser {
  id: string
  name: string
  age?: number | null
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

// Handle when a user is dropped into this unit
function handleDrop(event: SortableEvent) {
  // Access the Vue data from the dragged element
  const item = event.item as HTMLElement & { __draggable_context?: { element: DraggableUser } }
  const user = item.__draggable_context?.element
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
      <span class="font-medium">{{ unit.name }}</span>
      <div class="flex items-center gap-2">
        <UBadge variant="soft" size="sm">
          {{ unit.members.length }} pers
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
          :class="{ 'opacity-50': member.isRemoving }"
        >
          <div class="flex items-center gap-2">
            <Icon name="lucide:grip-vertical" class="size-4 text-dimmed" />
            <Icon
              v-if="member.isRemoving"
              name="lucide:loader-2"
              class="size-4 animate-spin"
            />
            <Icon v-else name="lucide:user" class="size-4" />
            <span>{{ member.name }}</span>
          </div>
          <div class="flex items-center gap-3">
            <span class="text-dimmed text-sm">
              {{ member.age ?? '-' }}
            </span>
            <UButton
              v-if="!member.isRemoving"
              icon="lucide:x"
              size="xs"
              variant="ghost"
              color="neutral"
              @click.stop="emit('removeMember', member.id, unit.id)"
            />
          </div>
        </div>
      </VueDraggable>
      <p v-if="unit.members.length === 0" class="text-dimmed text-sm py-2">
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
