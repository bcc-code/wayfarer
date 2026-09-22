// @vitest-environment nuxt
import { describe, it, expect, vi, beforeEach } from 'vitest'
import { mountSuspended, mockNuxtImport } from '@nuxt/test-utils/runtime'
import { QuizQuestionType } from '../../app/api/generated'
import AdminQuizForm from '../../layers/admin/app/components/admin/quiz/AdminQuizForm.vue'

const confirm = vi.fn(() => Promise.resolve(true))
mockNuxtImport('useConfirm', () => () => ({ confirm }))
mockNuxtImport('useToast', () => () => ({ add: vi.fn() }))

const question = (id: string, text: string, order: number, points: number) => ({
  id,
  questionType: QuizQuestionType.Predefined,
  questionText: text,
  questionOrder: order,
  points,
  allowMultipleSelection: false,
  predefinedAnswers: [
    { answerText: 'Ja', isCorrect: true, answerOrder: 1 },
    { answerText: 'Nei', isCorrect: false, answerOrder: 2 },
  ],
})

const quizData = {
  id: 'QZ1',
  name: 'Quiz 1',
  description: 'Beskrivelse',
  randomizeQuestions: false,
  revealCorrectAnswers: true,
  allowRetakes: false,
  completionPoints: 50,
  questions: [question('QQ1', 'Første', 1, 10), question('QQ2', 'Andre', 2, 5)],
}

const mount = () =>
  mountSuspended(AdminQuizForm, {
    props: { quizData, projectId: 'PR1', challengeId: 'CL1' },
  })

const rows = (wrapper: Awaited<ReturnType<typeof mount>>) =>
  wrapper.findAll('.drag-handle').map((handle) => handle.element.parentElement)

describe('AdminQuizForm', () => {
  beforeEach(() => {
    confirm.mockClear()
    confirm.mockResolvedValue(true)
  })

  // Neither number alone answers "what is this quiz worth".
  it('totals the question points and the completion points', async () => {
    const wrapper = await mount()

    expect(wrapper.text()).toContain('15 poeng fra spørsmål')
    expect(wrapper.text()).toContain('65 poeng')
  })

  it('names each question type from the shared option list', async () => {
    const wrapper = await mount()

    expect(rows(wrapper)[0]?.textContent).toContain('1. Flervalg')
    expect(rows(wrapper)[0]?.textContent).toContain('10 poeng')
  })

  it('asks before removing a question', async () => {
    const wrapper = await mount()

    const remove = wrapper
      .findAllComponents({ name: 'UButton' })
      .filter((button) => button.props('color') === 'error')
    await remove[0]!.trigger('click')

    expect(confirm).toHaveBeenCalledWith(
      expect.objectContaining({ title: 'Slette spørsmålet?' }),
    )
  })

  it('keeps the question when the confirmation is declined', async () => {
    confirm.mockResolvedValue(false)
    const wrapper = await mount()

    const remove = wrapper
      .findAllComponents({ name: 'UButton' })
      .filter((button) => button.props('color') === 'error')
    await remove[0]!.trigger('click')
    await wrapper.vm.$nextTick()

    expect(rows(wrapper)).toHaveLength(2)
  })

  // The editor is a dialog now, so the list stays visible behind it.
  it('opens the editor without hiding the list', async () => {
    const wrapper = await mount()

    const edit = wrapper
      .findAllComponents({ name: 'UButton' })
      .find((button) => button.text() === 'Rediger')
    await edit!.trigger('click')

    expect(wrapper.findComponent({ name: 'UModal' }).props('open')).toBe(true)
    expect(rows(wrapper)).toHaveLength(2)
  })

  // `randomizeQuestions` is stored and exposed but nothing orders questions by
  // it, so the setting is labelled as having no effect rather than pretending.
  it('says the randomise setting is not in use', async () => {
    const wrapper = await mount()

    expect(wrapper.text()).toContain('Ikke i bruk ennå')
  })

  // Editing question 7 of 12 should say so.
  it('names the question in the dialog heading', async () => {
    const wrapper = await mount()

    const edit = wrapper
      .findAllComponents({ name: 'UButton' })
      .filter((button) => button.text() === 'Rediger')
    await edit[1]!.trigger('click')

    // UModal teleports its content out of the wrapper, so the heading is read
    // from where it actually lands.
    expect(document.body.textContent).toContain('Spørsmål 2 av 2')
  })

  // "Lagre quiz" is the page's primary action; nothing else should compete.
  it('leaves only the save button solid', async () => {
    const wrapper = await mount()

    const solid = wrapper
      .findAllComponents({ name: 'UButton' })
      .filter((button) => (button.props('variant') ?? 'solid') === 'solid')

    expect(solid.map((button) => button.text())).toEqual(['Lagre quiz'])
  })

  it('explains what the working settings do', async () => {
    const wrapper = await mount()

    expect(wrapper.text()).toContain('Vis riktige svar')
    expect(wrapper.text()).toContain('bare én gang')
  })
})
