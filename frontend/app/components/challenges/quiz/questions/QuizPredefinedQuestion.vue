<script setup lang="ts">
import type {
  PredefinedQuestionData,
  PredefinedResponseData,
  QuestionResult,
  QuizActionState,
  QuizActionHandlers,
} from '../types'
import { QuizSessionState } from '~/api/generated'

const props = defineProps<{
  question: PredefinedQuestionData
  totalQuestions: number
  currentIndex: number
  submissionId: string
  // Controls whether to show correct/incorrect after answering
  revealCorrectAnswers?: boolean
  // Session-based betting props
  sessionState?: QuizSessionState
  // Existing response (for resuming quiz)
  existingResponse?: PredefinedResponseData
  // Review mode props
  readonly?: boolean
  preSelectedAnswerIds?: string[]
  showCorrectAnswers?: boolean
  showPreviousButton?: boolean
  isLastQuestion?: boolean
  // Betting
  betAmount?: number
}>()

const emit = defineEmits<{
  answerSubmitted: [result: QuestionResult]
  previous: []
  next: []
}>()

const { track } = useAnalytics()
const { executeMutation: submitAnswer } = useSubmitQuizAnswerMutation()

// Pre-populate from existing response (resuming quiz) or review mode props
const selectedAnswer = ref<string | undefined>(
  props.existingResponse?.selectedAnswers[0]?.id ??
    props.preSelectedAnswerIds?.[0],
)
const isAnswerConfirmed = ref(props.readonly || Boolean(props.existingResponse))

const isSessionLocked = computed(
  () => props.sessionState === QuizSessionState.Locked,
)
const isSessionFinished = computed(
  () => props.sessionState === QuizSessionState.Finished,
)
const isSubmitting = ref(false)
const submittedResult = ref<{ isCorrect: boolean | null } | null>(
  props.existingResponse
    ? { isCorrect: props.existingResponse.isCorrect ?? null }
    : null,
)

async function handleLockAnswer() {
  if (!selectedAnswer.value || isSubmitting.value) return

  isSubmitting.value = true

  const result = await submitAnswer({
    submissionId: props.submissionId,
    input: {
      questionId: props.question.id,
      selectedAnswerIds: [selectedAnswer.value],
      betAmount: props.betAmount ?? undefined,
    },
  })

  if (result.data?.submitQuizAnswer) {
    const response = result.data.submitQuizAnswer
    const isCorrect =
      response.__typename === 'PredefinedResponse'
        ? (response.isCorrect ?? null)
        : null

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

const { t } = useI18n()
const continueText = computed(() => {
  if (props.currentIndex === props.totalQuestions - 1) {
    return t('quiz.continue')
  }
  return t('quiz.nextQuestion')
})

// In readonly mode, we use isLastQuestion prop to determine the next button text
const nextButtonText = computed(() => {
  if (props.readonly) {
    return props.isLastQuestion
      ? t('quiz.finishReview')
      : t('quiz.nextQuestion')
  }
  return continueText.value
})

function handlePrevious() {
  emit('previous')
}

function handleNext() {
  emit('next')
}

// Determine if we should show correct/wrong highlighting
// In normal mode: only show if revealCorrectAnswers is true (defaults to true)
// In readonly mode: only show if showCorrectAnswers is true
function shouldShowCorrect(alternative: { isCorrect?: boolean | null }) {
  if (!isAnswerConfirmed.value) return false
  if (props.readonly && !props.showCorrectAnswers) return false
  if (!props.readonly && props.revealCorrectAnswers === false) return false
  return alternative.isCorrect === true
}

function shouldShowWrong(alternative: {
  id: string
  isCorrect?: boolean | null
}) {
  if (!isAnswerConfirmed.value) return false
  if (props.readonly && !props.showCorrectAnswers) return false
  if (!props.readonly && props.revealCorrectAnswers === false) return false
  return (
    selectedAnswer.value === alternative.id && alternative.isCorrect === false
  )
}

const isSessionBettingMode = computed(
  () => props.sessionState === QuizSessionState.Open,
)

const actionMode = computed(() => {
  if (props.readonly) return 'review' as const
  if (isSessionFinished.value) return 'session-results' as const
  if (isSessionLocked.value) return 'session-locked' as const
  if (isSessionBettingMode.value) return 'session-betting' as const
  return 'normal' as const
})

// Action state for parent component
const actionState = computed<QuizActionState>(() => ({
  mode: actionMode.value,
  canSubmit: selectedAnswer.value !== undefined,
  isSubmitting: isSubmitting.value,
  isAnswerLocked: isAnswerConfirmed.value,
  isBetSaved: isAnswerConfirmed.value,
  canChangeBet: false,
  isEditing: false,
  showPreviousButton: props.showPreviousButton ?? false,
  isLastQuestion: props.isLastQuestion ?? false,
}))

// Handlers for parent component
const handlers: QuizActionHandlers = {
  submit: handleLockAnswer,
  continue: handleContinue,
  changeBet: () => {},
  previous: handlePrevious,
  next: handleNext,
}

defineExpose({ actionState, handlers })
</script>

<template>
  <div class="text-center p-default flex flex-col gap-default grow">
    <template
      v-for="alternative in question.predefinedAnswers"
      :key="alternative.id"
    >
      <QuizAlternative
        :text="alternative.answerText"
        :highlighted="selectedAnswer === alternative.id"
        :confirmed="isAnswerConfirmed"
        :correct="shouldShowCorrect(alternative)"
        :wrong="shouldShowWrong(alternative)"
        :selected="selectedAnswer === alternative.id"
        :disabled="isAnswerConfirmed || isSessionLocked || isSessionFinished"
        @click="
          !isAnswerConfirmed &&
          !isSessionLocked &&
          !isSessionFinished &&
          (selectedAnswer = alternative.id)
        "
      />
    </template>
  </div>
</template>
