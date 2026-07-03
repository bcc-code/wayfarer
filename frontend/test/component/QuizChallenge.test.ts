// @vitest-environment nuxt
import { describe, it, expect, beforeEach, vi } from 'vitest'
import { flushPromises } from '@vue/test-utils'
import { nextTick } from 'vue'
import { mountSuspended, mockNuxtImport } from '@nuxt/test-utils/runtime'
import QuizChallenge from '~/components/challenges/QuizChallenge.vue'
import QuizResult from '~/components/challenges/quiz/QuizResult.vue'
import QuizReviewMode from '~/components/challenges/quiz/QuizReviewMode.vue'
import LoadingState from '~/components/LoadingState.vue'
import QuizPredefinedQuestion from '~/components/challenges/quiz/questions/QuizPredefinedQuestion.vue'
import QuizNumberQuestion from '~/components/challenges/quiz/questions/QuizNumberQuestion.vue'
import QuizFreeTextQuestion from '~/components/challenges/quiz/questions/QuizFreeTextQuestion.vue'
import QuizOrderingQuestion from '~/components/challenges/quiz/questions/QuizOrderingQuestion.vue'
import QuizBettingModule from '~/components/challenges/quiz/QuizBettingModule.vue'

// QuizChallenge's own composables + the child question's submit mutation.
const { analyticsMock, startMock, finalizeMock, submitMock } = vi.hoisted(
  () => ({
    analyticsMock: vi.fn(),
    startMock: vi.fn(),
    finalizeMock: vi.fn(),
    submitMock: vi.fn(),
  }),
)
mockNuxtImport('useAnalytics', () => analyticsMock)
mockNuxtImport('useStartQuizSessionMutation', () => startMock)
mockNuxtImport('useFinalizeQuizMutation', () => finalizeMock)
mockNuxtImport('useSubmitQuizAnswerMutation', () => submitMock)
// Ordering questions pull these in; harmless no-ops for the other types.
mockNuxtImport('useShake', () => () => ({ shake: vi.fn() }))
mockNuxtImport('usePulse', () => () => ({ pulse: vi.fn() }))

const track = vi.fn()
const startExec = vi.fn()
const finalizeQuiz = vi.fn()
const submitAnswer = vi.fn()

// --- Fixtures ----------------------------------------------------------------
function predefined(id: string, text: string, answers: [string, string]) {
  return {
    __typename: 'PredefinedQuestion',
    id,
    questionText: text,
    bettingEnabled: false,
    predefinedAnswers: [
      { id: `${id}a`, answerText: answers[0], isCorrect: true },
      { id: `${id}b`, answerText: answers[1], isCorrect: false },
    ],
  }
}
const TWO = [
  predefined('q1', 'Question one?', ['Yes', 'No']),
  predefined('q2', 'Question two?', ['A', 'B']),
]

const predefinedResponse = (qid: string) => ({
  __typename: 'PredefinedResponse',
  question: { id: qid },
  selectedAnswers: [{ id: `${qid}a` }],
  isCorrect: true,
  pointsEarned: 0,
})

function activeSub(questions = TWO, responses: unknown[] = []) {
  return {
    id: 'SUB1',
    completedAt: null,
    score: 0,
    maxScore: 0,
    pointsAwarded: 0,
    orderedQuestions: questions,
    responses,
  }
}
function completedSub(responses: unknown[] = []) {
  return {
    id: 'SUBC',
    completedAt: '2026-01-01T00:00:00Z',
    score: 2,
    maxScore: 3,
    pointsAwarded: 10,
    orderedQuestions: TWO,
    responses,
  }
}

type QuizOverrides = Record<string, unknown>
function makeChallenge(quiz: QuizOverrides = {}) {
  return {
    __typename: 'QuizChallenge',
    id: 'CL1',
    name: 'Survey',
    quiz: {
      id: 'QZ1',
      name: 'Survey',
      revealCorrectAnswers: false,
      timeoutSeconds: null,
      userActiveSession: null,
      userActiveSubmission: { id: 'SUB1' },
      userSubmissions: [activeSub()],
      userCanStart: false,
      ...quiz,
    },
  }
}

type Props = InstanceType<typeof QuizChallenge>['$props']
function mountQuiz(
  quiz: QuizOverrides = {},
  extraProps: Record<string, unknown> = {},
) {
  return mountSuspended(QuizChallenge, {
    props: {
      challenge: makeChallenge(quiz),
      ...extraProps,
    } as unknown as Props,
    // Stub the third-party drag lib used by ordering questions.
    global: { stubs: { VueDraggable: { template: '<div><slot /></div>' } } },
  })
}

// --- Ordering + betting fixtures ---------------------------------------------
function orderingQuestion() {
  return {
    __typename: 'OrderingQuestion',
    id: 'q1',
    questionText: 'Put these in order',
    bettingEnabled: true,
    orderingItems: [
      { id: 'i1', itemText: 'One', correctOrder: 1 },
      { id: 'i2', itemText: 'Two', correctOrder: 2 },
    ],
  }
}
const orderingResponse = (extra: Record<string, unknown> = {}) => ({
  __typename: 'OrderingResponse',
  id: 'R0',
  question: { id: 'q1' },
  submittedOrder: ['i1', 'i2'],
  isCorrect: true,
  betAmount: 10,
  pointsEarned: 20,
  ...extra,
})
function orderingSub(responses: unknown[] = [], completed = false) {
  return {
    id: completed ? 'SUBC' : 'SUB1',
    completedAt: completed ? '2026-01-01T00:00:00Z' : null,
    score: 0,
    maxScore: 0,
    pointsAwarded: 0,
    orderedQuestions: [orderingQuestion()],
    responses,
  }
}

// Minimal single-question fixtures, one per type, for the render-dispatch test.
function makeQuestion(type: string) {
  const base = { id: 'q1', questionText: `${type}?`, bettingEnabled: false }
  switch (type) {
    case 'NumberQuestion':
      return {
        ...base,
        __typename: type,
        minValue: 0,
        maxValue: 10,
        stepValue: 1,
      }
    case 'FreeTextQuestion':
      return { ...base, __typename: type }
    case 'OrderingQuestion':
      return {
        ...base,
        __typename: type,
        orderingItems: [
          { id: 'i1', itemText: 'One', correctOrder: 1 },
          { id: 'i2', itemText: 'Two', correctOrder: 2 },
        ],
      }
    default:
      return {
        ...base,
        __typename: 'PredefinedQuestion',
        predefinedAnswers: [{ id: 'a', answerText: 'A', isCorrect: true }],
      }
  }
}

type Wrapper = Awaited<ReturnType<typeof mountQuiz>>
const button = (w: Wrapper, text: string) =>
  w.findAll('button').find((b) => b.text().trim() === text)
const alt = (w: Wrapper, text: string) =>
  w.findAll('button').find((b) => b.text().includes(text))!

beforeEach(() => {
  vi.clearAllMocks()
  // Force reduced-motion so QuizResult's gsap point counter settles synchronously.
  window.matchMedia = (() => ({
    matches: true,
  })) as unknown as typeof matchMedia
  analyticsMock.mockReturnValue({ track })
  startMock.mockReturnValue({ executeMutation: startExec })
  finalizeMock.mockReturnValue({ executeMutation: finalizeQuiz })
  submitMock.mockReturnValue({ executeMutation: submitAnswer })
  startExec.mockResolvedValue({
    data: {
      startQuizSession: { id: 'SUBNEW', orderedQuestions: TWO, responses: [] },
    },
  })
  finalizeQuiz.mockResolvedValue({
    data: {
      finalizeQuiz: {
        score: 1,
        maxScore: 1,
        pointsAwarded: 5,
        completedAt: '2026-01-01T00:00:00Z',
      },
    },
  })
  submitAnswer.mockResolvedValue({
    data: {
      submitQuizAnswer: { __typename: 'PredefinedResponse', isCorrect: true },
    },
  })
})

describe('QuizChallenge survey flow (revealCorrectAnswers = false)', () => {
  it('labels the commit button "Next question" instead of "Lock answer"', async () => {
    const wrapper = await mountQuiz()

    expect(wrapper.text()).toContain('Question one?')
    expect(button(wrapper, 'Next question')).toBeTruthy()
    expect(button(wrapper, 'Lock answer')).toBeUndefined()
  })

  it('advances to the next question in a single commit tap', async () => {
    const wrapper = await mountQuiz()

    await alt(wrapper, 'Yes').trigger('click')
    await button(wrapper, 'Next question')!.trigger('click')
    await flushPromises()
    await nextTick()

    expect(submitAnswer).toHaveBeenCalledTimes(1)
    expect(wrapper.text()).toContain('Question two?')
    expect(wrapper.text()).not.toContain('Question one?')
  })
})

describe('QuizChallenge normal flow (revealCorrectAnswers = true)', () => {
  it('keeps the "Lock answer" label', async () => {
    const wrapper = await mountQuiz({ revealCorrectAnswers: true })

    expect(button(wrapper, 'Lock answer')).toBeTruthy()
    expect(button(wrapper, 'Next question')).toBeUndefined()
  })

  it('does not auto-advance — locks first, then shows a continue button', async () => {
    const wrapper = await mountQuiz({ revealCorrectAnswers: true })

    await alt(wrapper, 'Yes').trigger('click')
    await button(wrapper, 'Lock answer')!.trigger('click')
    await flushPromises()
    await nextTick()

    expect(wrapper.text()).toContain('Question one?')
    expect(wrapper.text()).not.toContain('Question two?')
    expect(button(wrapper, 'Next question')).toBeTruthy()
  })
})

describe('QuizChallenge view-state routing', () => {
  it('shows the loading state while the session is being started', async () => {
    startExec.mockReturnValue(new Promise(() => {})) // never resolves
    const wrapper = await mountQuiz({
      userActiveSession: { id: 'S1', state: 'OPEN' },
      userActiveSubmission: null,
      userSubmissions: [],
    })

    expect(wrapper.findComponent(LoadingState).exists()).toBe(true)
  })

  it('shows the result screen for a completed submission that cannot be retaken', async () => {
    const wrapper = await mountQuiz({
      revealCorrectAnswers: true,
      userActiveSubmission: null,
      userSubmissions: [completedSub([predefinedResponse('q1')])],
      userCanStart: false,
    })

    expect(wrapper.findComponent(QuizResult).exists()).toBe(true)
    expect(wrapper.text()).toContain('Well done!') // score 2/3
  })

  it('shows the not-submitted screen when the session finished with no submission', async () => {
    const wrapper = await mountQuiz({
      userActiveSession: { id: 'S1', state: 'FINISHED' },
      userActiveSubmission: null,
      userSubmissions: [],
    })

    expect(wrapper.text()).toContain('You have not delivered')
  })

  it('shows the unavailable screen when there is no session and nothing to show', async () => {
    const wrapper = await mountQuiz({
      userActiveSession: null,
      userActiveSubmission: null,
      userSubmissions: [],
      userCanStart: false,
    })

    expect(wrapper.text()).toContain(
      'This quiz is not available at the moment.',
    )
  })
})

describe('QuizChallenge lifecycle', () => {
  it('starts a session and emits "start" when one is open and unsubmitted', async () => {
    const wrapper = await mountQuiz({
      userActiveSession: { id: 'S1', state: 'OPEN' },
      userActiveSubmission: null,
      userSubmissions: [],
    })
    await flushPromises()

    expect(startExec).toHaveBeenCalledWith({ sessionId: 'S1' })
    expect(wrapper.emitted('start')).toBeTruthy()
  })

  it('finalizes and emits "complete" after the last question', async () => {
    const wrapper = await mountQuiz({
      userSubmissions: [activeSub([TWO[0]!])], // single question
    })

    await alt(wrapper, 'Yes').trigger('click')
    await button(wrapper, 'Continue')!.trigger('click') // last question label
    await flushPromises()
    await nextTick()

    expect(finalizeQuiz).toHaveBeenCalledWith({ submissionId: 'SUB1' })
    expect(wrapper.emitted('complete')).toBeTruthy()
    expect(wrapper.findComponent(QuizResult).exists()).toBe(true)
  })
})

describe('QuizChallenge review mode', () => {
  const reviewableQuiz = {
    revealCorrectAnswers: true,
    userActiveSubmission: null,
    userSubmissions: [completedSub([predefinedResponse('q1')])],
    userCanStart: false,
  }

  it('enters review mode from the result screen and returns on finish', async () => {
    const wrapper = await mountQuiz(reviewableQuiz)

    // Result screen offers a review button.
    await button(wrapper, 'See answers')!.trigger('click')
    await nextTick()
    expect(wrapper.findComponent(QuizReviewMode).exists()).toBe(true)

    // Finishing review returns to the result screen.
    wrapper.findComponent(QuizReviewMode).vm.$emit('finish')
    await nextTick()
    expect(wrapper.findComponent(QuizReviewMode).exists()).toBe(false)
    expect(wrapper.findComponent(QuizResult).exists()).toBe(true)
  })
})

describe('QuizChallenge resume', () => {
  it('jumps to the first unanswered question when resuming', async () => {
    const wrapper = await mountQuiz({
      userSubmissions: [activeSub(TWO, [predefinedResponse('q1')])],
    })

    expect(wrapper.text()).toContain('Question two?')
    expect(wrapper.text()).toContain('2 of 2')
  })
})

describe('QuizChallenge question-type dispatch', () => {
  it.each([
    ['PredefinedQuestion', QuizPredefinedQuestion],
    ['NumberQuestion', QuizNumberQuestion],
    ['FreeTextQuestion', QuizFreeTextQuestion],
    ['OrderingQuestion', QuizOrderingQuestion],
  ])('mounts the %s component for that question type', async (type, comp) => {
    const wrapper = await mountQuiz({
      userSubmissions: [activeSub([makeQuestion(type)])],
    })

    expect(wrapper.findComponent(comp).exists()).toBe(true)
  })
})

describe('QuizChallenge ordering + betting flow', () => {
  const bettingModule = (w: Wrapper) => w.findComponent(QuizBettingModule)

  it('shows the betting slider and a "Save answer" button in an open session', async () => {
    const wrapper = await mountQuiz(
      {
        userActiveSession: { id: 'S1', state: 'OPEN' },
        userSubmissions: [orderingSub()],
      },
      { userScore: 100 },
    )

    expect(bettingModule(wrapper).exists()).toBe(true)
    expect(bettingModule(wrapper).props('mode')).toBe('betting')
    expect(button(wrapper, 'Save answer')).toBeTruthy()
    expect(button(wrapper, 'Lock answer')).toBeUndefined()
  })

  it('saves a bet and switches to the change-answer state', async () => {
    submitAnswer.mockResolvedValue({
      data: { submitQuizAnswer: { id: 'R1' } },
    })
    const wrapper = await mountQuiz(
      {
        userActiveSession: { id: 'S1', state: 'OPEN' },
        userSubmissions: [orderingSub()],
      },
      { userScore: 100 },
    )

    await button(wrapper, 'Save answer')!.trigger('click')
    await flushPromises()
    await nextTick()

    expect(submitAnswer).toHaveBeenCalledTimes(1)
    expect(button(wrapper, 'Change answer')).toBeTruthy()
    expect(button(wrapper, 'Save answer')).toBeUndefined()
  })

  it('lets the user edit a saved bet again via "Change answer"', async () => {
    const wrapper = await mountQuiz(
      {
        userActiveSession: { id: 'S1', state: 'OPEN' },
        // A pre-saved bet resumes into the change-answer state.
        userSubmissions: [orderingSub([orderingResponse()])],
      },
      { userScore: 100 },
    )

    expect(button(wrapper, 'Change answer')).toBeTruthy()

    await button(wrapper, 'Change answer')!.trigger('click')
    await nextTick()

    expect(button(wrapper, 'Save answer')).toBeTruthy()
  })

  it('shows the "betting closed" view and a Close button when the session is locked', async () => {
    const wrapper = await mountQuiz(
      {
        userActiveSession: { id: 'S1', state: 'LOCKED' },
        userSubmissions: [orderingSub([orderingResponse()])],
      },
      { userScore: 100 },
    )

    expect(wrapper.text()).toContain('The betting is over')
    expect(bettingModule(wrapper).props('mode')).toBe('locked')
    expect(button(wrapper, 'Close')).toBeTruthy()
  })

  it('shows the betting results with points and correct-count when finished', async () => {
    const wrapper = await mountQuiz(
      {
        userActiveSession: { id: 'S1', state: 'FINISHED' },
        userSubmissions: [
          orderingSub([orderingResponse({ pointsEarned: 20, betAmount: 10 })]),
        ],
      },
      { userScore: 100 },
    )

    const mod = bettingModule(wrapper)
    expect(mod.props('mode')).toBe('results')
    expect(mod.props('pointsEarned')).toBe(20)
    // Both items in correct order -> correctCount 2 of 2 (orchestrator computed).
    expect(mod.props('correctCount')).toBe(2)
    expect(mod.props('totalCount')).toBe(2)
  })

  it('keeps a single finished ordering question on the question view (Game Night case)', async () => {
    const wrapper = await mountQuiz({
      userActiveSession: { id: 'S1', state: 'FINISHED' },
      userActiveSubmission: null,
      // Completed, single ordering question, retakes off.
      userSubmissions: [orderingSub([orderingResponse()], true)],
      userCanStart: false,
    })

    // isSingleOrderingQuestion overrides the results screen — the question stays.
    expect(wrapper.findComponent(QuizOrderingQuestion).exists()).toBe(true)
    expect(wrapper.findComponent(QuizResult).exists()).toBe(false)
  })
})
