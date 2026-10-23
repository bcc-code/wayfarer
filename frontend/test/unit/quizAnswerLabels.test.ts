import { describe, it, expect } from 'vitest'
import { quizAnswerLabel } from '../../layers/user/app/utils/quizAnswerLabels'
import { QuizAnswerLabelStyle } from '../../app/api/generated'

const { None, Numeric, Alpha } = QuizAnswerLabelStyle

describe('quizAnswerLabel', () => {
  it('numbers alternatives from one', () => {
    expect(quizAnswerLabel(0, Numeric)).toBe('1.')
    expect(quizAnswerLabel(4, Numeric)).toBe('5.')
  })

  it('letters alternatives from A', () => {
    expect(quizAnswerLabel(0, Alpha)).toBe('A.')
    expect(quizAnswerLabel(5, Alpha)).toBe('F.')
    expect(quizAnswerLabel(25, Alpha)).toBe('Z.')
  })

  // Not expected in practice, but two alternatives must never share a label.
  it('carries past Z instead of repeating a letter', () => {
    expect(quizAnswerLabel(26, Alpha)).toBe('AA.')
    expect(quizAnswerLabel(27, Alpha)).toBe('AB.')
    expect(quizAnswerLabel(51, Alpha)).toBe('AZ.')
    expect(quizAnswerLabel(52, Alpha)).toBe('BA.')
  })

  it('gives every position a distinct label', () => {
    const labels = Array.from({ length: 60 }, (_, i) =>
      quizAnswerLabel(i, Alpha),
    )
    expect(new Set(labels).size).toBe(labels.length)
  })

  it('labels nothing when the style is NONE or missing', () => {
    expect(quizAnswerLabel(0, None)).toBeNull()
    expect(quizAnswerLabel(0, null)).toBeNull()
    expect(quizAnswerLabel(0, undefined)).toBeNull()
  })

  it('labels nothing for a position that is not a real index', () => {
    expect(quizAnswerLabel(-1, Alpha)).toBeNull()
    expect(quizAnswerLabel(1.5, Alpha)).toBeNull()
  })
})
