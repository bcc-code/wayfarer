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
