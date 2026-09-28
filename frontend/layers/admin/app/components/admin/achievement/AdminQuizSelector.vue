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

const { isAuthReady } = useAuthReady()
const { data, fetching } = useAdminProjectQuizzesQuery({
  variables: computed(() => ({
    projectId: props.projectId,
  })),
  pause: computed(() => !isAuthReady.value),
})

const quizOptions = computed(
  () =>
    data.value?.quizzes?.edges.map((edge) => ({
      value: edge.node.id,
      label: edge.node.name,
    })) ?? [],
)

const hasQuizzes = computed(() => quizOptions.value.length > 0)

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
    <p class="text-muted text-sm">
      Utmerkelsen vurderes når deltakeren leverer quizen. Uten kravene nedenfor
      får alle som leverer den, uansett resultat.
    </p>

    <UFormField name="quizId" label="Quiz" required>
      <USelect
        :model-value="quizId"
        :items="quizOptions"
        value-key="value"
        :placeholder="
          fetching
            ? 'Laster quizer...'
            : hasQuizzes
              ? 'Velg quiz...'
              : 'Ingen quizer i dette prosjektet'
        "
        :loading="fetching"
        :disabled="!hasQuizzes"
        class="w-full"
        @update:model-value="(v) => emit('update:quizId', v as string)"
      />
      <!-- A quiz belongs to a quiz challenge, so there is nothing to pick
           until one exists. -->
      <template v-if="!fetching && !hasQuizzes" #help>
        Opprett en quiz-utfordring i prosjektet først.
      </template>
    </UFormField>

    <!-- `description`, not `help`: `help` is a UFormField prop, so on a
         UCheckbox it was silently dropped and no help text ever rendered. -->
    <UCheckbox
      :model-value="requireCompletion"
      label="Krev fullført quiz"
      description="Utmerkelsen gis bare hvis deltakeren fullfører hele quizen."
      @update:model-value="
        (v) => emit('update:requireCompletion', v as boolean)
      "
    />

    <UCheckbox
      v-model="useMinScore"
      label="Krev minste poengandel"
      description="Utmerkelsen gis bare hvis deltakeren får minst denne andelen av poengene i quizen."
    />

    <UFormField
      v-if="useMinScore"
      name="minScorePercentage"
      label="Minste poengandel"
    >
      <div class="flex items-center gap-2">
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
        <span class="text-muted text-sm">% av poengene i quizen</span>
      </div>
    </UFormField>
  </div>
</template>
