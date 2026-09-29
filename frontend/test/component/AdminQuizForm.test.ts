// @vitest-environment nuxt
import { describe, it, expect, vi, beforeEach } from 'vitest'
import { mountSuspended, mockNuxtImport } from '@nuxt/test-utils/runtime'
import { flushPromises } from '@vue/test-utils'
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

const addQuestion = async (wrapper: Awaited<ReturnType<typeof mount>>) => {
  const button = wrapper
    .findAll('button')
    .find((candidate) => candidate.text().includes('Legg til spørsmål'))
  await button?.trigger('click')
}

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
  // The order is drawn per participant when a session submission is created,
  // and the submission's stored order is what the quiz renders.
  it('says when the random order is drawn', async () => {
    const wrapper = await mount()

    expect(wrapper.text()).toContain('Tilfeldig spørsmålsrekkefølge')
    expect(wrapper.text()).toContain('trukket når de starter quizen')
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
    // Completion points land whatever the answers were.
    expect(wrapper.text()).toContain('uansett hvor mange svar som er riktige')
  })

  // You come to this page to write questions, not to set a timeout.
  it('leads with the questions, not the settings', async () => {
    const wrapper = await mount()

    const titles = wrapper
      .findAllComponents({ name: 'AdminSection' })
      .map((section) => section.props('title'))

    expect(titles).toEqual(['Spørsmål', 'Innstillinger'])
  })

  // Title, description and image are the challenge's; a second place to write
  // them is what made this page hard to read.
  it('does not ask again for what the challenge already says', async () => {
    const wrapper = await mount()

    expect(wrapper.text()).not.toContain('Tittel')
    expect(wrapper.text()).not.toContain('Beskrivelse')
    expect(wrapper.text()).not.toContain('Bilde')
  })

  it('saves the inherited title and description untouched', async () => {
    const wrapper = await mount()

    await wrapper.find('form').trigger('submit.prevent')
    await flushPromises()

    expect(wrapper.emitted('save')?.[0]?.[0]).toMatchObject({
      name: 'Quiz 1',
      description: 'Beskrivelse',
    })
  })

  // A stray click outside the dialog would throw away a half-written question.
  it('does not let the question dialog be dismissed by accident', async () => {
    const wrapper = await mount()

    expect(wrapper.findComponent({ name: 'UModal' }).props('dismissible')).toBe(
      false,
    )
  })

  // The editor shows these; the list is owned here.
  it('explains each question type to the editor', async () => {
    const wrapper = await mount()

    await addQuestion(wrapper)
    const options = wrapper
      .findComponent({ name: 'AdminQuizQuestionEditor' })
      .props('questionTypeOptions') as { description?: string }[]

    expect(options.every((option) => option.description)).toBe(true)
  })
})
