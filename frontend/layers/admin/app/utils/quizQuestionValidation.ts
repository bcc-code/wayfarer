import { QuizQuestionType } from '~/api/generated'

interface ValidatableQuestion {
  questionType: QuizQuestionType
  questionText: string
  predefinedAnswers?: { answerText: string; isCorrect: boolean }[]
  orderingItems?: { itemText: string }[]
  minValue?: number
  maxValue?: number
}

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

  return undefined
}
