import { describe, it, expect } from 'vitest'
import {
  resolveQuizViewState,
  type QuizViewStateInput,
} from '../../app/composables/useQuizViewState'

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
      expect(resolveQuizViewState(makeInput({ isQuizUnavailable: true }))).toBe(
        'unavailable',
      )
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
