/**
 * Turning quiz results into something you can paste into a keynote.
 *
 * The page renders from the GraphQL result directly; these helpers produce the
 * flat table behind the copy button and the CSV download. Kept pure and
 * structurally typed (rather than tied to the generated types) so the awkward
 * parts — decimal commas, quoting a free-text answer that contains a comma —
 * are unit-tested instead of eyeballed in a spreadsheet.
 */

/**
 * UTF-8 byte order mark. Built from a char code rather than written literally:
 * Prettier rewrites a "\uFEFF" escape into the invisible character itself, which
 * then trips eslint's no-irregular-whitespace and is impossible to see in a diff.
 */
const BOM = String.fromCharCode(0xfeff)

/** One row of a question's result table. */
export interface ResultRow {
  label: string
  count: number
  /** Share of that question's responses, 0–100. */
  percentage: number
  /** Marks the correct option. Undefined where correctness does not apply. */
  isCorrect?: boolean
}

/** One question, flattened to rows. */
export interface QuestionResultTable {
  order: number
  questionText: string
  type: string
  rows: ResultRow[]
}

/**
 * Norwegian decimal comma, at most one decimal, and no trailing ",0" — a slide
 * wants "74 %", not "74,0 %".
 */
export function formatPercent(value: number): string {
  return `${new Intl.NumberFormat('nb-NO', { maximumFractionDigits: 1 }).format(value)} %`
}

/** Same rounding as formatPercent, without the unit — for CSV cells. */
export function formatDecimal(value: number): string {
  return new Intl.NumberFormat('nb-NO', { maximumFractionDigits: 1 }).format(
    value,
  )
}

/**
 * Tab-separated, because that is what pastes into Keynote and Excel as a real
 * table. Commas would need quoting and then paste as one column.
 *
 * Tabs and newlines inside a free-text answer would break the row apart, so
 * they collapse to spaces — the clipboard has no quoting convention to fall
 * back on the way CSV does.
 */
export function toTsv(rows: ResultRow[]): string {
  const lines = ['Svar\tAntall\tProsent']
  for (const row of rows) {
    lines.push(
      [
        row.label.replace(/[\t\r\n]+/g, ' '),
        String(row.count),
        formatDecimal(row.percentage),
      ].join('\t'),
    )
  }
  return lines.join('\n')
}

/**
 * RFC 4180 quoting: wrap in quotes when the value contains a comma, a quote or
 * a newline, and double any embedded quote. Free-text answers routinely contain
 * all three.
 */
function csvCell(value: string): string {
  if (/[",\r\n]/.test(value)) {
    return `"${value.replace(/"/g, '""')}"`
  }
  return value
}

/**
 * One CSV for the whole quiz: a row per option, bucket or group, carrying its
 * question so the rows stay attributable once they are out of the panel.
 *
 * Prefixed with a UTF-8 BOM — without it Excel on Windows reads "Håp" as "HÃ¥p",
 * which is exactly the kind of thing nobody notices until it is on a screen.
 */
export function toCsv(tables: QuestionResultTable[]): string {
  const lines = [
    [
      'Spørsmålsnr',
      'Spørsmål',
      'Type',
      'Svar',
      'Antall',
      'Prosent',
      'Riktig',
    ].join(','),
  ]

  for (const table of tables) {
    for (const row of table.rows) {
      lines.push(
        [
          String(table.order),
          csvCell(table.questionText),
          csvCell(table.type),
          csvCell(row.label),
          String(row.count),
          csvCell(formatDecimal(row.percentage)),
          row.isCorrect === undefined ? '' : row.isCorrect ? 'ja' : 'nei',
        ].join(','),
      )
    }
  }

  return `${BOM}${lines.join('\r\n')}`
}

/** Filename for the CSV download, without unsafe or shell-awkward characters. */
export function csvFilename(quizName: string): string {
  const slug =
    quizName
      .toLowerCase()
      .replace(/[^a-z0-9æøå]+/g, '-')
      .replace(/^-+|-+$/g, '') || 'quiz'
  return `${slug}-resultater.csv`
}

// ==================== Mapping results to rows ====================

/**
 * The shape the page and the mapper need, kept structural so it does not drag
 * the whole generated result type (and its `Quiz`, and its `Project`...) into a
 * unit test.
 *
 * Every per-type field is optional because this one interface stands in for all
 * five result types; which fields are present is decided by `__typename`. Ids
 * are not optional: the list keys are built from them. The
 * mapper branches on the array fields rather than the typename, so a result
 * that carries `options` is a predefined one whatever it calls itself.
 */
export interface QuestionResultLike {
  __typename?: string
  question: { id: string; questionText: string; questionOrder: number }
  responseCount: number

  // PredefinedQuestionResults
  correctCount?: number
  options?: Array<{
    answer: { id: string; answerText: string }
    count: number
    percentage: number
    isCorrect: boolean
  }>

  // NumberQuestionResults
  average?: number | null
  median?: number | null
  min?: number | null
  max?: number | null
  buckets?: Array<{
    from: number
    to: number
    count: number
    percentage: number
  }>

  // FreeTextQuestionResults
  distinctCount?: number
  groups?: Array<{ text: string; count: number; percentage: number }>
  responses?: string[]

  // OrderingQuestionResults
  fullyCorrectCount?: number
  items?: Array<{
    item: { id: string; itemText: string }
    correctPosition: number
    correctlyPlacedCount: number
    percentage: number
  }>
}

/** Human label for a bucket range, e.g. "10–25". */
export function bucketLabel(from: number, to: number): string {
  if (from === to) return formatDecimal(from)
  return `${formatDecimal(from)}–${formatDecimal(to)}`
}

/** Norwegian label for a result type, used in the CSV's Type column. */
export function questionTypeLabel(typename: string | undefined): string {
  switch (typename) {
    case 'PredefinedQuestionResults':
      return 'Flervalg'
    case 'NumberQuestionResults':
      return 'Tall'
    case 'FreeTextQuestionResults':
      return 'Fritekst'
    case 'OrderingQuestionResults':
      return 'Rekkefølge'
    default:
      return 'JSON'
  }
}

/**
 * Flattens one question's results to rows. JSON questions yield none — there is
 * nothing to tabulate, and an empty table is more honest than a fabricated one.
 *
 * Ordering rows are per-position accuracy ("how many put this item here"), which
 * is what the page shows; the all-or-nothing correct count rides in the page
 * header rather than as a row, because it is not the same kind of number.
 */
export function questionResultRows(result: QuestionResultLike): ResultRow[] {
  if (result.options) {
    return result.options.map((option) => ({
      label: option.answer.answerText,
      count: option.count,
      percentage: option.percentage,
      isCorrect: option.isCorrect,
    }))
  }

  if (result.buckets) {
    return result.buckets.map((bucket) => ({
      label: bucketLabel(bucket.from, bucket.to),
      count: bucket.count,
      percentage: bucket.percentage,
    }))
  }

  if (result.groups) {
    return result.groups.map((group) => ({
      label: group.text,
      count: group.count,
      percentage: group.percentage,
    }))
  }

  if (result.items) {
    return result.items.map((item) => ({
      label: `${item.correctPosition}. ${item.item.itemText}`,
      count: item.correctlyPlacedCount,
      percentage: item.percentage,
    }))
  }

  return []
}

/** Flattens every question, for the whole-quiz CSV. */
export function questionResultTables(
  results: QuestionResultLike[],
): QuestionResultTable[] {
  return results.map((result) => ({
    order: result.question.questionOrder,
    questionText: result.question.questionText,
    type: questionTypeLabel(result.__typename),
    rows: questionResultRows(result),
  }))
}

/**
 * A count as a percentage of that question's responses, 0–100, rounded the same
 * way the backend rounds its own percentages so the two never disagree by a
 * tenth on the same screen.
 */
export function percentageOfResponses(
  count: number,
  responseCount: number,
): number {
  if (responseCount <= 0) return 0
  return Math.round((count / responseCount) * 1000) / 10
}
