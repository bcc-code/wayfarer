<script setup lang="ts">
import type {
  PredefinedQuestionData,
  PredefinedResponseData,
  QuestionResult,
  QuizActionState,
  QuizActionHandlers,
} from '../types'
import type { QuizSessionState } from '~/api/generated'

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

const session = useQuizSessionAnswer({
  submissionId: () => props.submissionId,
  questionId: () => props.question.id,
  sessionState: () => props.sessionState,
  readonly: () => props.readonly,
  existingResponseId: () => props.existingResponse?.id,
  buildAnswer: () => ({
    selectedAnswerIds: selectedAnswer.value ? [selectedAnswer.value] : [],
  }),
  betAmount: () => props.betAmount ?? undefined,
})
const {
  actionMode,
  isBetSaved,
  isEditing,
  isSubmitting,
  isSessionLocked,
  isSessionFinished,
} = session

/**
 * Whether the answer is settled enough to reveal correctness against. Inside a
 * session it stays open to revision until the session locks, so a saved answer
 * is not "confirmed" the way it is in a one-shot quiz.
 */
const isAnswerConfirmed = computed(() => {
  if (props.readonly) return true
  if (actionMode.value === 'session-betting') return false
  if (actionMode.value === 'normal') return lockedOutsideSession.value
  return true
})
const lockedOutsideSession = ref(Boolean(props.existingResponse))

/** Whether the answer can still be edited right now. */
const isAnswerLocked = computed(() => {
  if (props.readonly) return true
  if (actionMode.value === 'session-betting')
    return isBetSaved.value && !isEditing.value
  return isAnswerConfirmed.value
})

const submittedResult = ref<{ isCorrect: boolean | null } | null>(
  props.existingResponse
    ? { isCorrect: props.existingResponse.isCorrect ?? null }
    : null,
)

async function handleSaveAnswer() {
  if (!selectedAnswer.value) return
  await session.saveAnswer()
  track(AnalyticsEvent.QuizAnswerSubmitted, {
    question_id: props.question.id,
    is_correct: null, // Hidden while the session is OPEN
    current_question: props.currentIndex + 1,
    total_questions: props.totalQuestions,
    action: session.savedResponseId.value ? 'update_bet' : 'save_bet',
  })
}

// Legacy behavior for non-session quizzes
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
    const isCorrect = response.isCorrect ?? null

    submittedResult.value = { isCorrect }
    lockedOutsideSession.value = true

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

// Action state for parent component
const actionState = computed<QuizActionState>(() => ({
  mode: actionMode.value,
  canSubmit: selectedAnswer.value !== undefined,
  isSubmitting: isSubmitting.value,
  isAnswerLocked: isAnswerConfirmed.value,
  isBetSaved: isBetSaved.value,
  canChangeBet: true,
  isEditing: isEditing.value,
  showPreviousButton: props.showPreviousButton ?? false,
  isLastQuestion: props.isLastQuestion ?? false,
}))

// Handlers for parent component
const handlers: QuizActionHandlers = {
  submit: async () => {
    if (actionMode.value === 'session-betting') {
      await handleSaveAnswer()
    } else {
      await handleLockAnswer()
    }
  },
  continue: handleContinue,
  changeBet: session.changeAnswer,
  previous: handlePrevious,
  next: handleNext,
}

defineExpose({ actionState, handlers })
</script>

<template>
  <div class="text-center p-default flex flex-col gap-default grow">
    <template
      v-for="(alternative, index) in question.predefinedAnswers"
      :key="alternative.id"
    >
      <QuizAlternative
        :text="alternative.answerText"
        :label="quizAnswerLabel(index, question.answerLabelStyle)"
        :description="alternative.description"
        :highlighted="selectedAnswer === alternative.id"
        :confirmed="isAnswerConfirmed"
        :correct="shouldShowCorrect(alternative)"
        :wrong="shouldShowWrong(alternative)"
        :selected="selectedAnswer === alternative.id"
        :disabled="isAnswerLocked || isSessionLocked || isSessionFinished"
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
