// @vitest-environment nuxt
import { describe, it, expect, beforeEach, vi } from 'vitest'
import { flushPromises } from '@vue/test-utils'
import { mountSuspended, mockNuxtImport } from '@nuxt/test-utils/runtime'
import QuizNumberQuestion from '~/components/challenges/quiz/questions/QuizNumberQuestion.vue'
import type {
  QuizActionState,
  QuizActionHandlers,
} from '~/components/challenges/quiz/types'

const { analyticsMock, submitMock } = vi.hoisted(() => ({
  analyticsMock: vi.fn(),
  submitMock: vi.fn(),
}))
mockNuxtImport('useAnalytics', () => analyticsMock)
mockNuxtImport('useSubmitQuizAnswerMutation', () => submitMock)

const track = vi.fn()
const submitAnswer = vi.fn()

type Props = InstanceType<typeof QuizNumberQuestion>['$props']
const QUESTION = { id: 'Q1', minValue: 0, maxValue: 100, stepValue: 1 }

function mountQ(props: Partial<Props> = {}) {
  analyticsMock.mockReturnValue({ track })
  submitMock.mockReturnValue({ executeMutation: submitAnswer })
  return mountSuspended(QuizNumberQuestion, {
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
type Exposed = { actionState: QuizActionState; handlers: QuizActionHandlers }
const state = (w: Wrapper) => (w.vm as unknown as Exposed).actionState
const handlers = (w: Wrapper) => (w.vm as unknown as Exposed).handlers

describe('QuizNumberQuestion', () => {
  beforeEach(() => {
    vi.clearAllMocks()
    submitAnswer.mockResolvedValue({
      data: {
        submitQuizAnswer: { __typename: 'NumberResponse', numberResponse: 42 },
      },
    })
  })

  it('displays the current value', async () => {
    const wrapper = await mountQ({ preSelectedAnswer: 42 })
    expect(wrapper.text()).toContain('42')
  })

  it('submits when the value is zero (0 is a valid answer)', async () => {
    const wrapper = await mountQ({ preSelectedAnswer: 0 })
    await handlers(wrapper).submit()
    await flushPromises()

    expect(submitAnswer).toHaveBeenCalledWith({
      submissionId: 'SUB1',
      input: { questionId: 'Q1', numberResponse: 0, betAmount: undefined },
    })
    expect(state(wrapper).isAnswerLocked).toBe(true)
  })

  it('submits the number and records analytics', async () => {
    const wrapper = await mountQ({ preSelectedAnswer: 42 })
    await handlers(wrapper).submit()
    await flushPromises()

    expect(submitAnswer).toHaveBeenCalledWith({
      submissionId: 'SUB1',
      input: { questionId: 'Q1', numberResponse: 42, betAmount: undefined },
    })
    expect(track).toHaveBeenCalledWith(
      'quiz_answer_submitted',
      expect.objectContaining({ question_id: 'Q1', is_correct: true }),
    )
    expect(state(wrapper).isAnswerLocked).toBe(true)
  })

  it('emits answerSubmitted on continue', async () => {
    const wrapper = await mountQ({ preSelectedAnswer: 42 })
    await handlers(wrapper).submit()
    await flushPromises()

    handlers(wrapper).continue()

    expect(wrapper.emitted('answerSubmitted')?.[0]?.[0]).toMatchObject({
      questionId: 'Q1',
    })
  })

  it('resumes from an existing response as locked', async () => {
    const wrapper = await mountQ({
      existingResponse: { numberResponse: 7 } as never,
    })
    expect(wrapper.text()).toContain('7')
    expect(state(wrapper).isAnswerLocked).toBe(true)
  })

  it('is in review mode when readonly', async () => {
    const wrapper = await mountQ({ readonly: true })
    expect(state(wrapper).mode).toBe('review')
  })
})
