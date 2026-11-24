// @vitest-environment nuxt
import { describe, it, expect } from 'vitest'
import { mountSuspended } from '@nuxt/test-utils/runtime'
import QuizBettingModule from '../../layers/user/app/components/challenges/quiz/QuizBettingModule.vue'
import DesignSlider from '../../layers/user/app/components/design/DesignSlider.vue'

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

    /**
     * The slider placed its thumb from `(value - min) / (max - min)`, so
     * bounds that met produced `calc(NaN% - 32px)` and threw the thumb off
     * the track. Someone with no points has no amount to choose.
     */
    it('shows no slider when the limits leave nothing to stake', async () => {
      const wrapper = await mountBet({ availablePoints: 0 })
      expect(slider(wrapper).exists()).toBe(false)
      expect(wrapper.text()).toContain('You have no points to bet.')
    })

    it('shows no slider when min and max limits meet', async () => {
      const wrapper = await mountBet({
        availablePoints: 100,
        minAbsolute: 50,
        maxAbsolute: 50,
      })
      expect(slider(wrapper).exists()).toBe(false)
    })

    it('still offers the slider for a one-point stake', async () => {
      const wrapper = await mountBet({ availablePoints: 1 })
      expect(slider(wrapper).exists()).toBe(true)
      expect(slider(wrapper).props('max')).toBe(1)
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
    it('shows a win with its points value and the multiplier from the API', async () => {
      const wrapper = await mountBet({
        mode: 'results',
        betAmount: 10,
        pointsEarned: 20,
        outcome: 'correct',
        multiplier: 3,
      })
      expect(wrapper.text()).toContain('+20')
      expect(wrapper.text()).toContain('Points earned')
      expect(wrapper.text()).toContain('x 3')
    })

    it('renders a fractional multiplier to two decimals', async () => {
      const wrapper = await mountBet({
        mode: 'results',
        betAmount: 10,
        pointsEarned: 5,
        outcome: 'correct',
        multiplier: 1.5,
      })
      expect(wrapper.text()).toContain('x 1.5')
    })

    /**
     * The multiplier is the ledger's, not a reconstruction from the net
     * points: a floored payout makes the two disagree, and the ledger wins.
     */
    it('shows the API multiplier even when the net points imply another', async () => {
      const wrapper = await mountBet({
        mode: 'results',
        betAmount: 3,
        pointsEarned: 4, // reconstructing would give x 2.33
        outcome: 'correct',
        multiplier: 2.5,
      })
      expect(wrapper.text()).toContain('x 2.5')
      expect(wrapper.text()).not.toContain('2.33')
    })

    it('shows a loss with the negative points value', async () => {
      const wrapper = await mountBet({
        mode: 'results',
        betAmount: 10,
        pointsEarned: -5,
        outcome: 'incorrect',
        multiplier: 0.5,
      })
      expect(wrapper.text()).toContain('Points lost')
      expect(wrapper.text()).toContain('-5')
    })

    /**
     * A payout curve that returns the stake nets zero points. That is a won
     * bet, not a lost one, so the verdict decides and not the sign.
     */
    it('treats a break-even bet as a win', async () => {
      const wrapper = await mountBet({
        mode: 'results',
        betAmount: 10,
        pointsEarned: 0,
        outcome: 'correct',
        multiplier: 1,
      })
      expect(wrapper.text()).toContain('Points earned')
      expect(wrapper.text()).not.toContain('Points lost')
    })

    it('reads the verdict, not the points, for a zero-payout loss', async () => {
      const wrapper = await mountBet({
        mode: 'results',
        betAmount: 10,
        pointsEarned: -10,
        outcome: 'incorrect',
        multiplier: 0,
      })
      expect(wrapper.text()).toContain('You answered incorrectly.')
      expect(wrapper.text()).toContain('Points lost')
    })

    it('omits the outcome row entirely when the question was not graded', async () => {
      const wrapper = await mountBet({
        mode: 'results',
        betAmount: 10,
        pointsEarned: 20,
        outcome: null,
        multiplier: null,
      })
      expect(wrapper.text()).not.toContain('You answered')
      expect(wrapper.text()).not.toContain('x ')
      expect(wrapper.text()).toContain('+20')
    })

    it('shows the outcome without a multiplier when the bet is unsettled', async () => {
      const wrapper = await mountBet({
        mode: 'results',
        betAmount: 10,
        pointsEarned: 20,
        outcome: 'correct',
        multiplier: null,
      })
      expect(wrapper.text()).toContain('You answered correctly.')
      expect(wrapper.text()).not.toContain('x ')
    })
  })
})
