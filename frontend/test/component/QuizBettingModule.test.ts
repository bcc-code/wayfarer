// @vitest-environment nuxt
import { describe, it, expect } from 'vitest'
import { mountSuspended } from '@nuxt/test-utils/runtime'
import QuizBettingModule from '~/components/challenges/quiz/QuizBettingModule.vue'
import DesignSlider from '~/components/design/DesignSlider.vue'

type Props = InstanceType<typeof QuizBettingModule>['$props']

function mountBet(props: Partial<Props> = {}) {
  return mountSuspended(QuizBettingModule, {
    props: { availablePoints: 100, ...props } as Props,
  })
}

type Wrapper = Awaited<ReturnType<typeof mountBet>>
const slider = (w: Wrapper) => w.findComponent(DesignSlider)

describe('QuizBettingModule', () => {
  describe('betting mode', () => {
    it('shows the bet and remaining points', async () => {
      const wrapper = await mountBet({ availablePoints: 100, modelValue: 30 })
      expect(wrapper.text()).toContain('30') // your bet
      expect(wrapper.text()).toContain('70') // remaining
    })

    it('defaults the slider bounds to 0..availablePoints', async () => {
      const wrapper = await mountBet({ availablePoints: 100 })
      expect(slider(wrapper).props('min')).toBe(0)
      expect(slider(wrapper).props('max')).toBe(100)
    })

    it('applies percentage limits to the slider bounds', async () => {
      const wrapper = await mountBet({
        availablePoints: 100,
        minPercentage: 10,
        maxPercentage: 50,
      })
      expect(slider(wrapper).props('min')).toBe(10)
      expect(slider(wrapper).props('max')).toBe(50)
    })

    it('takes the stricter of percentage and absolute limits', async () => {
      const wrapper = await mountBet({
        availablePoints: 100,
        minPercentage: 10, // -> 10
        minAbsolute: 20, // stricter min wins
        maxPercentage: 50, // -> 50
        maxAbsolute: 30, // stricter max wins
      })
      expect(slider(wrapper).props('min')).toBe(20)
      expect(slider(wrapper).props('max')).toBe(30)
    })

    it('shows the registered-bet message instead of the slider when locked', async () => {
      const wrapper = await mountBet({ mode: 'locked' })
      expect(wrapper.text()).toContain('The bet has been registered')
      expect(slider(wrapper).exists()).toBe(false)
    })

    it('shows the registered-bet message when disabled', async () => {
      const wrapper = await mountBet({ disabled: true })
      expect(wrapper.text()).toContain('The bet has been registered')
      expect(slider(wrapper).exists()).toBe(false)
    })
  })

  describe('results mode', () => {
    it('shows a win with a positive points value and multiplier', async () => {
      const wrapper = await mountBet({
        mode: 'results',
        betAmount: 10,
        pointsEarned: 20, // resultAmount 30 -> x3
        correctCount: 2,
        totalCount: 3,
      })
      expect(wrapper.text()).toContain('+20')
      expect(wrapper.text()).toContain('Points earned')
      expect(wrapper.text()).toContain('x 3')
    })

    it('renders a fractional multiplier to two decimals', async () => {
      const wrapper = await mountBet({
        mode: 'results',
        betAmount: 10,
        pointsEarned: 5, // resultAmount 15 -> x1.5
        correctCount: 1,
        totalCount: 3,
      })
      expect(wrapper.text()).toContain('x 1.5')
    })

    it('shows a loss with the negative points value', async () => {
      const wrapper = await mountBet({
        mode: 'results',
        betAmount: 10,
        pointsEarned: -5,
        correctCount: 0,
        totalCount: 3,
      })
      expect(wrapper.text()).toContain('Points lost')
      expect(wrapper.text()).toContain('-5')
    })

    it('hides the multiplier when there are no correct answers', async () => {
      const wrapper = await mountBet({
        mode: 'results',
        betAmount: 10,
        pointsEarned: -10,
        correctCount: 0,
        totalCount: 3,
      })
      // correctCount 0 -> "none" label, no multiplier row
      expect(wrapper.text()).toContain("You didn't have any correct ones.")
      expect(wrapper.text()).not.toContain('x ')
    })
  })
})
