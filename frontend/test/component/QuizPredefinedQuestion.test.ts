// @vitest-environment nuxt
import { describe, it, expect, beforeEach, vi } from 'vitest'
import { flushPromises } from '@vue/test-utils'
import { mountSuspended, mockNuxtImport } from '@nuxt/test-utils/runtime'
import QuizPredefinedQuestion from '~/components/challenges/quiz/questions/QuizPredefinedQuestion.vue'
import type {
  QuizActionState,
  QuizActionHandlers,
} from '~/components/challenges/quiz/types'
import { QuizSessionState } from '~/api/generated'

const { analyticsMock, submitMock } = vi.hoisted(() => ({
  analyticsMock: vi.fn(),
  submitMock: vi.fn(),
}))
mockNuxtImport('useAnalytics', () => analyticsMock)
mockNuxtImport('useSubmitQuizAnswerMutation', () => submitMock)

const track = vi.fn()
const submitAnswer = vi.fn()

type Props = InstanceType<typeof QuizPredefinedQuestion>['$props']

const QUESTION = {
  id: 'Q1',
  predefinedAnswers: [
    { id: 'a1', answerText: 'Yes', isCorrect: true },
    { id: 'a2', answerText: 'No', isCorrect: false },
  ],
}

// Make the submit mutation resolve as the given answer correctness.
function resolveSubmitAs(isCorrect: boolean | null) {
  submitAnswer.mockResolvedValue({
    data: {
      submitQuizAnswer: { __typename: 'PredefinedResponse', isCorrect },
    },
  })
}

function mountQ(props: Partial<Props> = {}) {
  analyticsMock.mockReturnValue({ track })
  submitMock.mockReturnValue({ executeMutation: submitAnswer })
  return mountSuspended(QuizPredefinedQuestion, {
    props: {
      question: QUESTION,
      totalQuestions: 3,
      currentIndex: 0,
      submissionId: 'SUB1',
      ...props,
    } as Props,
  })
}

type Wrapper = Awaited<ReturnType<typeof mountQ>>
const alt = (w: Wrapper, text: string) =>
  w.findAll('button').find((b) => b.text().includes(text))!
// Exposed via defineExpose({ actionState, handlers }).
type Exposed = { actionState: QuizActionState; handlers: QuizActionHandlers }
const state = (w: Wrapper) => (w.vm as unknown as Exposed).actionState
const handlers = (w: Wrapper) => (w.vm as unknown as Exposed).handlers

describe('QuizPredefinedQuestion', () => {
  beforeEach(() => {
    vi.clearAllMocks()
    resolveSubmitAs(true)
  })

  it('renders every alternative', async () => {
    const wrapper = await mountQ()
    expect(wrapper.text()).toContain('Yes')
    expect(wrapper.text()).toContain('No')
  })

  it('cannot submit until an answer is selected', async () => {
    const wrapper = await mountQ()
    expect(state(wrapper).canSubmit).toBe(false)

    await alt(wrapper, 'Yes').trigger('click')
    expect(state(wrapper).canSubmit).toBe(true)
  })

  it('submits the selected answer and records analytics', async () => {
    const wrapper = await mountQ()
    await alt(wrapper, 'Yes').trigger('click')

    await handlers(wrapper).submit()
    await flushPromises()

    expect(submitAnswer).toHaveBeenCalledWith({
      submissionId: 'SUB1',
      input: {
        questionId: 'Q1',
        selectedAnswerIds: ['a1'],
        betAmount: undefined,
      },
    })
    expect(track).toHaveBeenCalledWith(
      'quiz_answer_submitted',
      expect.objectContaining({ question_id: 'Q1', is_correct: true }),
    )
    expect(state(wrapper).isAnswerLocked).toBe(true)
  })

  it('reveals correct/incorrect answers when revealCorrectAnswers is set', async () => {
    // NB: revealCorrectAnswers is a Boolean prop, so an *absent* value coerces
    // to false (Vue prop casting) — the parent must pass it explicitly to reveal.
    resolveSubmitAs(false)
    const wrapper = await mountQ({ revealCorrectAnswers: true })
    await alt(wrapper, 'No').trigger('click') // pick the wrong one
    await handlers(wrapper).submit()
    await flushPromises()

    // The correct answer ('Yes') is flagged, and the selected wrong answer
    // ('No') gets the negative "your answer" treatment.
    expect(wrapper.text()).toContain('Correct answer')
    expect(wrapper.find('.bg-accent-positive').exists()).toBe(true)
    expect(wrapper.find('.bg-accent-negative').exists()).toBe(true)
  })

  it('hides correctness when revealCorrectAnswers is false', async () => {
    resolveSubmitAs(false)
    const wrapper = await mountQ({ revealCorrectAnswers: false })
    await alt(wrapper, 'No').trigger('click')
    await handlers(wrapper).submit()
    await flushPromises()

    expect(wrapper.text()).not.toContain('Correct answer')
    expect(wrapper.text()).not.toContain('Incorrect answer')
    // The selected answer is still shown as locked ("Your answer").
    expect(wrapper.text()).toContain('Your answer')
  })

  it('emits answerSubmitted with the result on continue', async () => {
    const wrapper = await mountQ()
    await alt(wrapper, 'Yes').trigger('click')
    await handlers(wrapper).submit()
    await flushPromises()

    handlers(wrapper).continue()

    const emitted = wrapper.emitted('answerSubmitted')
    expect(emitted).toBeTruthy()
    expect(emitted![0]![0]).toEqual({ questionId: 'Q1', isCorrect: true })
  })

  it('does not submit when nothing is selected', async () => {
    const wrapper = await mountQ()
    await handlers(wrapper).submit()
    await flushPromises()
    expect(submitAnswer).not.toHaveBeenCalled()
  })

  describe('actionMode reflects the session state', () => {
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

  it('resumes as locked when an existing response is provided', async () => {
    const wrapper = await mountQ({
      existingResponse: {
        selectedAnswers: [{ id: 'a1' }],
        isCorrect: true,
      } as never,
    })

    expect(state(wrapper).isAnswerLocked).toBe(true)
    // Alternatives are disabled once locked.
    expect(alt(wrapper, 'Yes').attributes('disabled')).toBeDefined()
  })
})
