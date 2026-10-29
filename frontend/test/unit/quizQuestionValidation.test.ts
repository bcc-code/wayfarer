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

  it('rejects points on a question with betting', () => {
    const answers = [
      { answerText: 'Paulus', isCorrect: true },
      { answerText: 'Peter', isCorrect: false },
    ]
    expect(
      validateQuizQuestion({
        ...predefined(answers),
        bettingEnabled: true,
        points: 10,
      }),
    ).toContain('innsats kan ikke også gi poeng')
    expect(
      validateQuizQuestion({
        ...predefined(answers),
        bettingEnabled: true,
        points: 0,
      }),
    ).toBeUndefined()
    expect(
      validateQuizQuestion({
        ...predefined(answers),
        bettingEnabled: false,
        points: 10,
      }),
    ).toBeUndefined()
  })

  describe('betting payout multipliers', () => {
    const betting = (extra: Record<string, unknown> = {}) => ({
      questionType: QuizQuestionType.FreeText,
      questionText: 'Hvorfor?',
      bettingEnabled: true,
      ...extra,
    })

    it('accepts a question that leaves the payout at the default', () => {
      expect(validateQuizQuestion(betting())).toBeUndefined()
    })

    it('accepts a valid custom curve', () => {
      expect(
        validateQuizQuestion(
          betting({
            bettingMultiplierCorrect: 2.5,
            bettingMultiplierWrong: 0.5,
          }),
        ),
      ).toBeUndefined()
    })

    it.each([1.1, 2.3, 1.13])(
      'accepts the valid decimal multiplier %s despite floating-point rounding',
      (value) => {
        expect(
          validateQuizQuestion(betting({ bettingMultiplierCorrect: value })),
        ).toBeUndefined()
      },
    )

    it('rejects a wrong-answer payout above the correct one', () => {
      expect(
        validateQuizQuestion(
          betting({ bettingMultiplierCorrect: 1, bettingMultiplierWrong: 2 }),
        ),
      ).toContain('kan ikke være høyere')
    })

    /**
     * The default correct multiplier is 2, so an unset correct field still
     * caps what the wrong field may be.
     */
    it('compares an unset correct multiplier against its default', () => {
      expect(
        validateQuizQuestion(betting({ bettingMultiplierWrong: 3 })),
      ).toContain('kan ikke være høyere')
      expect(
        validateQuizQuestion(betting({ bettingMultiplierWrong: 1.5 })),
      ).toBeUndefined()
    })

    it('rejects more than two decimals', () => {
      expect(
        validateQuizQuestion(betting({ bettingMultiplierCorrect: 1.234 })),
      ).toContain('to desimaler')
    })

    it.each([-1, 101])('rejects an out-of-range multiplier (%s)', (value) => {
      expect(
        validateQuizQuestion(betting({ bettingMultiplierCorrect: value })),
      ).toContain('mellom 0 og 100')
    })

    it('ignores the payout entirely when betting is off', () => {
      expect(
        validateQuizQuestion(
          betting({
            bettingEnabled: false,
            bettingMultiplierCorrect: 1,
            bettingMultiplierWrong: 99,
          }),
        ),
      ).toBeUndefined()
    })
  })
})
