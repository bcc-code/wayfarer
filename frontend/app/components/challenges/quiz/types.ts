import type { ComputedRef } from 'vue'
import type {
  ChallengePageQuery,
  SubmitQuizAnswerMutation,
} from '~/api/generated'

export type QuizChallengeData = Extract<
  ChallengePageQuery['challenge'],
  { __typename: 'QuizChallenge' }
>

export type QuizSubmission =
  QuizChallengeData['quiz']['userSubmissions'][number]

export type OrderedQuestion = QuizSubmission['orderedQuestions'][number]

export type QuizResponseData = QuizSubmission['responses'][number]

export type PredefinedResponseData = Extract<
  QuizResponseData,
  { __typename: 'PredefinedResponse' }
>

export type SubmitAnswerResponse = SubmitQuizAnswerMutation['submitQuizAnswer']

export type PredefinedQuestionData = Extract<
  OrderedQuestion,
  { __typename: 'PredefinedQuestion' }
>

export type NumberQuestionData = Extract<
  OrderedQuestion,
  { __typename: 'NumberQuestion' }
>

export type JsonQuestionData = Extract<
  OrderedQuestion,
  { __typename: 'JsonQuestion' }
>

export type FreeTextQuestionData = Extract<
  OrderedQuestion,
  { __typename: 'FreeTextQuestion' }
>

export type OrderingQuestionData = Extract<
  OrderedQuestion,
  { __typename: 'OrderingQuestion' }
>

export type OrderingResponseData = Extract<
  QuizResponseData,
  { __typename: 'OrderingResponse' }
>

export interface QuestionResult {
  questionId: string
  isCorrect: boolean | null
}

export type QuizActionMode =
  | 'normal' // Standard: lock answer -> continue
  | 'session-betting' // Ordering with session: save -> change workflow
  | 'session-locked' // Session locked, no actions
  | 'review' // Review mode: previous/next navigation

export interface QuizActionState {
  mode: QuizActionMode
  canSubmit: boolean // Has selection/order to submit
  isSubmitting: boolean // Currently submitting
  isAnswerLocked: boolean // Answer confirmed
  isBetSaved: boolean // Bet saved (session betting)
  isEditing: boolean // Editing saved bet
  showPreviousButton: boolean
  isLastQuestion: boolean
}

export interface QuizActionHandlers {
  submit: () => Promise<void>
  continue: () => void
  changeBet: () => void
  previous: () => void
  next: () => void
}

export interface QuizQuestionExposed {
  actionState: ComputedRef<QuizActionState>
  handlers: QuizActionHandlers
}
