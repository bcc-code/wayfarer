import type { QuizSessionState } from '~/api/generated'

export type QuizViewState =
  | 'loading'
  | 'just-completed'
  | 'results'
  | 'active-question'
  | 'results-fallback'
  | 'not-submitted'
  | 'unavailable'
  | 'unknown'

export interface QuizViewStateInput {
  isLoading: boolean
  quizCompleted: boolean
  hasFinalResult: boolean
  hasCompletedSubmission: boolean
  canStartQuiz: boolean
  sessionState: QuizSessionState | undefined
  isSingleOrderingQuestion: boolean
  hasCurrentQuestion: boolean
  isNotSubmitted: boolean
  isQuizUnavailable: boolean
}

/**
 * Determines which view to show in the quiz challenge component.
 *
 * The view states, in priority order:
 * 1. loading — waiting for initial data
 * 2. just-completed — user just answered the last question this session
 * 3. results — completed submission exists, show score/review
 * 4. active-question — show the current question (answering, locked, or session results)
 * 5. results-fallback — completed submission with retakes available
 * 6. not-submitted — session ended without user submitting
 * 7. unavailable — no session and can't start
 * 8. unknown — unexpected state
 *
 * Special cases for session states:
 * - LOCKED: always show active-question (user views their locked answers)
 * - FINISHED with single ordering question: show active-question (betting results inline)
 * - FINISHED with multiple questions: show results screen
 */
export function resolveQuizViewState(
  input: QuizViewStateInput,
): QuizViewState {
  if (input.isLoading) return 'loading'

  if (input.quizCompleted && input.hasFinalResult) return 'just-completed'

  if (input.hasCompletedSubmission && !input.canStartQuiz) {
    const shouldShowQuestionView =
      input.sessionState === 'LOCKED' ||
      (input.sessionState === 'FINISHED' && input.isSingleOrderingQuestion)

    if (!shouldShowQuestionView) return 'results'
  }

  if (input.hasCurrentQuestion) return 'active-question'

  if (input.hasCompletedSubmission) return 'results-fallback'

  if (input.isNotSubmitted) return 'not-submitted'

  if (input.isQuizUnavailable) return 'unavailable'

  return 'unknown'
}
