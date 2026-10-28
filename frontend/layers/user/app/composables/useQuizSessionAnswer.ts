import type { QuizSessionState } from '~/api/generated'
import type { QuizActionMode } from '../components/challenges/quiz/types'

/**
 * The parts of an answer that differ per question type, as the shape the
 * submit/update mutations expect (`selectedAnswerIds`, `submittedOrder`, …).
 */
export type QuizAnswerInput = Record<string, unknown>

export interface QuizSessionAnswerOptions {
  submissionId: () => string
  questionId: () => string
  sessionState: () => QuizSessionState | undefined
  readonly: () => boolean | undefined
  /** The response this question already has, if the user answered it before. */
  existingResponseId: () => string | undefined
  /** The type-specific answer fields to send. */
  buildAnswer: () => QuizAnswerInput
  betAmount: () => number | undefined
  onSaved?: (responseId: string) => void
}

/**
 * The session half of answering a question: which action mode the footer is in,
 * and saving or revising the answer and its bet.
 *
 * Every question type shares this, so a bet behaves the same whatever is being
 * bet on. The rule the modes encode: while the session is OPEN the answer and
 * the bet can both be changed as often as the user likes; locking the session
 * freezes them.
 */
export function useQuizSessionAnswer(options: QuizSessionAnswerOptions) {
  const { executeMutation: submitAnswer } = useSubmitQuizAnswerMutation()
  const { executeMutation: updateAnswer } = useUpdateQuizAnswerMutation()

  const savedResponseId = ref(options.existingResponseId())
  const isBetSaved = ref(Boolean(options.existingResponseId()))
  const isEditing = ref(false)
  const isSubmitting = ref(false)

  const isSessionOpen = computed(() => options.sessionState() === 'OPEN')
  const isSessionLocked = computed(() => options.sessionState() === 'LOCKED')
  const isSessionFinished = computed(
    () => options.sessionState() === 'FINISHED',
  )

  const actionMode = computed<QuizActionMode>(() => {
    if (options.readonly()) return 'review'
    if (isSessionFinished.value) return 'session-results'
    if (isSessionLocked.value) return 'session-locked'
    if (isSessionOpen.value) return 'session-betting'
    return 'normal'
  })

  async function saveAnswer() {
    if (isSubmitting.value) return
    isSubmitting.value = true

    try {
      const answer = options.buildAnswer()
      const betAmount = options.betAmount()
      const existingId = savedResponseId.value

      if (existingId) {
        const result = await updateAnswer({
          responseId: existingId,
          input: { ...answer, betAmount },
        })
        if (!result.data?.updateQuizAnswer) return
      } else {
        const result = await submitAnswer({
          submissionId: options.submissionId(),
          input: { questionId: options.questionId(), ...answer, betAmount },
        })
        const response = result.data?.submitQuizAnswer
        if (!response) return
        savedResponseId.value = response.id
      }

      isBetSaved.value = true
      isEditing.value = false
      if (savedResponseId.value) options.onSaved?.(savedResponseId.value)
    } finally {
      isSubmitting.value = false
    }
  }

  function changeAnswer() {
    isEditing.value = true
  }

  return {
    actionMode,
    savedResponseId,
    isBetSaved,
    isEditing,
    isSubmitting,
    isSessionOpen,
    isSessionLocked,
    isSessionFinished,
    saveAnswer,
    changeAnswer,
  }
}
