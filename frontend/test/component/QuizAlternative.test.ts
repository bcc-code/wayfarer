// @vitest-environment nuxt
import { describe, it, expect } from 'vitest'
import { mountSuspended } from '@nuxt/test-utils/runtime'
import QuizAlternative from '../../layers/user/app/components/challenges/quiz/QuizAlternative.vue'

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

describe('QuizAlternative alignment', () => {
  it('centres a bare alternative', async () => {
    const wrapper = await mountSuspended(QuizAlternative, {
      props: { text: 'Paulus' },
    })

    const classes = wrapper.find('button').classes()
    expect(classes).toContain('text-center')
    expect(classes).not.toContain('text-left')
  })

  // Two centred wrapped lines give the eye no edge to come back to.
  it('aligns an alternative with a description to the left', async () => {
    const wrapper = await mountSuspended(QuizAlternative, {
      props: {
        text: 'A. Kjøttgryten fra Egypt',
        description:
          'En gryte som symboliserer maten israelittene lengtet tilbake til.',
      },
    })

    const classes = wrapper.find('button').classes()
    expect(classes).toContain('text-left')
    expect(classes).not.toContain('text-center')
  })

  it('still shows both the text and the description', async () => {
    const wrapper = await mountSuspended(QuizAlternative, {
      props: { text: 'A. Kjøttgryten', description: 'En gryte fra Egypt.' },
    })

    expect(wrapper.text()).toContain('A. Kjøttgryten')
    expect(wrapper.text()).toContain('En gryte fra Egypt.')
  })
})

/**
 * The label is a column of its own, not a prefix inside the title. Glued to
 * the front of the text it moves with each title's width and the description
 * starts at a different x than the title it explains.
 */
describe('QuizAlternative label', () => {
  const labelled = () =>
    mountSuspended(QuizAlternative, {
      props: {
        text: 'Kjøttgryten fra Egypt',
        label: 'A.',
        description: 'En gryte som symboliserer maten israelittene lengtet.',
      },
    })

  it('renders the label in its own element, outside the text', async () => {
    const wrapper = await labelled()
    const spans = wrapper.find('button').element.children

    const labelEl = spans[0] as HTMLElement
    expect(labelEl.textContent?.trim()).toBe('A.')
    // The title is not inside the label, and the label is not inside the title.
    expect(labelEl.textContent).not.toContain('Kjøttgryten')
  })

  it('keeps the title and its description in one block, sharing a left edge', async () => {
    const wrapper = await labelled()
    const spans = wrapper.find('button').element.children

    const textBlock = spans[1] as HTMLElement
    expect(textBlock.textContent).toContain('Kjøttgryten fra Egypt')
    expect(textBlock.textContent).toContain('En gryte som symboliserer')
    // Stacked in one column, so both start at the same x.
    expect(textBlock.className).toContain('flex-col')
  })

  it('aligns a labelled alternative to the left even without a description', async () => {
    const wrapper = await mountSuspended(QuizAlternative, {
      props: { text: 'Kjøttgryten fra Egypt', label: 'A.' },
    })

    expect(wrapper.find('button').classes()).toContain('text-left')
  })

  it('renders no label element when the question is unlabelled', async () => {
    const wrapper = await mountSuspended(QuizAlternative, {
      props: { text: 'Kjøttgryten fra Egypt' },
    })

    expect(wrapper.find('button').element.children).toHaveLength(1)
  })
})
