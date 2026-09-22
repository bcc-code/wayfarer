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

  const answered = () =>
    predefined([
      { answerText: 'Paulus', isCorrect: true, answerOrder: 1 },
      { answerText: 'Peter', isCorrect: false, answerOrder: 2 },
    ])

  it('says what the checkbox means before the answers, not after', async () => {
    const wrapper = await mount(answered())

    const text = wrapper.text()
    expect(text.indexOf('Kryss av for riktig(e) svar')).toBeGreaterThan(-1)
    expect(text.indexOf('Kryss av for riktig(e) svar')).toBeLessThan(
      text.indexOf('Tillat flere svar'),
    )
  })

  // Answers are whole sentences; a single-line input hid the end of most.
  it('gives each answer a growing textarea, not a one-line input', async () => {
    const wrapper = await mount(answered())

    const areas = wrapper.findAllComponents({ name: 'UTextarea' })
    // The question text plus one per answer.
    expect(areas).toHaveLength(3)
    expect(areas[1]!.props('autoresize')).toBe(true)
  })

  // The answers are the question; scoring and betting are settings about it.
  it('puts the answers before the scoring and betting settings', async () => {
    const wrapper = await mount(answered())

    const text = wrapper.text()
    const answers = text.indexOf('Svaralternativer')
    expect(answers).toBeGreaterThan(-1)
    expect(answers).toBeLessThan(text.indexOf('Poeng'))
    expect(answers).toBeLessThan(text.indexOf('Aktiver betting'))
  })

  // It used to sit above the betting box, away from the list it governs.
  it('keeps "Tillat flere svar" with the answers', async () => {
    const wrapper = await mount(answered())

    const text = wrapper.text()
    expect(text.indexOf('Tillat flere svar')).toBeLessThan(
      text.indexOf('Aktiver betting'),
    )
  })
})
