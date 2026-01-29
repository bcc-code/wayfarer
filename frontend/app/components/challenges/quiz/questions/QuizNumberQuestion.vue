<script setup lang="ts">
import type {
  NumberQuestionData,
  QuestionResult,
  QuizActionHandlers,
  QuizActionState,
} from '../types'

const props = defineProps<{
  question: NumberQuestionData
  totalQuestions: number
  currentIndex: number
  submissionId: string
  // Review mode props
  readonly?: boolean
  preSelectedAnswer?: number
  showCorrectAnswers?: boolean
  showPreviousButton?: boolean
  isLastQuestion?: boolean
}>()

const emit = defineEmits<{
  answerSubmitted: [result: QuestionResult]
  previous: []
  next: []
}>()

const { track } = useAnalytics()
const { executeMutation: submitAnswer } = useSubmitQuizAnswerMutation()

const currentValue = ref(0)
const isSubmitting = ref(false)
const isAnswerConfirmed = ref(false)
const submittedResult = ref<{ isCorrect: boolean | null } | null>(null)

async function handleLockAnswer() {
  if (!currentValue.value || isSubmitting.value) return

  isSubmitting.value = true

  const result = await submitAnswer({
    submissionId: props.submissionId,
    input: {
      questionId: props.question.id,
      numberResponse: currentValue.value,
    },
  })

  if (result.data?.submitQuizAnswer) {
    const response = result.data.submitQuizAnswer
    const numberResponse =
      response.__typename === 'NumberResponse'
        ? (response.numberResponse ?? null)
        : null
    const isCorrect = numberResponse !== null // TODO: add the possibility to have a correct number answer

    submittedResult.value = { isCorrect }
    isAnswerConfirmed.value = true

    track(AnalyticsEvent.QuizAnswerSubmitted, {
      question_id: props.question.id,
      is_correct: isCorrect,
      current_question: props.currentIndex + 1,
      total_questions: props.totalQuestions,
    })
  }

  isSubmitting.value = false
}

function handleContinue() {
  emit('answerSubmitted', {
    questionId: props.question.id,
    isCorrect: submittedResult.value?.isCorrect ?? null,
  })
}

// Action state for parent component
const actionState = computed<QuizActionState>(() => ({
  mode: props.readonly ? 'review' : 'normal',
  canSubmit: currentValue.value !== undefined,
  isSubmitting: isSubmitting.value,
  isAnswerLocked: isAnswerConfirmed.value,
  isBetSaved: false,
  isEditing: false,
  showPreviousButton: props.showPreviousButton ?? false,
  isLastQuestion: props.isLastQuestion ?? false,
}))

// Handlers for parent component
const handlers: QuizActionHandlers = {
  submit: handleLockAnswer,
  continue: handleContinue,
  changeBet: () => {},
  previous: () => {
    emit('previous')
  },
  next: () => {
    emit('next')
  },
}

defineExpose({ actionState, handlers })
</script>

<template>
  <div class="grow p-default text-center flex flex-col gap-default">
    <div class="mt-18 mb-10">
      <p class="text-caption text-accent-contrast">
        {{ $t('quiz.yourNumber') }}
      </p>
      <p class="text-hero text-text-default tabular-nums">
        {{ formatNumber(currentValue) }}
      </p>
    </div>
    <DesignSlider
      v-model="currentValue"
      :min="question.minValue ?? 0"
      :max="question.maxValue ?? 100"
      :step="question.stepValue ?? 1"
    />
  </div>
</template>
