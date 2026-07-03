// @vitest-environment nuxt
import { describe, it, expect, beforeEach, vi } from 'vitest'
import { flushPromises } from '@vue/test-utils'
import { mountSuspended, mockNuxtImport } from '@nuxt/test-utils/runtime'
import QuizFreeTextQuestion from '~/components/challenges/quiz/questions/QuizFreeTextQuestion.vue'
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

type Props = InstanceType<typeof QuizFreeTextQuestion>['$props']
const QUESTION = { id: 'Q1' }

function mountQ(props: Partial<Props> = {}) {
  analyticsMock.mockReturnValue({ track })
  submitMock.mockReturnValue({ executeMutation: submitAnswer })
  return mountSuspended(QuizFreeTextQuestion, {
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

describe('QuizFreeTextQuestion', () => {
  beforeEach(() => {
    vi.clearAllMocks()
    submitAnswer.mockResolvedValue({
      data: {
        submitQuizAnswer: {
          __typename: 'FreeTextResponse',
          textResponse: 'hello',
        },
      },
    })
  })

  it('renders a textarea', async () => {
    const wrapper = await mountQ()
    expect(wrapper.find('textarea').exists()).toBe(true)
  })

  it('does not submit when the text is empty', async () => {
    const wrapper = await mountQ() // defaults to ''
    await handlers(wrapper).submit()
    await flushPromises()
    expect(submitAnswer).not.toHaveBeenCalled()
  })

  it('submits the text and records analytics', async () => {
    const wrapper = await mountQ({ preSelectedAnswer: 'hello' })
    await handlers(wrapper).submit()
    await flushPromises()

    expect(submitAnswer).toHaveBeenCalledWith({
      submissionId: 'SUB1',
      input: { questionId: 'Q1', textResponse: 'hello', betAmount: undefined },
    })
    expect(track).toHaveBeenCalledWith(
      'quiz_answer_submitted',
      expect.objectContaining({ question_id: 'Q1', is_correct: true }),
    )
    expect(state(wrapper).isAnswerLocked).toBe(true)
  })

  it('emits answerSubmitted on continue', async () => {
    const wrapper = await mountQ({ preSelectedAnswer: 'hello' })
    await handlers(wrapper).submit()
    await flushPromises()

    handlers(wrapper).continue()

    expect(wrapper.emitted('answerSubmitted')?.[0]?.[0]).toMatchObject({
      questionId: 'Q1',
    })
  })

  it('resumes from an existing response as locked', async () => {
    const wrapper = await mountQ({
      existingResponse: { textResponse: 'resumed' } as never,
    })
    expect(wrapper.find('textarea').element.value).toBe('resumed')
    expect(state(wrapper).isAnswerLocked).toBe(true)
  })

  it('is in review mode when readonly', async () => {
    const wrapper = await mountQ({ readonly: true })
    expect(state(wrapper).mode).toBe('review')
  })
})
