// @vitest-environment nuxt
import { describe, it, expect, beforeEach } from 'vitest'
import { mountSuspended } from '@nuxt/test-utils/runtime'
import QuizResult from '~/components/challenges/quiz/QuizResult.vue'
import QuizProgress from '~/components/challenges/quiz/QuizProgress.vue'

type Props = InstanceType<typeof QuizResult>['$props']
type Result = { questionId: string; isCorrect: boolean | null }

const results = (n: number): Result[] =>
  Array.from({ length: n }, (_, i) => ({
    questionId: `q${i}`,
    isCorrect: true,
  }))

function mountResult(props: Partial<Props> = {}) {
  return mountSuspended(QuizResult, {
    props: {
      score: 0,
      maxScore: 3,
      pointsAwarded: 0,
      results: results(3),
      ...props,
    } as Props,
  })
}

type Wrapper = Awaited<ReturnType<typeof mountResult>>
const btn = (w: Wrapper, text: string) =>
  w.findAll('button').find((b) => b.text().includes(text))

describe('QuizResult', () => {
  beforeEach(() => {
    // Force reduced-motion so the points counter is set synchronously (skips gsap).
    window.matchMedia = (() => ({
      matches: true,
    })) as unknown as typeof matchMedia
  })

  describe('results hidden (revealCorrectAnswers = false)', () => {
    it('shows a thank-you message and no score breakdown', async () => {
      const wrapper = await mountResult({ revealCorrectAnswers: false })
      expect(wrapper.text()).toContain('Thank you for your answers!')
      // Survey use: there is no later reveal, so that line must not appear.
      expect(wrapper.text()).not.toContain(
        'The results will be revealed later.',
      )
      expect(wrapper.findComponent(QuizProgress).exists()).toBe(false)
    })
  })

  describe('results shown (revealCorrectAnswers = true)', () => {
    it('renders the score progress dots', async () => {
      const wrapper = await mountResult({ revealCorrectAnswers: true })
      expect(wrapper.findComponent(QuizProgress).exists()).toBe(true)
    })

    it('says "Perfect!" on a perfect score', async () => {
      const wrapper = await mountResult({
        revealCorrectAnswers: true,
        score: 3,
        maxScore: 3,
      })
      expect(wrapper.text()).toContain('Perfect!')
    })

    it('says "Well done!" on a partial score', async () => {
      const wrapper = await mountResult({
        revealCorrectAnswers: true,
        score: 1,
        maxScore: 3,
      })
      expect(wrapper.text()).toContain('Well done!')
    })

    it('encourages the user on a zero score', async () => {
      const wrapper = await mountResult({
        revealCorrectAnswers: true,
        score: 0,
        maxScore: 3,
      })
      expect(wrapper.text()).toContain('Better luck next time!')
    })

    it('states when no points were awarded', async () => {
      const wrapper = await mountResult({
        revealCorrectAnswers: true,
        pointsAwarded: 0,
      })
      expect(wrapper.text()).toContain('You did not receive any points.')
    })

    it('shows the awarded points (counter settled)', async () => {
      const wrapper = await mountResult({
        revealCorrectAnswers: true,
        score: 2,
        maxScore: 3,
        pointsAwarded: 50,
      })
      expect(wrapper.text()).toContain('You received 50 points.')
    })
  })

  describe('review action', () => {
    it('shows the review button and emits startReview when clicked', async () => {
      const wrapper = await mountResult({
        revealCorrectAnswers: true,
        canReview: true,
      })
      const review = btn(wrapper, 'See answers')
      expect(review).toBeTruthy()

      await review!.trigger('click')
      expect(wrapper.emitted('startReview')).toBeTruthy()
    })

    it('hides the review button when review is unavailable', async () => {
      const wrapper = await mountResult({
        revealCorrectAnswers: true,
        canReview: false,
      })
      expect(btn(wrapper, 'See answers')).toBeUndefined()
      expect(btn(wrapper, 'Done')).toBeTruthy()
    })
  })
})
