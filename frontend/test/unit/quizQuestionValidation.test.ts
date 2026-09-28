import { describe, it, expect } from 'vitest'
import { QuizQuestionType } from '../../app/api/generated'
import { validateQuizQuestion } from '../../layers/admin/app/utils/quizQuestionValidation'

const predefined = (
  answers: { answerText: string; isCorrect: boolean }[],
  questionText = 'Hvem skrev Romerbrevet?',
) => ({
  questionType: QuizQuestionType.Predefined,
  questionText,
  predefinedAnswers: answers,
})

describe('validateQuizQuestion', () => {
  it('accepts a complete multiple-choice question', () => {
    expect(
      validateQuizQuestion(
        predefined([
          { answerText: 'Paulus', isCorrect: true },
          { answerText: 'Peter', isCorrect: false },
        ]),
      ),
    ).toBeUndefined()
  })

  it('requires question text', () => {
    expect(
      validateQuizQuestion(
        predefined([{ answerText: 'Paulus', isCorrect: true }], '   '),
      ),
    ).toContain('mangler tekst')
  })

  it('requires two answers', () => {
    expect(
      validateQuizQuestion(
        predefined([{ answerText: 'Paulus', isCorrect: true }]),
      ),
    ).toContain('minst to svar')
  })

  // The case that silently did nothing before.
  it('requires a correct answer', () => {
    expect(
      validateQuizQuestion(
        predefined([
          { answerText: 'Paulus', isCorrect: false },
          { answerText: 'Peter', isCorrect: false },
        ]),
      ),
    ).toContain('riktig svar')
  })

  it('requires every answer to have text', () => {
    expect(
      validateQuizQuestion(
        predefined([
          { answerText: 'Paulus', isCorrect: true },
          { answerText: '  ', isCorrect: false },
        ]),
      ),
    ).toContain('må ha tekst')
  })

  it('checks ordering questions for enough non-empty items', () => {
    const ordering = (items: { itemText: string }[]) => ({
      questionType: QuizQuestionType.Ordering,
      questionText: 'Sorter hendelsene',
      orderingItems: items,
    })

    expect(validateQuizQuestion(ordering([{ itemText: 'Ett' }]))).toContain(
      'minst to ledd',
    )
    expect(
      validateQuizQuestion(ordering([{ itemText: 'Ett' }, { itemText: '' }])),
    ).toContain('må ha tekst')
    expect(
      validateQuizQuestion(ordering([{ itemText: 'Ett' }, { itemText: 'To' }])),
    ).toBeUndefined()
  })

  it('rejects an inverted number range', () => {
    const number = (minValue?: number, maxValue?: number) => ({
      questionType: QuizQuestionType.Number,
      questionText: 'Hvor mange?',
      minValue,
      maxValue,
    })

    expect(validateQuizQuestion(number(10, 5))).toContain('større enn')
    expect(validateQuizQuestion(number(5, 10))).toBeUndefined()
    expect(validateQuizQuestion(number(undefined, 10))).toBeUndefined()
    expect(validateQuizQuestion(number(Number.NaN, 5))).toBeUndefined()
  })

  it('accepts a free-text question with only its text', () => {
    expect(
      validateQuizQuestion({
        questionType: QuizQuestionType.FreeText,
        questionText: 'Skriv en refleksjon',
      }),
    ).toBeUndefined()
  })
})
