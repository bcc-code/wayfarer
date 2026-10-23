import { QuizAnswerLabelStyle } from '~/api/generated'

/**
 * The label shown beside an alternative, derived from its position rather
 * than stored per answer.
 *
 * Answers are always served in `answer_order` and are never shuffled — only
 * questions are — so a derived label cannot disagree with the order the
 * alternatives are shown in. Storing one would also put it through
 * translation along with the answer text.
 *
 * `index` is zero-based, as the list is rendered.
 */
export function quizAnswerLabel(
  index: number,
  style: QuizAnswerLabelStyle | null | undefined,
): string | null {
  if (index < 0 || !Number.isInteger(index)) return null

  switch (style) {
    case QuizAnswerLabelStyle.Numeric:
      return `${index + 1}.`
    case QuizAnswerLabelStyle.Alpha:
      return `${alphaLabel(index)}.`
    default:
      return null
  }
}

/**
 * A, B, … Z, AA, AB — spreadsheet style. A question with more than 26
 * alternatives is not something we expect, but running out of letters should
 * repeat nothing rather than produce two answers labelled the same.
 */
function alphaLabel(index: number): string {
  let label = ''
  let n = index
  while (n >= 0) {
    label = String.fromCharCode(65 + (n % 26)) + label
    n = Math.floor(n / 26) - 1
  }
  return label
}
