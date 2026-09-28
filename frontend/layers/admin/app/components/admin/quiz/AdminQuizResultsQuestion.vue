<script setup lang="ts">
import {
  questionResultRows,
  questionTypeLabel,
  toTsv,
  type QuestionResultLike,
} from '../../../utils/quizResultsExport'

const props = defineProps<{
  result: QuestionResultLike
}>()

const toast = useToast()

const typeLabel = computed(() => questionTypeLabel(props.result.__typename))

const copied = ref(false)

/**
 * Tab-separated to the clipboard, because that is what pastes into Keynote and
 * Excel as a table rather than one column of text.
 */
async function copyRows() {
  const rows = questionResultRows(props.result)
  if (!rows.length) return

  try {
    await navigator.clipboard.writeText(toTsv(rows))
    copied.value = true
    setTimeout(() => (copied.value = false), 2000)
  } catch {
    // Clipboard access is denied outside a secure context, and a silent
    // no-op here reads as a broken button.
    toast.add({
      title: 'Kunne ikke kopiere',
      description: 'Nettleseren tillot ikke tilgang til utklippstavlen.',
      color: 'error',
    })
  }
}

const hasRows = computed(() => questionResultRows(props.result).length > 0)
</script>

<template>
  <UCard>
    <div class="mb-4 flex items-start justify-between gap-4">
      <div class="min-w-0">
        <p class="text-dimmed text-xs">
          Spørsmål {{ result.question.questionOrder }}
          <span class="text-dimmed">·</span>
          {{ typeLabel }}
        </p>
        <h3 class="mt-0.5 font-semibold">{{ result.question.questionText }}</h3>
      </div>
      <UButton
        v-if="hasRows"
        variant="ghost"
        size="xs"
        :icon="copied ? 'lucide:check' : 'lucide:copy'"
        class="shrink-0"
        @click="copyRows"
      >
        {{ copied ? 'Kopiert' : 'Kopier' }}
      </UButton>
    </div>

    <p v-if="result.responseCount === 0" class="text-dimmed text-sm">
      Ingen har svart på dette spørsmålet ennå.
    </p>

    <AdminQuizResultsPredefined
      v-else-if="result.__typename === 'PredefinedQuestionResults'"
      :options="result.options ?? []"
      :response-count="result.responseCount"
      :correct-count="result.correctCount ?? 0"
    />
    <AdminQuizResultsNumber
      v-else-if="result.__typename === 'NumberQuestionResults'"
      :buckets="result.buckets ?? []"
      :average="result.average"
      :median="result.median"
      :min="result.min"
      :max="result.max"
    />
    <AdminQuizResultsFreeText
      v-else-if="result.__typename === 'FreeTextQuestionResults'"
      :groups="result.groups ?? []"
      :responses="result.responses ?? []"
      :distinct-count="result.distinctCount ?? 0"
    />
    <AdminQuizResultsOrdering
      v-else-if="result.__typename === 'OrderingQuestionResults'"
      :items="result.items ?? []"
      :response-count="result.responseCount"
      :fully-correct-count="result.fullyCorrectCount ?? 0"
    />
    <AdminQuizResultsJson v-else :response-count="result.responseCount" />
  </UCard>
</template>
