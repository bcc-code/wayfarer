import { describe, it, expect } from 'vitest'
import {
  bucketLabel,
  csvFilename,
  formatDecimal,
  formatPercent,
  percentageOfResponses,
  questionResultRows,
  questionResultTables,
  questionTypeLabel,
  toCsv,
  toTsv,
  type QuestionResultLike,
} from '../../layers/admin/app/utils/quizResultsExport'

describe('formatPercent', () => {
  it('uses a Norwegian decimal comma', () => {
    expect(formatPercent(33.3)).toBe('33,3 %')
  })

  it('drops a trailing zero decimal', () => {
    expect(formatPercent(74)).toBe('74 %')
  })

  it('rounds to one decimal', () => {
    expect(formatPercent(12.34)).toBe('12,3 %')
  })

  it('handles zero', () => {
    expect(formatPercent(0)).toBe('0 %')
  })
})

describe('formatDecimal', () => {
  it('formats without a unit', () => {
    expect(formatDecimal(41.6)).toBe('41,6')
  })

  it('leaves whole numbers whole', () => {
    expect(formatDecimal(40)).toBe('40')
  })
})

describe('percentageOfResponses', () => {
  it('computes a share', () => {
    expect(percentageOfResponses(31, 42)).toBe(73.8)
  })

  it('returns zero rather than dividing by zero', () => {
    expect(percentageOfResponses(0, 0)).toBe(0)
  })

  it('treats a negative total as empty', () => {
    expect(percentageOfResponses(3, -1)).toBe(0)
  })
})

describe('bucketLabel', () => {
  it('renders a range', () => {
    expect(bucketLabel(10, 25)).toBe('10–25')
  })

  it('collapses a zero-width bucket to one number', () => {
    expect(bucketLabel(40, 40)).toBe('40')
  })
})

describe('questionTypeLabel', () => {
  it.each([
    ['PredefinedQuestionResults', 'Flervalg'],
    ['NumberQuestionResults', 'Tall'],
    ['FreeTextQuestionResults', 'Fritekst'],
    ['OrderingQuestionResults', 'Rekkefølge'],
    ['JsonQuestionResults', 'JSON'],
    [undefined, 'JSON'],
  ])('maps %s', (typename, expected) => {
    expect(questionTypeLabel(typename)).toBe(expected)
  })
})

describe('toTsv', () => {
  it('emits a header and one tab-separated row per entry', () => {
    const tsv = toTsv([
      { label: 'Oslo', count: 31, percentage: 73.8, isCorrect: true },
      { label: 'Bergen', count: 7, percentage: 16.7 },
    ])

    expect(tsv.split('\n')).toEqual([
      'Svar\tAntall\tProsent',
      'Oslo\t31\t73,8',
      'Bergen\t7\t16,7',
    ])
  })

  it('collapses tabs and newlines in an answer so the row stays one row', () => {
    const tsv = toTsv([{ label: 'to\tlinjer\nher', count: 1, percentage: 100 }])

    const rows = tsv.split('\n')
    expect(rows).toHaveLength(2)
    expect(rows[1]).toBe('to linjer her\t1\t100')
  })

  it('emits only the header for no rows', () => {
    expect(toTsv([])).toBe('Svar\tAntall\tProsent')
  })
})

describe('toCsv', () => {
  const table = (rows: Parameters<typeof toCsv>[0][number]['rows']) => [
    {
      order: 1,
      questionText: 'Hva er hovedstaden?',
      type: 'Flervalg',
      rows,
    },
  ]

  it('starts with a BOM so Excel reads Norwegian characters correctly', () => {
    const csv = toCsv(table([{ label: 'Håp', count: 1, percentage: 100 }]))
    expect(csv.startsWith(String.fromCharCode(0xfeff))).toBe(true)
  })

  it('uses CRLF line endings', () => {
    const csv = toCsv(table([{ label: 'Oslo', count: 1, percentage: 100 }]))
    expect(csv).toContain('\r\n')
  })

  it('quotes a value containing a comma', () => {
    const csv = toCsv(
      table([{ label: 'håp, glede', count: 2, percentage: 50 }]),
    )
    expect(csv).toContain('"håp, glede"')
  })

  it('doubles an embedded quote', () => {
    const csv = toCsv(
      table([{ label: 'han sa "hei"', count: 1, percentage: 100 }]),
    )
    expect(csv).toContain('"han sa ""hei"""')
  })

  it('quotes a value containing a newline instead of splitting the row', () => {
    const csv = toCsv(
      table([{ label: 'to\nlinjer', count: 1, percentage: 100 }]),
    )
    expect(csv).toContain('"to\nlinjer"')
  })

  it('writes correctness as ja/nei, and blank where it does not apply', () => {
    const csv = toCsv(
      table([
        { label: 'Oslo', count: 1, percentage: 50, isCorrect: true },
        { label: 'Bergen', count: 1, percentage: 50, isCorrect: false },
        { label: 'Uten fasit', count: 1, percentage: 50 },
      ]),
    )

    const lines = csv.split('\r\n')
    expect(lines[1]).toMatch(/,ja$/)
    expect(lines[2]).toMatch(/,nei$/)
    expect(lines[3]).toMatch(/,$/)
  })

  it('carries the question onto every one of its rows', () => {
    const csv = toCsv(
      table([
        { label: 'Oslo', count: 1, percentage: 50 },
        { label: 'Bergen', count: 1, percentage: 50 },
      ]),
    )

    const lines = csv.split('\r\n')
    expect(lines[1]).toContain('Hva er hovedstaden?')
    expect(lines[2]).toContain('Hva er hovedstaden?')
  })

  it('emits only the header when nothing has rows', () => {
    expect(toCsv([]).split('\r\n')).toHaveLength(1)
  })
})

describe('csvFilename', () => {
  it('slugifies the quiz name', () => {
    expect(csvFilename('Bibelquiz Kveld 2')).toBe(
      'bibelquiz-kveld-2-resultater.csv',
    )
  })

  it('keeps Norwegian letters', () => {
    expect(csvFilename('Påske')).toBe('påske-resultater.csv')
  })

  it('falls back when the name has nothing usable', () => {
    expect(csvFilename('!!!')).toBe('quiz-resultater.csv')
  })
})

describe('questionResultRows', () => {
  const question = { id: 'QQ1', questionText: 'Spørsmål', questionOrder: 1 }

  it('maps predefined options, keeping correctness', () => {
    const result: QuestionResultLike = {
      __typename: 'PredefinedQuestionResults',
      question,
      responseCount: 3,
      options: [
        {
          answer: { id: 'QA1', answerText: 'Oslo' },
          count: 2,
          percentage: 66.7,
          isCorrect: true,
        },
        {
          answer: { id: 'QA2', answerText: 'Bergen' },
          count: 1,
          percentage: 33.3,
          isCorrect: false,
        },
      ],
    }

    expect(questionResultRows(result)).toEqual([
      { label: 'Oslo', count: 2, percentage: 66.7, isCorrect: true },
      { label: 'Bergen', count: 1, percentage: 33.3, isCorrect: false },
    ])
  })

  it('maps number buckets to labelled ranges', () => {
    const result: QuestionResultLike = {
      __typename: 'NumberQuestionResults',
      question,
      responseCount: 2,
      buckets: [{ from: 10, to: 25, count: 2, percentage: 100 }],
    }

    expect(questionResultRows(result)).toEqual([
      { label: '10–25', count: 2, percentage: 100 },
    ])
  })

  it('maps free-text groups', () => {
    const result: QuestionResultLike = {
      __typename: 'FreeTextQuestionResults',
      question,
      responseCount: 3,
      groups: [{ text: 'håp', count: 3, percentage: 100 }],
      responses: ['håp', 'Håp', 'HÅP'],
    }

    expect(questionResultRows(result)).toEqual([
      { label: 'håp', count: 3, percentage: 100 },
    ])
  })

  it('maps ordering items to numbered positions', () => {
    const result: QuestionResultLike = {
      __typename: 'OrderingQuestionResults',
      question,
      responseCount: 4,
      items: [
        {
          item: { id: 'QA1', itemText: 'Skapelsen' },
          correctPosition: 1,
          correctlyPlacedCount: 4,
          percentage: 100,
        },
      ],
    }

    expect(questionResultRows(result)).toEqual([
      { label: '1. Skapelsen', count: 4, percentage: 100 },
    ])
  })

  it('yields no rows for a JSON question', () => {
    const result: QuestionResultLike = {
      __typename: 'JsonQuestionResults',
      question,
      responseCount: 8,
    }

    expect(questionResultRows(result)).toEqual([])
  })
})

describe('questionResultTables', () => {
  it('carries the order, text and type of each question', () => {
    const tables = questionResultTables([
      {
        __typename: 'FreeTextQuestionResults',
        question: {
          id: 'QQ2',
          questionText: 'Hva tar du med?',
          questionOrder: 2,
        },
        responseCount: 1,
        groups: [{ text: 'håp', count: 1, percentage: 100 }],
        responses: ['håp'],
      },
    ])

    expect(tables).toEqual([
      {
        order: 2,
        questionText: 'Hva tar du med?',
        type: 'Fritekst',
        rows: [{ label: 'håp', count: 1, percentage: 100 }],
      },
    ])
  })
})
