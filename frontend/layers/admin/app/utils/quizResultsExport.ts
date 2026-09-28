/** The flat tables behind the copy button and the CSV download. */

// From a char code because Prettier rewrites a "\uFEFF" escape into the
// invisible character, which then trips eslint's no-irregular-whitespace.
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

/** Norwegian decimal comma, one decimal, no trailing ",0". */
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
 * Tab-separated: what pastes into Keynote and Excel as a table. Tabs and
 * newlines in an answer collapse to spaces — the clipboard has no quoting
 * convention to fall back on.
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

/** RFC 4180 quoting. Free-text answers routinely need it. */
function csvCell(value: string): string {
  if (/[",\r\n]/.test(value)) {
    return `"${value.replace(/"/g, '""')}"`
  }
  return value
}

/**
 * A row per option, bucket or group, carrying its question. The BOM is what
 * stops Excel on Windows reading "Håp" as "HÃ¥p".
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
 * Structural so tests need not build the whole generated type. One interface
 * stands in for all five result types, hence the optional per-type fields.
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
 * Branches on the array fields rather than `__typename`. JSON yields no rows;
 * ordering yields per-position accuracy, not the all-or-nothing count.
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

/** Rounded as the backend rounds, so the two never disagree by a tenth. */
export function percentageOfResponses(
  count: number,
  responseCount: number,
): number {
  if (responseCount <= 0) return 0
  return Math.round((count / responseCount) * 1000) / 10
}
