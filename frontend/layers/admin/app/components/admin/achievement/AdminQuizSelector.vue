<script setup lang="ts">
const props = defineProps<{
  projectId: string
  quizId?: string
  minScorePercentage?: number
  requireCompletion: boolean
}>()

const emit = defineEmits<{
  'update:quizId': [id: string]
  'update:minScorePercentage': [percentage: number | undefined]
  'update:requireCompletion': [required: boolean]
}>()

const { data, fetching } = useAdminProjectQuizzesQuery({
  variables: computed(() => ({
    projectId: props.projectId,
  })),
})

const quizOptions = computed(() => {
  if (!data.value?.quizzes?.edges) return []
  return data.value.quizzes.edges.map((edge) => ({
    value: edge.node.id,
    label: edge.node.name,
  }))
})

const useMinScore = ref(props.minScorePercentage !== undefined)

watch(useMinScore, (enabled) => {
  if (!enabled) {
    emit('update:minScorePercentage', undefined)
  } else if (props.minScorePercentage === undefined) {
    emit('update:minScorePercentage', 70)
  }
})
</script>

<template>
  <div class="flex flex-col gap-4">
    <UFormField name="quizId" label="Quiz" required>
      <USelect
        :model-value="quizId"
        :items="quizOptions"
        value-key="value"
        placeholder="Select a quiz..."
        :loading="fetching"
        class="w-full"
        @update:model-value="(v) => emit('update:quizId', v as string)"
      />
    </UFormField>

    <UFormField name="requireCompletion">
      <UCheckbox
        :model-value="requireCompletion"
        label="Require completion"
        help="User must complete the quiz to earn the achievement"
        @update:model-value="
          (v) => emit('update:requireCompletion', v as boolean)
        "
      />
    </UFormField>

    <UFormField name="useMinScore">
      <UCheckbox
        v-model="useMinScore"
        label="Require minimum score"
        help="User must achieve a minimum score percentage"
      />
    </UFormField>

    <UFormField
      v-if="useMinScore"
      name="minScorePercentage"
      label="Minimum Score (%)"
    >
      <div class="flex items-center gap-4">
        <UInput
          :model-value="minScorePercentage ?? 70"
          type="number"
          min="0"
          max="100"
          class="w-24"
          @update:model-value="
            (v) => emit('update:minScorePercentage', Number(v))
          "
        />
        <span class="text-muted text-sm">%</span>
      </div>
    </UFormField>
  </div>
</template>
