import { QuizQuestionType } from '~/api/generated'

interface ValidatableQuestion {
  questionType: QuizQuestionType
  questionText: string
  predefinedAnswers?: { answerText: string; isCorrect: boolean }[]
  orderingItems?: { itemText: string }[]
  minValue?: number
  maxValue?: number
  points?: number
  bettingEnabled?: boolean
  bettingMultiplierCorrect?: number
  bettingMultiplierWrong?: number
}

// Mirrors betting.DefaultMultiplierCorrect / DefaultMultiplierWrong / MaxMultiplier.
export const DEFAULT_MULTIPLIER_CORRECT = 2
export const DEFAULT_MULTIPLIER_WRONG = 0
export const MAX_MULTIPLIER = 100

const multiplier = (value: number | undefined, fallback: number) =>
  value === undefined || Number.isNaN(value) ? fallback : value

/**
 * Why a question cannot be saved yet, or undefined when it can. Returned as a
 * message rather than a boolean: the editor used to `return` silently on a
 * failed check, so the save button did nothing and said nothing.
 */
export function validateQuizQuestion(
  question: ValidatableQuestion,
): string | undefined {
  if (!question.questionText.trim()) return 'Spørsmålet mangler tekst.'

  if (question.questionType === QuizQuestionType.Predefined) {
    const answers = question.predefinedAnswers ?? []
    if (answers.length < 2) return 'Et flervalgsspørsmål trenger minst to svar.'
    if (answers.some((answer) => !answer.answerText.trim())) {
      return 'Alle svaralternativer må ha tekst.'
    }
    if (!answers.some((answer) => answer.isCorrect)) {
      return 'Kryss av minst ett riktig svar.'
    }
  }

  if (question.questionType === QuizQuestionType.Ordering) {
    const items = question.orderingItems ?? []
    if (items.length < 2) return 'Et rekkefølgespørsmål trenger minst to ledd.'
    if (items.some((item) => !item.itemText.trim())) {
      return 'Alle ledd må ha tekst.'
    }
  }

  if (
    question.questionType === QuizQuestionType.Number &&
    question.minValue !== undefined &&
    question.maxValue !== undefined &&
    !Number.isNaN(question.minValue) &&
    !Number.isNaN(question.maxValue) &&
    question.minValue > question.maxValue
  ) {
    return 'Minimumsverdien kan ikke være større enn maksimumsverdien.'
  }

  // A settled bet would overwrite the question points on the answer, so the
  // backend rejects the combination
  if (question.bettingEnabled && (question.points ?? 0) > 0) {
    return 'Et spørsmål med innsats kan ikke også gi poeng. Fjern poengene eller slå av innsats.'
  }

  if (question.bettingEnabled) {
    const correct = multiplier(
      question.bettingMultiplierCorrect,
      DEFAULT_MULTIPLIER_CORRECT,
    )
    const wrong = multiplier(
      question.bettingMultiplierWrong,
      DEFAULT_MULTIPLIER_WRONG,
    )
    for (const value of [correct, wrong]) {
      if (value < 0 || value > MAX_MULTIPLIER) {
        return `Multiplikator må være mellom 0 og ${MAX_MULTIPLIER}.`
      }
      // Same tolerance as betting.MultiplierFromFloat: absorbs float error (1.1 * 100)
      if (Math.abs(value * 100 - Math.round(value * 100)) > 1e-6) {
        return 'Multiplikator kan ha høyst to desimaler.'
      }
    }
    if (wrong > correct) {
      return 'Multiplikator ved feil svar kan ikke være høyere enn ved riktig svar.'
    }
  }

  return undefined
}
