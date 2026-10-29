import type { QuizQuestionFormData } from '../components/admin/quiz/AdminQuizForm.vue'

interface PlannableQuestion {
  id?: string
  questionOrder: number
}

export interface QuizSavePlan<T> {
  /** Ids of questions removed from the list. */
  deletes: string[]
  updates: { id: string; question: T }[]
  /**
   * New questions with an order past everything already stored:
   * `quiz_questions` has UNIQUE (quiz_id, question_order), so an insert at a
   * position a live row still holds would be rejected. Final positions are
   * applied afterwards by `reorderQuizQuestions`.
   */
  adds: { question: T; order: number; position: number }[]
}

export function planQuizQuestionSave<T extends PlannableQuestion>(
  existing: readonly T[],
  next: readonly T[],
): QuizSavePlan<T> {
  const keptIds = new Set(next.map((q) => q.id).filter(Boolean))

  const parkedFrom =
    existing.reduce((max, q) => Math.max(max, q.questionOrder), 0) + 1

  const plan: QuizSavePlan<T> = { deletes: [], updates: [], adds: [] }

  for (const question of existing) {
    if (question.id && !keptIds.has(question.id)) plan.deletes.push(question.id)
  }

  next.forEach((question, position) => {
    if (question.id) {
      plan.updates.push({ id: question.id, question })
    } else {
      plan.adds.push({
        question,
        order: parkedFrom + plan.adds.length,
        position,
      })
    }
  })

  return plan
}

/**
 * The final question order, once new questions have ids. Positions with no id
 * are dropped — an add that failed must not silently reorder the rest.
 */
export function orderedQuestionIds<T extends PlannableQuestion>(
  next: readonly T[],
  addedIdsByPosition: ReadonlyMap<number, string>,
): string[] {
  return next
    .map(
      (question, position) => question.id ?? addedIdsByPosition.get(position),
    )
    .filter((id): id is string => !!id)
}

const hasValue = (value: number | null | undefined) =>
  value !== null && value !== undefined && !Number.isNaN(value)

/**
 * Betting absolutes need an explicit clear flag: the update input treats an
 * omitted value as "leave alone", so emptying the field would otherwise keep
 * the old number.
 */
export function bettingClearFlags(
  original:
    { bettingMinAbsolute?: number; bettingMaxAbsolute?: number } | undefined,
  updated: { bettingMinAbsolute?: number; bettingMaxAbsolute?: number },
) {
  return {
    clearBettingMinAbsolute:
      hasValue(original?.bettingMinAbsolute) &&
      !hasValue(updated.bettingMinAbsolute)
        ? true
        : undefined,
    clearBettingMaxAbsolute:
      hasValue(original?.bettingMaxAbsolute) &&
      !hasValue(updated.bettingMaxAbsolute)
        ? true
        : undefined,
  }
}

/**
 * The input shared by `addQuizQuestion` and `updateQuizQuestion`. Every field
 * of an answer must be carried here: the update replaces the whole answer set,
 * so a field left out is erased, not kept.
 */
export function quizQuestionInput(question: QuizQuestionFormData) {
  return {
    questionText: question.questionText,
    // Blank, not omitted, when there is no image: the update input reads an
    // omitted field as "leave alone", so a removed image would survive.
    imageUrl: question.image ?? '',
    timeoutSeconds: question.timeoutSeconds,
    points: question.points,
    allowMultipleSelection: question.allowMultipleSelection,
    answerLabelStyle: question.answerLabelStyle,
    predefinedAnswers: question.predefinedAnswers?.map((a) => ({
      answerText: a.answerText,
      description: a.description,
      isCorrect: a.isCorrect,
      answerOrder: a.answerOrder,
    })),
    orderingItems: question.orderingItems?.map((item) => ({
      itemText: item.itemText,
      correctOrder: item.correctOrder,
    })),
    minValue: question.minValue,
    maxValue: question.maxValue,
    stepValue: question.stepValue,
    bettingEnabled: question.bettingEnabled,
    bettingMinPercentage: question.bettingMinPercentage,
    bettingMaxPercentage: question.bettingMaxPercentage,
    bettingMinAbsolute: question.bettingMinAbsolute,
    bettingMaxAbsolute: question.bettingMaxAbsolute,
  }
}
