import type { ChallengePageQuery } from '~/api/generated'

export type QuizChallengeData = Extract<
  ChallengePageQuery['challenge'],
  { __typename: 'QuizChallenge' }
>

export type QuizSubmission =
  QuizChallengeData['quiz']['userSubmissions'][number]

export type OrderedQuestion = QuizSubmission['orderedQuestions'][number]

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
