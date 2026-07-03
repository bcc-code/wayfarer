// @vitest-environment nuxt
import { describe, it, expect } from 'vitest'
import { mountSuspended } from '@nuxt/test-utils/runtime'
import QuizAlternative from '~/components/challenges/quiz/QuizAlternative.vue'

type Props = InstanceType<typeof QuizAlternative>['$props']

// The test env renders the English (fallback) locale.
const T = {
  yourAnswer: 'Your answer',
  wrongAnswer: 'Incorrect answer',
  correctAnswer: 'Correct answer',
}

function mountAlt(props: Partial<Props> = {}) {
  return mountSuspended(QuizAlternative, {
    props: { text: 'Option A', ...props } as Props,
  })
}

const hasIcon = (w: Awaited<ReturnType<typeof mountAlt>>, name: string) =>
  w.findComponent({ name }).exists()

describe('QuizAlternative', () => {
  it('renders the answer text', async () => {
    const wrapper = await mountAlt({ text: 'The capital is Oslo' })
    expect(wrapper.text()).toContain('The capital is Oslo')
  })

  it('applies the disabled attribute', async () => {
    const wrapper = await mountAlt({ disabled: true })
    expect(wrapper.find('button').attributes('disabled')).toBeDefined()
  })

  it('shows no status badge before confirmation', async () => {
    const wrapper = await mountAlt({
      highlighted: true,
      selected: true,
      confirmed: false,
    })
    expect(hasIcon(wrapper, 'IconCheck')).toBe(false)
    expect(hasIcon(wrapper, 'IconClose')).toBe(false)
    expect(hasIcon(wrapper, 'IconLock')).toBe(false)
  })

  it('marks a correct answer with a positive badge', async () => {
    const wrapper = await mountAlt({ confirmed: true, correct: true })
    expect(hasIcon(wrapper, 'IconCheck')).toBe(true)
    expect(wrapper.text()).toContain(T.correctAnswer)
    expect(wrapper.find('.bg-accent-positive').exists()).toBe(true)
  })

  it('marks your own wrong answer as "your answer"', async () => {
    const wrapper = await mountAlt({
      confirmed: true,
      wrong: true,
      selected: true,
    })
    expect(hasIcon(wrapper, 'IconClose')).toBe(true)
    expect(wrapper.text()).toContain(T.yourAnswer)
    expect(wrapper.find('.bg-accent-negative').exists()).toBe(true)
  })

  it('marks a non-selected wrong answer as "wrong answer"', async () => {
    const wrapper = await mountAlt({
      confirmed: true,
      wrong: true,
      selected: false,
    })
    expect(hasIcon(wrapper, 'IconClose')).toBe(true)
    expect(wrapper.text()).toContain(T.wrongAnswer)
  })

  it('marks your selected answer as locked when correctness is hidden', async () => {
    const wrapper = await mountAlt({
      confirmed: true,
      selected: true,
      correct: false,
      wrong: false,
    })
    // Lock icon + "your answer", using the neutral accent (not positive/negative).
    expect(hasIcon(wrapper, 'IconLock')).toBe(true)
    expect(wrapper.text()).toContain(T.yourAnswer)
    expect(wrapper.find('.bg-accent-positive').exists()).toBe(false)
    expect(wrapper.find('.bg-accent-negative').exists()).toBe(false)
  })
})
