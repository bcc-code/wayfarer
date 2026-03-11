import { describe, it, expect } from 'vitest'
import {
  resolveQuizViewState,
  resolveFooterState,
  type QuizViewStateInput,
  type FooterInput,
} from '../../app/composables/useQuizViewState'
import type { QuizActionState } from '../../app/components/challenges/quiz/types'

function makeInput(
  overrides: Partial<QuizViewStateInput> = {},
): QuizViewStateInput {
  return {
    isLoading: false,
    quizCompleted: false,
    hasFinalResult: false,
    hasCompletedSubmission: false,
    canStartQuiz: false,
    sessionState: undefined,
    isSingleOrderingQuestion: false,
    hasCurrentQuestion: false,
    isNotSubmitted: false,
    isQuizUnavailable: false,
    ...overrides,
  }
}

function makeActionState(
  overrides: Partial<QuizActionState> = {},
): QuizActionState {
  return {
    mode: 'normal',
    canSubmit: false,
    isSubmitting: false,
    isAnswerLocked: false,
    isBetSaved: false,
    canChangeBet: false,
    isEditing: false,
    showPreviousButton: false,
    isLastQuestion: false,
    ...overrides,
  }
}

function makeFooterInput(overrides: Partial<FooterInput> = {}): FooterInput {
  return {
    actionState: makeActionState(),
    isBettingEnabled: false,
    isSingleQuestion: false,
    ...overrides,
  }
}

describe('resolveQuizViewState', () => {
  describe('loading', () => {
    it('returns loading when isLoading is true', () => {
      expect(resolveQuizViewState(makeInput({ isLoading: true }))).toBe(
        'loading',
      )
    })

    it('loading takes priority over everything else', () => {
      expect(
        resolveQuizViewState(
          makeInput({
            isLoading: true,
            quizCompleted: true,
            hasFinalResult: true,
            hasCompletedSubmission: true,
            hasCurrentQuestion: true,
          }),
        ),
      ).toBe('loading')
    })
  })

  describe('just-completed', () => {
    it('returns just-completed when quiz is completed with final result', () => {
      expect(
        resolveQuizViewState(
          makeInput({ quizCompleted: true, hasFinalResult: true }),
        ),
      ).toBe('just-completed')
    })

    it('does not return just-completed without final result', () => {
      expect(
        resolveQuizViewState(
          makeInput({
            quizCompleted: true,
            hasFinalResult: false,
            hasCurrentQuestion: true,
          }),
        ),
      ).not.toBe('just-completed')
    })
  })

  describe('results', () => {
    it('returns results when completed submission exists and cannot restart', () => {
      expect(
        resolveQuizViewState(
          makeInput({
            hasCompletedSubmission: true,
            canStartQuiz: false,
          }),
        ),
      ).toBe('results')
    })

    it('does not return results when canStartQuiz is true', () => {
      expect(
        resolveQuizViewState(
          makeInput({
            hasCompletedSubmission: true,
            canStartQuiz: true,
            hasCurrentQuestion: true,
          }),
        ),
      ).not.toBe('results')
    })
  })

  describe('session LOCKED — always show active-question', () => {
    it('shows active-question when session is LOCKED with completed submission', () => {
      expect(
        resolveQuizViewState(
          makeInput({
            hasCompletedSubmission: true,
            canStartQuiz: false,
            sessionState: 'LOCKED',
            hasCurrentQuestion: true,
          }),
        ),
      ).toBe('active-question')
    })

    it('shows active-question when LOCKED even with multiple questions', () => {
      expect(
        resolveQuizViewState(
          makeInput({
            hasCompletedSubmission: true,
            canStartQuiz: false,
            sessionState: 'LOCKED',
            isSingleOrderingQuestion: false,
            hasCurrentQuestion: true,
          }),
        ),
      ).toBe('active-question')
    })

    it('shows active-question when LOCKED with single ordering question', () => {
      expect(
        resolveQuizViewState(
          makeInput({
            hasCompletedSubmission: true,
            canStartQuiz: false,
            sessionState: 'LOCKED',
            isSingleOrderingQuestion: true,
            hasCurrentQuestion: true,
          }),
        ),
      ).toBe('active-question')
    })
  })

  describe('session FINISHED — single ordering question', () => {
    it('shows active-question for single ordering question (betting results)', () => {
      expect(
        resolveQuizViewState(
          makeInput({
            hasCompletedSubmission: true,
            canStartQuiz: false,
            sessionState: 'FINISHED',
            isSingleOrderingQuestion: true,
            hasCurrentQuestion: true,
          }),
        ),
      ).toBe('active-question')
    })
  })

  describe('session FINISHED — multiple questions', () => {
    it('shows results for multiple questions', () => {
      expect(
        resolveQuizViewState(
          makeInput({
            hasCompletedSubmission: true,
            canStartQuiz: false,
            sessionState: 'FINISHED',
            isSingleOrderingQuestion: false,
            hasCurrentQuestion: true,
          }),
        ),
      ).toBe('results')
    })

    it('shows results for single non-ordering question', () => {
      expect(
        resolveQuizViewState(
          makeInput({
            hasCompletedSubmission: true,
            canStartQuiz: false,
            sessionState: 'FINISHED',
            isSingleOrderingQuestion: false,
          }),
        ),
      ).toBe('results')
    })
  })

  describe('session OPEN or no session', () => {
    it('shows results when session is OPEN and quiz is completed', () => {
      expect(
        resolveQuizViewState(
          makeInput({
            hasCompletedSubmission: true,
            canStartQuiz: false,
            sessionState: 'OPEN',
          }),
        ),
      ).toBe('results')
    })

    it('shows results when no session state', () => {
      expect(
        resolveQuizViewState(
          makeInput({
            hasCompletedSubmission: true,
            canStartQuiz: false,
            sessionState: undefined,
          }),
        ),
      ).toBe('results')
    })
  })

  describe('active-question', () => {
    it('returns active-question when there is a current question', () => {
      expect(
        resolveQuizViewState(makeInput({ hasCurrentQuestion: true })),
      ).toBe('active-question')
    })
  })

  describe('results-fallback', () => {
    it('returns results-fallback when completed submission exists but canStartQuiz is true', () => {
      expect(
        resolveQuizViewState(
          makeInput({
            hasCompletedSubmission: true,
            canStartQuiz: true,
          }),
        ),
      ).toBe('results-fallback')
    })
  })

  describe('not-submitted', () => {
    it('returns not-submitted when session ended without submission', () => {
      expect(resolveQuizViewState(makeInput({ isNotSubmitted: true }))).toBe(
        'not-submitted',
      )
    })
  })

  describe('unavailable', () => {
    it('returns unavailable when quiz is unavailable', () => {
      expect(
        resolveQuizViewState(makeInput({ isQuizUnavailable: true })),
      ).toBe('unavailable')
    })
  })

  describe('unknown', () => {
    it('returns unknown when no conditions match', () => {
      expect(resolveQuizViewState(makeInput())).toBe('unknown')
    })
  })

  describe('priority order', () => {
    it('just-completed takes priority over results', () => {
      expect(
        resolveQuizViewState(
          makeInput({
            quizCompleted: true,
            hasFinalResult: true,
            hasCompletedSubmission: true,
            canStartQuiz: false,
          }),
        ),
      ).toBe('just-completed')
    })

    it('results takes priority over active-question (when not locked/finished-ordering)', () => {
      expect(
        resolveQuizViewState(
          makeInput({
            hasCompletedSubmission: true,
            canStartQuiz: false,
            hasCurrentQuestion: true,
          }),
        ),
      ).toBe('results')
    })

    it('not-submitted takes priority over unavailable', () => {
      expect(
        resolveQuizViewState(
          makeInput({
            isNotSubmitted: true,
            isQuizUnavailable: true,
          }),
        ),
      ).toBe('not-submitted')
    })
  })
})

describe('resolveFooterState', () => {
  describe('normal mode', () => {
    it('shows lock-answer button when answer not locked', () => {
      const result = resolveFooterState(
        makeFooterInput({
          actionState: makeActionState({ mode: 'normal', canSubmit: true }),
        }),
      )
      expect(result.bettingModule.visible).toBe(false)
      expect(result.button).toEqual({ type: 'lock-answer', disabled: false })
    })

    it('disables lock-answer when canSubmit is false', () => {
      const result = resolveFooterState(
        makeFooterInput({
          actionState: makeActionState({ mode: 'normal', canSubmit: false }),
        }),
      )
      expect(result.button).toEqual({ type: 'lock-answer', disabled: true })
    })

    it('disables lock-answer when submitting', () => {
      const result = resolveFooterState(
        makeFooterInput({
          actionState: makeActionState({
            mode: 'normal',
            canSubmit: true,
            isSubmitting: true,
          }),
        }),
      )
      expect(result.button).toEqual({ type: 'lock-answer', disabled: true })
    })

    it('shows continue button when answer is locked', () => {
      const result = resolveFooterState(
        makeFooterInput({
          actionState: makeActionState({
            mode: 'normal',
            isAnswerLocked: true,
          }),
        }),
      )
      expect(result.bettingModule.visible).toBe(false)
      expect(result.button).toEqual({ type: 'continue' })
    })

    it('never shows betting module in normal mode', () => {
      const result = resolveFooterState(
        makeFooterInput({
          actionState: makeActionState({ mode: 'normal' }),
          isBettingEnabled: true,
        }),
      )
      expect(result.bettingModule.visible).toBe(false)
    })
  })

  describe('session-betting mode', () => {
    it('shows betting module before answer is locked', () => {
      const result = resolveFooterState(
        makeFooterInput({
          actionState: makeActionState({
            mode: 'session-betting',
            isAnswerLocked: false,
          }),
          isBettingEnabled: true,
        }),
      )
      expect(result.bettingModule).toEqual({
        visible: true,
        mode: 'betting',
        disabled: false,
      })
    })

    it('hides betting module after answer is locked', () => {
      const result = resolveFooterState(
        makeFooterInput({
          actionState: makeActionState({
            mode: 'session-betting',
            isAnswerLocked: true,
            isBetSaved: true,
          }),
          isBettingEnabled: true,
        }),
      )
      expect(result.bettingModule.visible).toBe(false)
    })

    it('hides betting module when betting is not enabled', () => {
      const result = resolveFooterState(
        makeFooterInput({
          actionState: makeActionState({
            mode: 'session-betting',
            isAnswerLocked: false,
          }),
          isBettingEnabled: false,
        }),
      )
      expect(result.bettingModule.visible).toBe(false)
    })

    it('disables betting module when bet is saved and not editing', () => {
      const result = resolveFooterState(
        makeFooterInput({
          actionState: makeActionState({
            mode: 'session-betting',
            isAnswerLocked: false,
            isBetSaved: true,
            isEditing: false,
          }),
          isBettingEnabled: true,
        }),
      )
      expect(result.bettingModule).toEqual({
        visible: true,
        mode: 'betting',
        disabled: true,
      })
    })

    it('enables betting module when editing', () => {
      const result = resolveFooterState(
        makeFooterInput({
          actionState: makeActionState({
            mode: 'session-betting',
            isAnswerLocked: false,
            isBetSaved: true,
            isEditing: true,
          }),
          isBettingEnabled: true,
        }),
      )
      expect(result.bettingModule).toEqual({
        visible: true,
        mode: 'betting',
        disabled: false,
      })
    })

    it('shows save-answer button before bet is saved', () => {
      const result = resolveFooterState(
        makeFooterInput({
          actionState: makeActionState({
            mode: 'session-betting',
            isBetSaved: false,
            canSubmit: true,
          }),
          isBettingEnabled: true,
        }),
      )
      expect(result.button).toEqual({ type: 'save-answer', disabled: false })
    })

    it('disables save-answer when canSubmit is false', () => {
      const result = resolveFooterState(
        makeFooterInput({
          actionState: makeActionState({
            mode: 'session-betting',
            isBetSaved: false,
            canSubmit: false,
          }),
          isBettingEnabled: true,
        }),
      )
      expect(result.button).toEqual({ type: 'save-answer', disabled: true })
    })

    it('shows save-answer button when editing', () => {
      const result = resolveFooterState(
        makeFooterInput({
          actionState: makeActionState({
            mode: 'session-betting',
            isBetSaved: true,
            isEditing: true,
            canSubmit: true,
          }),
          isBettingEnabled: true,
        }),
      )
      expect(result.button).toEqual({ type: 'save-answer', disabled: false })
    })

    describe('after bet saved — ordering (canChangeBet: true)', () => {
      it('shows change-answer with next action when not last question', () => {
        const result = resolveFooterState(
          makeFooterInput({
            actionState: makeActionState({
              mode: 'session-betting',
              isBetSaved: true,
              canChangeBet: true,
              isEditing: false,
              isLastQuestion: false,
            }),
            isBettingEnabled: true,
          }),
        )
        expect(result.button).toEqual({
          type: 'change-answer',
          secondaryAction: 'next',
        })
      })

      it('shows change-answer with finalize on last question (multi-question quiz)', () => {
        const result = resolveFooterState(
          makeFooterInput({
            actionState: makeActionState({
              mode: 'session-betting',
              isBetSaved: true,
              canChangeBet: true,
              isEditing: false,
              isLastQuestion: true,
            }),
            isBettingEnabled: true,
            isSingleQuestion: false,
          }),
        )
        expect(result.button).toEqual({
          type: 'change-answer',
          secondaryAction: 'finalize',
        })
      })

      it('shows only change-answer on single ordering question quiz', () => {
        const result = resolveFooterState(
          makeFooterInput({
            actionState: makeActionState({
              mode: 'session-betting',
              isBetSaved: true,
              canChangeBet: true,
              isEditing: false,
              isLastQuestion: true,
            }),
            isBettingEnabled: true,
            isSingleQuestion: true,
          }),
        )
        expect(result.button).toEqual({
          type: 'change-answer',
          secondaryAction: 'none',
        })
      })
    })

    describe('after bet saved — predefined (canChangeBet: false)', () => {
      it('shows continue (next question) when not last question', () => {
        const result = resolveFooterState(
          makeFooterInput({
            actionState: makeActionState({
              mode: 'session-betting',
              isBetSaved: true,
              canChangeBet: false,
              isEditing: false,
              isLastQuestion: false,
            }),
            isBettingEnabled: true,
          }),
        )
        expect(result.button).toEqual({ type: 'continue' })
      })

      it('shows finalize on last question', () => {
        const result = resolveFooterState(
          makeFooterInput({
            actionState: makeActionState({
              mode: 'session-betting',
              isBetSaved: true,
              canChangeBet: false,
              isEditing: false,
              isLastQuestion: true,
            }),
            isBettingEnabled: true,
          }),
        )
        expect(result.button).toEqual({ type: 'finalize' })
      })
    })
  })

  describe('session-locked mode', () => {
    it('shows locked betting module when betting enabled and answer not locked', () => {
      const result = resolveFooterState(
        makeFooterInput({
          actionState: makeActionState({
            mode: 'session-locked',
            isAnswerLocked: false,
          }),
          isBettingEnabled: true,
        }),
      )
      expect(result.bettingModule).toEqual({
        visible: true,
        mode: 'locked',
        disabled: false,
      })
      expect(result.button).toEqual({ type: 'close' })
    })

    it('hides betting module when answer is locked', () => {
      const result = resolveFooterState(
        makeFooterInput({
          actionState: makeActionState({
            mode: 'session-locked',
            isAnswerLocked: true,
          }),
          isBettingEnabled: true,
        }),
      )
      expect(result.bettingModule.visible).toBe(false)
      expect(result.button).toEqual({ type: 'close' })
    })

    it('hides betting module when betting not enabled', () => {
      const result = resolveFooterState(
        makeFooterInput({
          actionState: makeActionState({ mode: 'session-locked' }),
          isBettingEnabled: false,
        }),
      )
      expect(result.bettingModule.visible).toBe(false)
      expect(result.button).toEqual({ type: 'close' })
    })
  })

  describe('session-results mode', () => {
    it('shows results betting module when betting enabled', () => {
      const result = resolveFooterState(
        makeFooterInput({
          actionState: makeActionState({ mode: 'session-results' }),
          isBettingEnabled: true,
        }),
      )
      expect(result.bettingModule).toEqual({ visible: true, mode: 'results' })
      expect(result.button).toEqual({ type: 'done' })
    })

    it('hides betting module when betting not enabled', () => {
      const result = resolveFooterState(
        makeFooterInput({
          actionState: makeActionState({ mode: 'session-results' }),
          isBettingEnabled: false,
        }),
      )
      expect(result.bettingModule.visible).toBe(false)
      expect(result.button).toEqual({ type: 'done' })
    })
  })

  describe('review mode', () => {
    it('shows review-nav button and no betting module', () => {
      const result = resolveFooterState(
        makeFooterInput({
          actionState: makeActionState({ mode: 'review' }),
          isBettingEnabled: true,
        }),
      )
      expect(result.bettingModule.visible).toBe(false)
      expect(result.button).toEqual({ type: 'review-nav' })
    })
  })

  describe('predefined question with betting — full flow', () => {
    it('OPEN: shows betting module + save button before selecting answer', () => {
      const result = resolveFooterState(
        makeFooterInput({
          actionState: makeActionState({
            mode: 'session-betting',
            canSubmit: false,
            isAnswerLocked: false,
            isBetSaved: false,
            canChangeBet: false,
          }),
          isBettingEnabled: true,
        }),
      )
      expect(result.bettingModule).toEqual({
        visible: true,
        mode: 'betting',
        disabled: false,
      })
      expect(result.button).toEqual({ type: 'save-answer', disabled: true })
    })

    it('OPEN: enables save button after selecting answer', () => {
      const result = resolveFooterState(
        makeFooterInput({
          actionState: makeActionState({
            mode: 'session-betting',
            canSubmit: true,
            isAnswerLocked: false,
            isBetSaved: false,
            canChangeBet: false,
          }),
          isBettingEnabled: true,
        }),
      )
      expect(result.bettingModule).toEqual({
        visible: true,
        mode: 'betting',
        disabled: false,
      })
      expect(result.button).toEqual({ type: 'save-answer', disabled: false })
    })

    it('OPEN: shows next question after locking answer (not last)', () => {
      const result = resolveFooterState(
        makeFooterInput({
          actionState: makeActionState({
            mode: 'session-betting',
            canSubmit: true,
            isAnswerLocked: true,
            isBetSaved: true,
            canChangeBet: false,
            isLastQuestion: false,
          }),
          isBettingEnabled: true,
        }),
      )
      expect(result.bettingModule.visible).toBe(false)
      expect(result.button).toEqual({ type: 'continue' })
    })

    it('OPEN: shows finalize after locking answer (last question)', () => {
      const result = resolveFooterState(
        makeFooterInput({
          actionState: makeActionState({
            mode: 'session-betting',
            canSubmit: true,
            isAnswerLocked: true,
            isBetSaved: true,
            canChangeBet: false,
            isLastQuestion: true,
          }),
          isBettingEnabled: true,
        }),
      )
      expect(result.bettingModule.visible).toBe(false)
      expect(result.button).toEqual({ type: 'finalize' })
    })

    it('LOCKED: shows close button with no betting module (answer already locked)', () => {
      const result = resolveFooterState(
        makeFooterInput({
          actionState: makeActionState({
            mode: 'session-locked',
            isAnswerLocked: true,
            isBetSaved: true,
            canChangeBet: false,
          }),
          isBettingEnabled: true,
        }),
      )
      expect(result.bettingModule.visible).toBe(false)
      expect(result.button).toEqual({ type: 'close' })
    })

    it('FINISHED: shows results betting module + done button', () => {
      const result = resolveFooterState(
        makeFooterInput({
          actionState: makeActionState({
            mode: 'session-results',
            isAnswerLocked: true,
            isBetSaved: true,
            canChangeBet: false,
          }),
          isBettingEnabled: true,
        }),
      )
      expect(result.bettingModule).toEqual({ visible: true, mode: 'results' })
      expect(result.button).toEqual({ type: 'done' })
    })
  })

  describe('ordering question with betting — full flow', () => {
    it('OPEN: shows betting module + save button', () => {
      const result = resolveFooterState(
        makeFooterInput({
          actionState: makeActionState({
            mode: 'session-betting',
            canSubmit: true,
            isAnswerLocked: false,
            isBetSaved: false,
            canChangeBet: true,
          }),
          isBettingEnabled: true,
        }),
      )
      expect(result.bettingModule).toEqual({
        visible: true,
        mode: 'betting',
        disabled: false,
      })
      expect(result.button).toEqual({ type: 'save-answer', disabled: false })
    })

    it('OPEN: shows change-answer with next after saving (not last question)', () => {
      const result = resolveFooterState(
        makeFooterInput({
          actionState: makeActionState({
            mode: 'session-betting',
            canSubmit: true,
            isAnswerLocked: false,
            isBetSaved: true,
            canChangeBet: true,
            isEditing: false,
            isLastQuestion: false,
          }),
          isBettingEnabled: true,
        }),
      )
      expect(result.button).toEqual({
        type: 'change-answer',
        secondaryAction: 'next',
      })
    })

    it('OPEN: shows change-answer with done after saving (last question)', () => {
      const result = resolveFooterState(
        makeFooterInput({
          actionState: makeActionState({
            mode: 'session-betting',
            canSubmit: true,
            isAnswerLocked: false,
            isBetSaved: true,
            canChangeBet: true,
            isEditing: false,
            isLastQuestion: true,
          }),
          isBettingEnabled: true,
        }),
      )
      expect(result.button).toEqual({
        type: 'change-answer',
        secondaryAction: 'finalize',
      })
    })

    it('OPEN: shows save-answer when editing', () => {
      const result = resolveFooterState(
        makeFooterInput({
          actionState: makeActionState({
            mode: 'session-betting',
            canSubmit: true,
            isAnswerLocked: false,
            isBetSaved: true,
            canChangeBet: true,
            isEditing: true,
          }),
          isBettingEnabled: true,
        }),
      )
      expect(result.button).toEqual({ type: 'save-answer', disabled: false })
    })

    it('LOCKED: shows locked betting module + close', () => {
      const result = resolveFooterState(
        makeFooterInput({
          actionState: makeActionState({
            mode: 'session-locked',
            isAnswerLocked: false,
            isBetSaved: true,
            canChangeBet: true,
          }),
          isBettingEnabled: true,
        }),
      )
      expect(result.bettingModule).toEqual({
        visible: true,
        mode: 'locked',
        disabled: true,
      })
      expect(result.button).toEqual({ type: 'close' })
    })

    it('FINISHED: shows results betting module + done', () => {
      const result = resolveFooterState(
        makeFooterInput({
          actionState: makeActionState({
            mode: 'session-results',
            isAnswerLocked: true,
            isBetSaved: true,
            canChangeBet: true,
          }),
          isBettingEnabled: true,
        }),
      )
      expect(result.bettingModule).toEqual({ visible: true, mode: 'results' })
      expect(result.button).toEqual({ type: 'done' })
    })
  })
})
