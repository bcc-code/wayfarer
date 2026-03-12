import type { QuizSessionState } from '~/api/generated'
import type { QuizActionMode, QuizActionState } from '~/components/challenges/quiz/types'

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

// --- Footer resolution ---

export type FooterBettingModule =
  | { visible: false }
  | { visible: true; mode: 'betting'; disabled: boolean }
  | { visible: true; mode: 'locked'; disabled: boolean }
  | { visible: true; mode: 'results' }

export type FooterButton =
  | { type: 'lock-answer'; disabled: boolean }
  | { type: 'continue' }
  | { type: 'save-answer'; disabled: boolean }
  | { type: 'change-answer'; secondaryAction: 'next' | 'finalize' | 'none' }
  | { type: 'finalize' }
  | { type: 'close' }
  | { type: 'done' }
  | { type: 'review-nav' }

export interface FooterState {
  bettingModule: FooterBettingModule
  button: FooterButton
}

export interface FooterInput {
  actionState: QuizActionState
  isBettingEnabled: boolean
  isSingleQuestion: boolean
}

/**
 * Resolves which footer elements to show based on action state and betting.
 * Maps the template v-if/v-else-if chain to a data structure for testing.
 */
export function resolveFooterState(input: FooterInput): FooterState {
  const { actionState, isBettingEnabled } = input
  const { mode } = actionState

  switch (mode) {
    case 'normal':
      return {
        bettingModule: { visible: false },
        button: actionState.isAnswerLocked
          ? { type: 'continue' }
          : {
              type: 'lock-answer',
              disabled: !actionState.canSubmit || actionState.isSubmitting,
            },
      }

    case 'session-betting': {
      const showBettingModule =
        isBettingEnabled && !actionState.isAnswerLocked
      let button: FooterButton
      if (!actionState.isBetSaved || actionState.isEditing) {
        button = {
          type: 'save-answer',
          disabled: !actionState.canSubmit || actionState.isSubmitting,
        }
      } else if (actionState.canChangeBet) {
        const secondaryAction = !actionState.isLastQuestion
          ? 'next'
          : input.isSingleQuestion
            ? 'none'
            : 'finalize'
        button = { type: 'change-answer', secondaryAction }
      } else if (!actionState.isLastQuestion) {
        button = { type: 'continue' }
      } else {
        button = { type: 'finalize' }
      }
      return {
        bettingModule: showBettingModule
          ? {
              visible: true,
              mode: 'betting',
              disabled: actionState.isBetSaved && !actionState.isEditing,
            }
          : { visible: false },
        button,
      }
    }

    case 'session-locked': {
      const showBettingModule =
        isBettingEnabled && !actionState.isAnswerLocked
      return {
        bettingModule: showBettingModule
          ? {
              visible: true,
              mode: 'locked',
              disabled: actionState.isBetSaved && !actionState.isEditing,
            }
          : { visible: false },
        button: { type: 'close' },
      }
    }

    case 'session-results':
      return {
        bettingModule: isBettingEnabled
          ? { visible: true, mode: 'results' }
          : { visible: false },
        button: { type: 'done' },
      }

    case 'review':
      return {
        bettingModule: { visible: false },
        button: { type: 'review-nav' },
      }
  }
}
