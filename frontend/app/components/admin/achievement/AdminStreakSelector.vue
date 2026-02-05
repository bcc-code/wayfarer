<script setup lang="ts">
const props = defineProps<{
  projectId: string
  streakId?: string
  neededStreak: number
}>()

const emit = defineEmits<{
  'update:streakId': [id: string]
  'update:neededStreak': [days: number]
}>()

const { data, fetching } = useAdminProjectStreaksQuery({
  variables: computed(() => ({
    projectId: props.projectId,
  })),
})

const streakOptions = computed(() => {
  if (!data.value?.streaks?.edges) return []
  return data.value.streaks.edges.map((edge) => ({
    value: edge.node.id,
    label: edge.node.name,
    description: edge.node.description,
  }))
})

const selectedStreak = computed(() => {
  return streakOptions.value.find((s) => s.value === props.streakId)
})
</script>

<template>
  <div class="flex flex-col gap-4">
    <UFormField name="streakId" label="Streak" required>
      <USelect
        :model-value="streakId"
        :items="streakOptions"
        value-key="value"
        placeholder="Select a streak..."
        :loading="fetching"
        class="w-full"
        @update:model-value="(v) => emit('update:streakId', v as string)"
      />
      <template v-if="selectedStreak?.description" #help>
        {{ selectedStreak.description }}
      </template>
    </UFormField>

    <UFormField
      name="neededStreak"
      label="Required Days"
      help="Number of consecutive days user must maintain the streak"
      required
    >
      <UInput
        :model-value="neededStreak"
        type="number"
        min="1"
        required
        class="w-full"
        @update:model-value="(v) => emit('update:neededStreak', Number(v))"
      />
    </UFormField>
  </div>
</template>
