<script setup lang="ts">
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

function toggle() {
  isOpen.value = !isOpen.value
}

function handleUserSelect(item: UserItem | undefined) {
  if (item) {
    emit('addMember', item.user.id, props.unit.id, props.unit.name, item.user)
  }
}
</script>

<template>
  <div class="rounded-lg border border-default bg-elevated/50 overflow-hidden">
    <!-- Header -->
    <button
      class="w-full p-3 flex items-center justify-between hover:bg-elevated transition-colors"
      @click="toggle"
    >
      <span class="font-medium">{{ unit.name }}</span>
      <div class="flex items-center gap-2">
        <UBadge variant="soft" size="sm">
          {{ unit.members.length }} pers
        </UBadge>
        <Icon
          :name="isExpanded ? 'lucide:chevron-up' : 'lucide:chevron-down'"
          class="size-4"
        />
      </div>
    </button>

    <!-- Expanded content -->
    <div v-if="isExpanded" class="border-t border-default p-4">
      <!-- Members list -->
      <div class="mb-4">
        <div
          v-for="member in members"
          :key="member.id"
          class="flex items-center justify-between py-2 border-b border-default last:border-0"
          :class="{ 'opacity-50': member.isRemoving }"
        >
          <div class="flex items-center gap-2">
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
        <p v-if="unit.members.length === 0" class="text-dimmed text-sm py-2">
          Ingen medlemmer
        </p>
      </div>

      <!-- Add person search -->
      <UInputMenu
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
