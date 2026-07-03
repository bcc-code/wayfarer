// @vitest-environment nuxt
import { describe, it, expect, beforeEach, vi } from 'vitest'
import { nextTick } from 'vue'
import { mountSuspended, mockNuxtImport } from '@nuxt/test-utils/runtime'
import QuizReviewMode from '~/components/challenges/quiz/QuizReviewMode.vue'
import QuizPredefinedQuestion from '~/components/challenges/quiz/questions/QuizPredefinedQuestion.vue'
import QuizNumberQuestion from '~/components/challenges/quiz/questions/QuizNumberQuestion.vue'
import type { QuizActionState } from '~/components/challenges/quiz/types'

// Child question components need these; review mode itself has no data deps.
const { analyticsMock, submitMock, updateMock } = vi.hoisted(() => ({
  analyticsMock: vi.fn(),
  submitMock: vi.fn(),
  updateMock: vi.fn(),
}))
mockNuxtImport('useAnalytics', () => analyticsMock)
mockNuxtImport('useSubmitQuizAnswerMutation', () => submitMock)
mockNuxtImport('useUpdateQuizAnswerMutation', () => updateMock)

const QUESTIONS = [
  {
    __typename: 'PredefinedQuestion',
    id: 'q1',
    questionText: 'What is 2+2?',
    predefinedAnswers: [
      { id: 'a1', answerText: 'Four', isCorrect: true },
      { id: 'a2', answerText: 'Five', isCorrect: false },
    ],
  },
  {
    __typename: 'NumberQuestion',
    id: 'q2',
    questionText: 'Guess a number',
    minValue: 0,
    maxValue: 10,
    stepValue: 1,
  },
]

const RESPONSES = [
  {
    __typename: 'PredefinedResponse',
    question: { id: 'q1' },
    selectedAnswers: [{ id: 'a1' }],
    isCorrect: true,
  },
  {
    __typename: 'NumberResponse',
    question: { id: 'q2' },
    numberResponse: 7,
  },
]

// eslint-disable-next-line @typescript-eslint/no-explicit-any
type AnyProps = any
function mountReview(props: Partial<AnyProps> = {}) {
  analyticsMock.mockReturnValue({ track: vi.fn() })
  submitMock.mockReturnValue({ executeMutation: vi.fn() })
  updateMock.mockReturnValue({ executeMutation: vi.fn() })
  return mountSuspended(QuizReviewMode, {
    props: {
      questions: QUESTIONS,
      responses: RESPONSES,
      revealCorrectAnswers: true,
      ...props,
    } as AnyProps,
  })
}

type Wrapper = Awaited<ReturnType<typeof mountReview>>
type Exposed = {
  actionState: QuizActionState | undefined
  handlers: { next: () => void; previous: () => void }
}
const vm = (w: Wrapper) => w.vm as unknown as Exposed

describe('QuizReviewMode', () => {
  beforeEach(() => {
    vi.clearAllMocks()
  })

  it('renders the first question with its progress label', async () => {
    const wrapper = await mountReview()
    expect(wrapper.text()).toContain('What is 2+2?')
    expect(wrapper.text()).toContain('1 of 2')
    expect(wrapper.findComponent(QuizPredefinedQuestion).exists()).toBe(true)
    expect(wrapper.findComponent(QuizNumberQuestion).exists()).toBe(false)
  })

  it('passes the previously selected answer to the question', async () => {
    const wrapper = await mountReview()
    expect(
      wrapper
        .findComponent(QuizPredefinedQuestion)
        .props('preSelectedAnswerIds'),
    ).toEqual(['a1'])
  })

  it('navigates forward to the next question', async () => {
    const wrapper = await mountReview()

    vm(wrapper).handlers.next()
    await nextTick()

    expect(wrapper.text()).toContain('Guess a number')
    expect(wrapper.text()).toContain('2 of 2')
    expect(wrapper.findComponent(QuizNumberQuestion).exists()).toBe(true)
  })

  it('navigates back to the previous question', async () => {
    const wrapper = await mountReview()

    vm(wrapper).handlers.next()
    await nextTick()
    vm(wrapper).handlers.previous()
    await nextTick()

    expect(wrapper.text()).toContain('What is 2+2?')
  })

  it('emits finish when advancing past the last question', async () => {
    const wrapper = await mountReview()

    vm(wrapper).handlers.next() // -> q2 (last)
    await nextTick()
    vm(wrapper).handlers.next() // past last -> finish
    await nextTick()

    expect(wrapper.emitted('finish')).toBeTruthy()
  })

  it('forwards the current question action state (review mode)', async () => {
    const wrapper = await mountReview()
    expect(vm(wrapper).actionState?.mode).toBe('review')
  })
})
