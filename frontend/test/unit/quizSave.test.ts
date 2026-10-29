import { describe, it, expect } from 'vitest'
import {
  planQuizQuestionSave,
  orderedQuestionIds,
  bettingClearFlags,
  quizQuestionInput,
} from '../../layers/admin/app/utils/quizSave'
import { QuizQuestionType } from '../../app/api/generated'

const q = (id: string | undefined, questionOrder: number) => ({
  id,
  questionOrder,
})

describe('planQuizQuestionSave', () => {
  it('splits the list into deletes, updates and adds', () => {
    const existing = [q('QQ1', 1), q('QQ2', 2), q('QQ3', 3)]
    const next = [q('QQ3', 1), q(undefined, 2), q('QQ1', 3)]

    const plan = planQuizQuestionSave(existing, next)

    expect(plan.deletes).toEqual(['QQ2'])
    expect(plan.updates.map((u) => u.id)).toEqual(['QQ3', 'QQ1'])
    expect(plan.adds).toHaveLength(1)
  })

  // UNIQUE (quiz_id, question_order): inserting at a position a live row still
  // holds is rejected, so new questions are parked past the end.
  it('parks new questions past every stored order', () => {
    const plan = planQuizQuestionSave(
      [q('QQ1', 1), q('QQ2', 7)],
      [q(undefined, 1), q('QQ1', 2), q(undefined, 3)],
    )

    expect(plan.adds.map((a) => a.order)).toEqual([8, 9])
    expect(plan.adds.map((a) => a.position)).toEqual([0, 2])
  })

  it('parks from 1 when nothing is stored yet', () => {
    const plan = planQuizQuestionSave([], [q(undefined, 1), q(undefined, 2)])

    expect(plan.adds.map((a) => a.order)).toEqual([1, 2])
    expect(plan.deletes).toEqual([])
    expect(plan.updates).toEqual([])
  })

  it('deletes every question when the list is emptied', () => {
    const plan = planQuizQuestionSave([q('QQ1', 1), q('QQ2', 2)], [])

    expect(plan.deletes).toEqual(['QQ1', 'QQ2'])
  })
})

describe('orderedQuestionIds', () => {
  it('substitutes the ids of newly added questions by position', () => {
    const next = [q(undefined, 1), q('QQ1', 2), q(undefined, 3)]

    expect(
      orderedQuestionIds(
        next,
        new Map([
          [0, 'QQ9'],
          [2, 'QQ8'],
        ]),
      ),
    ).toEqual(['QQ9', 'QQ1', 'QQ8'])
  })

  // An add that failed must not silently reorder everything else around it.
  it('drops positions that never got an id', () => {
    const next = [q(undefined, 1), q('QQ1', 2)]

    expect(orderedQuestionIds(next, new Map())).toEqual(['QQ1'])
  })
})

describe('bettingClearFlags', () => {
  it('flags a value that was set and is now empty', () => {
    expect(
      bettingClearFlags(
        { bettingMinAbsolute: 10, bettingMaxAbsolute: 100 },
        { bettingMaxAbsolute: 100 },
      ),
    ).toEqual({
      clearBettingMinAbsolute: true,
      clearBettingMaxAbsolute: undefined,
    })
  })

  it('flags nothing when the value was never set', () => {
    expect(bettingClearFlags(undefined, {})).toEqual({
      clearBettingMinAbsolute: undefined,
      clearBettingMaxAbsolute: undefined,
    })
  })

  // A number input that has been emptied reads back as NaN, not undefined.
  it('treats NaN as empty', () => {
    expect(
      bettingClearFlags(
        { bettingMinAbsolute: 10 },
        { bettingMinAbsolute: Number.NaN },
      ).clearBettingMinAbsolute,
    ).toBe(true)
  })
})

describe('quizQuestionInput', () => {
  const question = {
    id: 'QQ1',
    localKey: 'q1',
    questionType: QuizQuestionType.Predefined,
    questionText: 'Hva er ukas greie?',
    questionOrder: 1,
    predefinedAnswers: [
      {
        id: 'QA1',
        localKey: 'a1',
        answerText: 'Kjøttgryten fra Egypt',
        description: 'En gryte som symboliserer maten.',
        isCorrect: true,
        answerOrder: 1,
        translationStatus: [{ languageCode: 'en', fields: ['description'] }],
      },
      {
        id: 'QA2',
        localKey: 'a2',
        answerText: 'Melkekrukken',
        isCorrect: false,
        answerOrder: 2,
      },
    ],
  }

  // The update replaces every answer, so a description left out here is
  // erased on the next save of the question.
  it('sends each answer with its description', () => {
    expect(quizQuestionInput(question).predefinedAnswers).toEqual([
      {
        answerText: 'Kjøttgryten fra Egypt',
        description: 'En gryte som symboliserer maten.',
        isCorrect: true,
        answerOrder: 1,
      },
      {
        answerText: 'Melkekrukken',
        description: undefined,
        isCorrect: false,
        answerOrder: 2,
      },
    ])
  })

  it('sends no answers for a question without them', () => {
    expect(
      quizQuestionInput({
        questionType: QuizQuestionType.FreeText,
        questionText: 'Hva husker du best?',
        questionOrder: 1,
      }).predefinedAnswers,
    ).toBeUndefined()
  })
})

describe('quizQuestionInput image', () => {
  const base = {
    questionType: QuizQuestionType.Predefined,
    questionText: 'Hvem skrev Romerbrevet?',
    questionOrder: 1,
  }

  it('sends the image url when the question has one', () => {
    const input = quizQuestionInput({
      ...base,
      image: 'https://cdn.example.test/q.png',
    })

    expect(input.imageUrl).toBe('https://cdn.example.test/q.png')
  })

  // An omitted field means "leave alone" to updateQuizQuestion, so removing
  // an image in the editor has to travel as a blank string to clear it.
  it('sends a blank string rather than nothing when there is no image', () => {
    expect(quizQuestionInput(base).imageUrl).toBe('')
  })
})
