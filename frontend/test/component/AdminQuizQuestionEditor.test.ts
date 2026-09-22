// @vitest-environment nuxt
import { describe, it, expect } from 'vitest'
import { mountSuspended } from '@nuxt/test-utils/runtime'
import { QuizQuestionType } from '../../app/api/generated'
import AdminQuizQuestionEditor from '../../layers/admin/app/components/admin/quiz/AdminQuizQuestionEditor.vue'

const questionTypeOptions = [
  { value: QuizQuestionType.Predefined, label: 'Flervalg' },
  { value: QuizQuestionType.FreeText, label: 'Fritekst' },
  { value: QuizQuestionType.Number, label: 'Tall' },
  { value: QuizQuestionType.Ordering, label: 'Rekkefølge' },
]

const predefined = (
  answers: { answerText: string; isCorrect: boolean; answerOrder: number }[],
) => ({
  questionType: QuizQuestionType.Predefined,
  questionText: 'Hvem skrev Romerbrevet?',
  questionOrder: 1,
  points: 1,
  allowMultipleSelection: false,
  predefinedAnswers: answers,
})

const mount = (question: Record<string, unknown>) =>
  mountSuspended(AdminQuizQuestionEditor, {
    props: { question, questionTypeOptions },
  })

const saveButton = (wrapper: Awaited<ReturnType<typeof mount>>) =>
  wrapper
    .findAllComponents({ name: 'UButton' })
    .find((button) => button.text().includes('spørsmål'))!

describe('AdminQuizQuestionEditor', () => {
  it('saves a complete question', async () => {
    const wrapper = await mount(
      predefined([
        { answerText: 'Paulus', isCorrect: true, answerOrder: 1 },
        { answerText: 'Peter', isCorrect: false, answerOrder: 2 },
      ]),
    )

    expect(saveButton(wrapper).props('disabled')).toBe(false)
    await saveButton(wrapper).trigger('click')
    expect(wrapper.emitted('save')).toHaveLength(1)
  })

  // It used to `return` silently, so the button looked broken.
  it('says why it cannot be saved, instead of doing nothing', async () => {
    const wrapper = await mount(
      predefined([
        { answerText: 'Paulus', isCorrect: false, answerOrder: 1 },
        { answerText: 'Peter', isCorrect: false, answerOrder: 2 },
      ]),
    )

    expect(wrapper.text()).toContain('Kryss av minst ett riktig svar')
    expect(saveButton(wrapper).props('disabled')).toBe(true)

    await saveButton(wrapper).trigger('click')
    expect(wrapper.emitted('save')).toBeUndefined()
  })

  it('tells the checkbox what it means before the answers, not after', async () => {
    const wrapper = await mount(
      predefined([
        { answerText: 'Paulus', isCorrect: true, answerOrder: 1 },
        { answerText: 'Peter', isCorrect: false, answerOrder: 2 },
      ]),
    )

    const hint = wrapper.text().indexOf('Kryss av for riktig(e) svar')
    const firstAnswer = wrapper.text().indexOf('Riktig')
    expect(hint).toBeGreaterThan(-1)
    expect(hint).toBeLessThan(firstAnswer)
  })

  it('marks which answers are correct at a glance', async () => {
    const wrapper = await mount(
      predefined([
        { answerText: 'Paulus', isCorrect: true, answerOrder: 1 },
        { answerText: 'Peter', isCorrect: false, answerOrder: 2 },
      ]),
    )

    const badges = wrapper
      .findAllComponents({ name: 'UBadge' })
      .filter((badge) => badge.text() === 'Riktig')
    expect(badges).toHaveLength(1)
  })
})
