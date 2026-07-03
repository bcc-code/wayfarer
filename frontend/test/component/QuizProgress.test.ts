// @vitest-environment nuxt
import { describe, it, expect } from 'vitest'
import { mountSuspended } from '@nuxt/test-utils/runtime'
import QuizProgress from '~/components/challenges/quiz/QuizProgress.vue'

type Props = InstanceType<typeof QuizProgress>['$props']
type Result = { questionId: string; isCorrect: boolean | null }

function mountProgress(props: Partial<Props> = {}) {
  return mountSuspended(QuizProgress, {
    props: {
      currentIndex: 0,
      totalQuestions: 3,
      results: [],
      ...props,
    } as Props,
  })
}

// The dots are the only elements carrying `transition-colors`.
const dots = (w: Awaited<ReturnType<typeof mountProgress>>) =>
  w.findAll('.transition-colors')

const r = (isCorrect: boolean | null): Result => ({
  questionId: 'q',
  isCorrect,
})

describe('QuizProgress', () => {
  it('renders one dot per question', async () => {
    const wrapper = await mountProgress({ totalQuestions: 5 })
    expect(dots(wrapper)).toHaveLength(5)
  })

  it('marks the current question dot as current', async () => {
    const wrapper = await mountProgress({ currentIndex: 1, totalQuestions: 3 })
    expect(dots(wrapper)[1]!.classes()).toContain('bg-accent-contrast')
  })

  it('shows unanswered, non-current dots as pending', async () => {
    const wrapper = await mountProgress({ currentIndex: 0, totalQuestions: 3 })
    expect(dots(wrapper)[2]!.classes()).toContain('bg-border-default')
  })

  it('colors answered dots by correctness when revealing answers', async () => {
    const wrapper = await mountProgress({
      currentIndex: 2,
      totalQuestions: 3,
      results: [r(true), r(false)],
      revealCorrectAnswers: true,
    })
    const d = dots(wrapper)
    expect(d[0]!.classes()).toContain('bg-accent-positive') // correct
    expect(d[1]!.classes()).toContain('bg-accent-negative') // wrong
  })

  it('hides correctness when revealCorrectAnswers is false', async () => {
    const wrapper = await mountProgress({
      currentIndex: 2,
      totalQuestions: 3,
      results: [r(true), r(false)],
      revealCorrectAnswers: false,
    })
    const d = dots(wrapper)
    // Both answered dots use the neutral "answered" color, not pos/neg.
    expect(d[0]!.classes()).toContain('bg-accent-contrast')
    expect(d[1]!.classes()).toContain('bg-accent-contrast')
    expect(d[0]!.classes()).not.toContain('bg-accent-positive')
    expect(d[1]!.classes()).not.toContain('bg-accent-negative')
  })

  it('prioritizes the current state over an existing result', async () => {
    // Index 0 has a result but is also the current question -> current wins.
    const wrapper = await mountProgress({
      currentIndex: 0,
      totalQuestions: 3,
      results: [r(true)],
    })
    expect(dots(wrapper)[0]!.classes()).toContain('bg-accent-contrast')
  })
})
