<script setup lang="ts">
import type {
  FreeTextQuestionData,
  FreeTextResponseData,
  QuestionResult,
  QuizActionHandlers,
  QuizActionState,
} from '../types'

const props = defineProps<{
  question: FreeTextQuestionData
  totalQuestions: number
  currentIndex: number
  submissionId: string
  // Controls whether to show correct/incorrect after answering
  revealCorrectAnswers?: boolean
  // Existing response (for resuming quiz)
  existingResponse?: FreeTextResponseData
  // Review mode props
  readonly?: boolean
  preSelectedAnswer?: string
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

// Pre-populate from existing response (resuming quiz) or review mode props
const currentValue = ref(
  props.existingResponse?.textResponse ?? props.preSelectedAnswer ?? '',
)
const isSubmitting = ref(false)
const isAnswerConfirmed = ref(!!props.existingResponse)
const submittedResult = ref<{ isCorrect: boolean | null } | null>(
  props.existingResponse ? { isCorrect: null } : null,
)

async function handleLockAnswer() {
  if (!currentValue.value || isSubmitting.value) return

  isSubmitting.value = true

  const result = await submitAnswer({
    submissionId: props.submissionId,
    input: {
      questionId: props.question.id,
      textResponse: currentValue.value,
    },
  })

  if (result.data?.submitQuizAnswer) {
    const response = result.data.submitQuizAnswer
    const textResponse =
      response.__typename === 'FreeTextResponse'
        ? (response.textResponse ?? null)
        : null
    const isCorrect = textResponse !== null // TODO: add the possibility to have a correct number answer

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
    <DesignTextarea
      v-model="currentValue"
      :placeholder="$t('quiz.writeYourAnswer')"
      :disabled="readonly"
    />
  </div>
</template>
