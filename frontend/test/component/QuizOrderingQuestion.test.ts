// @vitest-environment nuxt
import { describe, it, expect, beforeEach, vi } from 'vitest'
import { flushPromises } from '@vue/test-utils'
import { mountSuspended, mockNuxtImport } from '@nuxt/test-utils/runtime'
import QuizOrderingQuestion from '~/components/challenges/quiz/questions/QuizOrderingQuestion.vue'
import type {
  QuizActionState,
  QuizActionHandlers,
} from '~/components/challenges/quiz/types'
import { QuizSessionState } from '~/api/generated'

const { analyticsMock, submitMock, updateMock, shakeMock, pulseMock } =
  vi.hoisted(() => ({
    analyticsMock: vi.fn(),
    submitMock: vi.fn(),
    updateMock: vi.fn(),
    shakeMock: vi.fn(),
    pulseMock: vi.fn(),
  }))
mockNuxtImport('useAnalytics', () => analyticsMock)
mockNuxtImport('useSubmitQuizAnswerMutation', () => submitMock)
mockNuxtImport('useUpdateQuizAnswerMutation', () => updateMock)
mockNuxtImport('useShake', () => () => ({ shake: shakeMock }))
mockNuxtImport('usePulse', () => () => ({ pulse: pulseMock }))

const track = vi.fn()
const submitAnswer = vi.fn()
const updateAnswer = vi.fn()

type Props = InstanceType<typeof QuizOrderingQuestion>['$props']

const ITEMS = [
  { id: 'i1', itemText: 'First', correctOrder: 1 },
  { id: 'i2', itemText: 'Second', correctOrder: 2 },
  { id: 'i3', itemText: 'Third', correctOrder: 3 },
]
const QUESTION = { id: 'Q1', orderingItems: ITEMS }

// Stub the third-party drag lib; render its slot so the items are still shown.
const stubs = { VueDraggable: { template: '<div><slot /></div>' } }

function mountQ(props: Partial<Props> = {}) {
  analyticsMock.mockReturnValue({ track })
  submitMock.mockReturnValue({ executeMutation: submitAnswer })
  updateMock.mockReturnValue({ executeMutation: updateAnswer })
  return mountSuspended(QuizOrderingQuestion, {
    props: {
      question: QUESTION,
      totalQuestions: 3,
      currentIndex: 0,
      submissionId: 'SUB1',
      ...props,
    } as Props,
    global: { stubs },
  })
}

type Wrapper = Awaited<ReturnType<typeof mountQ>>
type Exposed = { actionState: QuizActionState; handlers: QuizActionHandlers }
const state = (w: Wrapper) => (w.vm as unknown as Exposed).actionState
const handlers = (w: Wrapper) => (w.vm as unknown as Exposed).handlers
const itemTexts = (w: Wrapper) => w.findAll('span.flex-1').map((s) => s.text())

describe('QuizOrderingQuestion', () => {
  beforeEach(() => {
    vi.clearAllMocks()
    submitAnswer.mockResolvedValue({
      data: {
        submitQuizAnswer: {
          __typename: 'OrderingResponse',
          id: 'r1',
          isCorrect: true,
        },
      },
    })
    updateAnswer.mockResolvedValue({ data: { updateQuizAnswer: { id: 'r0' } } })
  })

  it('renders every ordering item', async () => {
    const wrapper = await mountQ()
    expect(itemTexts(wrapper).sort()).toEqual(['First', 'Second', 'Third'])
  })

  it('orders items by preSubmittedOrder in review mode', async () => {
    const wrapper = await mountQ({
      readonly: true,
      preSubmittedOrder: ['i3', 'i1', 'i2'],
    })
    expect(itemTexts(wrapper)).toEqual(['Third', 'First', 'Second'])
  })

  it('orders items by an existing response', async () => {
    const wrapper = await mountQ({
      existingResponse: {
        id: 'r0',
        submittedOrder: ['i2', 'i3', 'i1'],
      } as never,
    })
    expect(itemTexts(wrapper)).toEqual(['Second', 'Third', 'First'])
  })

  describe('drag gating', () => {
    it('is draggable in normal mode', async () => {
      const wrapper = await mountQ()
      expect(wrapper.find('.handle').exists()).toBe(true)
    })

    it('is not draggable when readonly', async () => {
      const wrapper = await mountQ({ readonly: true })
      expect(wrapper.find('.handle').exists()).toBe(false)
    })

    it('is not draggable when the session is locked', async () => {
      const wrapper = await mountQ({ sessionState: QuizSessionState.Locked })
      expect(wrapper.find('.handle').exists()).toBe(false)
    })

    it('is not draggable once a bet is saved (and not editing)', async () => {
      const wrapper = await mountQ({
        sessionState: QuizSessionState.Open,
        existingResponse: {
          id: 'r0',
          submittedOrder: ['i1', 'i2', 'i3'],
        } as never,
      })
      expect(wrapper.find('.handle').exists()).toBe(false)
    })
  })

  describe('actionMode', () => {
    it.each([
      { props: {}, mode: 'normal' },
      {
        props: { sessionState: QuizSessionState.Open },
        mode: 'session-betting',
      },
      {
        props: { sessionState: QuizSessionState.Locked },
        mode: 'session-locked',
      },
      {
        props: { sessionState: QuizSessionState.Finished },
        mode: 'session-results',
      },
      { props: { readonly: true }, mode: 'review' },
    ])('is $mode', async ({ props, mode }) => {
      const wrapper = await mountQ(props)
      expect(state(wrapper).mode).toBe(mode)
    })
  })

  it('submits the current order in normal mode', async () => {
    const wrapper = await mountQ()
    await handlers(wrapper).submit()
    await flushPromises()

    expect(submitAnswer).toHaveBeenCalledTimes(1)
    const arg = submitAnswer.mock.calls[0]![0]
    expect(arg.submissionId).toBe('SUB1')
    expect(arg.input.questionId).toBe('Q1')
    expect([...arg.input.submittedOrder].sort()).toEqual(['i1', 'i2', 'i3'])
    expect(state(wrapper).isAnswerLocked).toBe(true)
  })

  it('saves a new bet in session-betting mode', async () => {
    const wrapper = await mountQ({ sessionState: QuizSessionState.Open })
    await handlers(wrapper).submit()
    await flushPromises()

    expect(submitAnswer).toHaveBeenCalledTimes(1)
    expect(updateAnswer).not.toHaveBeenCalled()
    expect(wrapper.emitted('betSaved')?.[0]).toEqual(['r1'])
    expect(track).toHaveBeenCalledWith(
      'quiz_answer_submitted',
      expect.objectContaining({ action: 'save_bet' }),
    )
    expect(state(wrapper).isBetSaved).toBe(true)
  })

  it('updates an existing bet in session-betting mode', async () => {
    const wrapper = await mountQ({
      sessionState: QuizSessionState.Open,
      existingResponse: {
        id: 'r0',
        submittedOrder: ['i1', 'i2', 'i3'],
      } as never,
    })
    await handlers(wrapper).submit()
    await flushPromises()

    expect(updateAnswer).toHaveBeenCalledTimes(1)
    expect(submitAnswer).not.toHaveBeenCalled()
    expect(updateAnswer.mock.calls[0]![0].responseId).toBe('r0')
    expect(wrapper.emitted('betSaved')?.[0]).toEqual(['r0'])
    expect(track).toHaveBeenCalledWith(
      'quiz_answer_submitted',
      expect.objectContaining({ action: 'update_bet' }),
    )
  })

  it('enters editing mode via changeBet', async () => {
    const wrapper = await mountQ({
      sessionState: QuizSessionState.Open,
      existingResponse: {
        id: 'r0',
        submittedOrder: ['i1', 'i2', 'i3'],
      } as never,
    })
    expect(state(wrapper).isEditing).toBe(false)

    handlers(wrapper).changeBet()
    await flushPromises()

    expect(state(wrapper).isEditing).toBe(true)
    // Editing re-enables dragging.
    expect(wrapper.find('.handle').exists()).toBe(true)
  })

  describe('finished session results', () => {
    it('marks all positions correct when the order matches', async () => {
      const wrapper = await mountQ({
        sessionState: QuizSessionState.Finished,
        existingResponse: {
          id: 'r0',
          submittedOrder: ['i1', 'i2', 'i3'],
        } as never,
      })
      expect(wrapper.findAll('.bg-accent-positive')).toHaveLength(3)
      expect(wrapper.findAll('.bg-accent-negative')).toHaveLength(0)
    })

    it('marks mismatched positions wrong', async () => {
      const wrapper = await mountQ({
        sessionState: QuizSessionState.Finished,
        existingResponse: {
          id: 'r0',
          submittedOrder: ['i3', 'i2', 'i1'], // i3,i1 wrong; i2 stays correct
        } as never,
      })
      expect(wrapper.findAll('.bg-accent-positive')).toHaveLength(1)
      expect(wrapper.findAll('.bg-accent-negative')).toHaveLength(2)
    })
  })
})
