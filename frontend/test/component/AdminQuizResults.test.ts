// @vitest-environment nuxt
import { describe, it, expect, vi, beforeEach } from 'vitest'
import { mountSuspended } from '@nuxt/test-utils/runtime'
import AdminQuizResultBar from '../../layers/admin/app/components/admin/quiz/AdminQuizResultBar.vue'
import AdminQuizResultsQuestion from '../../layers/admin/app/components/admin/quiz/AdminQuizResultsQuestion.vue'
import type { QuestionResultLike } from '../../layers/admin/app/utils/quizResultsExport'

const question = {
  id: 'QQ1',
  questionText: 'Hva er hovedstaden?',
  questionOrder: 3,
}

describe('AdminQuizResultBar', () => {
  it('shows the label, the absolute and the percentage', async () => {
    const wrapper = await mountSuspended(AdminQuizResultBar, {
      props: { label: 'Oslo', count: 31, percentage: 73.8 },
    })

    expect(wrapper.text()).toContain('Oslo')
    expect(wrapper.text()).toContain('31')
    expect(wrapper.text()).toContain('73,8 %')
  })

  it('sizes the fill to the percentage', async () => {
    const wrapper = await mountSuspended(AdminQuizResultBar, {
      props: { label: 'Oslo', count: 31, percentage: 74 },
    })

    const fill = wrapper.find('[role="img"] > div')
    expect(fill.attributes('style')).toContain('width: 74%')
  })

  it('clamps a percentage outside 0–100 rather than overflowing the track', async () => {
    const over = await mountSuspended(AdminQuizResultBar, {
      props: { label: 'Oslo', count: 1, percentage: 140 },
    })
    expect(over.find('[role="img"] > div').attributes('style')).toContain(
      'width: 100%',
    )

    const under = await mountSuspended(AdminQuizResultBar, {
      props: { label: 'Oslo', count: 0, percentage: -5 },
    })
    expect(under.find('[role="img"] > div').attributes('style')).toContain(
      'width: 0%',
    )
  })

  it('names the bar for screen readers so the graphic is not silent', async () => {
    const wrapper = await mountSuspended(AdminQuizResultBar, {
      props: { label: 'Oslo', count: 31, percentage: 74 },
    })

    expect(wrapper.find('[role="img"]').attributes('aria-label')).toBe(
      'Oslo: 31, 74 %',
    )
  })

  it('marks a correct answer with a word, not colour alone', async () => {
    const wrapper = await mountSuspended(AdminQuizResultBar, {
      props: { label: 'Oslo', count: 31, percentage: 74, isCorrect: true },
    })

    expect(wrapper.text()).toContain('Riktig')
  })

  it('says nothing about correctness where it does not apply', async () => {
    const wrapper = await mountSuspended(AdminQuizResultBar, {
      props: { label: '10–25', count: 6, percentage: 14.3 },
    })

    expect(wrapper.text()).not.toContain('Riktig')
  })
})

describe('AdminQuizResultsQuestion', () => {
  const mount = (result: QuestionResultLike) =>
    mountSuspended(AdminQuizResultsQuestion, { props: { result } })

  it('heads the card with the question number and text', async () => {
    const wrapper = await mount({
      __typename: 'PredefinedQuestionResults',
      question,
      responseCount: 2,
      correctCount: 1,
      options: [
        {
          answer: { id: 'QA1', answerText: 'Oslo' },
          count: 1,
          percentage: 50,
          isCorrect: true,
        },
      ],
    })

    expect(wrapper.text()).toContain('Spørsmål 3')
    expect(wrapper.text()).toContain('Hva er hovedstaden?')
    expect(wrapper.text()).toContain('Flervalg')
  })

  it('renders one bar per predefined option, zero-count ones included', async () => {
    const wrapper = await mount({
      __typename: 'PredefinedQuestionResults',
      question,
      responseCount: 3,
      correctCount: 2,
      options: [
        {
          answer: { id: 'QA1', answerText: 'Oslo' },
          count: 2,
          percentage: 66.7,
          isCorrect: true,
        },
        {
          answer: { id: 'QA2', answerText: 'Bergen' },
          count: 1,
          percentage: 33.3,
          isCorrect: false,
        },
        {
          answer: { id: 'QA3', answerText: 'Tromsø' },
          count: 0,
          percentage: 0,
          isCorrect: false,
        },
      ],
    })

    expect(wrapper.findAllComponents(AdminQuizResultBar)).toHaveLength(3)
    expect(wrapper.text()).toContain('Tromsø')
    expect(wrapper.text()).toContain('2 av 3 svarte helt riktig')
  })

  it('shows stats and buckets for a number question', async () => {
    const wrapper = await mount({
      __typename: 'NumberQuestionResults',
      question,
      responseCount: 3,
      average: 41.6,
      median: 40,
      min: 12,
      max: 95,
      buckets: [
        { from: 10, to: 50, count: 2, percentage: 66.7 },
        { from: 50, to: 95, count: 1, percentage: 33.3 },
      ],
    })

    expect(wrapper.text()).toContain('Snitt')
    expect(wrapper.text()).toContain('41,6')
    expect(wrapper.text()).toContain('Median')
    expect(wrapper.text()).toContain('Høyeste')
    expect(wrapper.findAllComponents(AdminQuizResultBar)).toHaveLength(2)
  })

  it('groups free text and hides the raw list until asked', async () => {
    const wrapper = await mount({
      __typename: 'FreeTextQuestionResults',
      question,
      responseCount: 3,
      distinctCount: 2,
      groups: [
        { text: 'håp', count: 2, percentage: 66.7 },
        { text: 'glede', count: 1, percentage: 33.3 },
      ],
      responses: ['håp', 'Håp', 'glede'],
    })

    expect(wrapper.text()).toContain('3 svar')
    expect(wrapper.text()).toContain('2 unike')
    expect(wrapper.findAll('ul li')).toHaveLength(0)

    const buttons = wrapper.findAll('button')
    const showAll = buttons.find((b) => b.text().includes('Vis alle 3 svar'))
    expect(showAll).toBeDefined()

    await showAll!.trigger('click')
    expect(wrapper.findAll('ul li')).toHaveLength(3)
  })

  it('shows per-position accuracy and the fully-correct count for ordering', async () => {
    const wrapper = await mount({
      __typename: 'OrderingQuestionResults',
      question,
      responseCount: 4,
      fullyCorrectCount: 1,
      items: [
        {
          item: { id: 'QA1', itemText: 'Skapelsen' },
          correctPosition: 1,
          correctlyPlacedCount: 4,
          percentage: 100,
        },
        {
          item: { id: 'QA2', itemText: 'Utgangen' },
          correctPosition: 2,
          correctlyPlacedCount: 2,
          percentage: 50,
        },
      ],
    })

    expect(wrapper.text()).toContain('1 av 4 hadde hele rekkefølgen riktig')
    expect(wrapper.text()).toContain('Riktig plassert')
    expect(wrapper.text()).toContain('1. Skapelsen')
    expect(wrapper.findAllComponents(AdminQuizResultBar)).toHaveLength(2)
  })

  it('says a JSON question cannot be summarised rather than faking a chart', async () => {
    const wrapper = await mount({
      __typename: 'JsonQuestionResults',
      question,
      responseCount: 8,
    })

    expect(wrapper.text()).toContain('8 svarte')
    expect(wrapper.text()).toContain('kan ikke oppsummeres automatisk')
    expect(wrapper.findAllComponents(AdminQuizResultBar)).toHaveLength(0)
  })

  it('says so plainly when nobody has answered yet', async () => {
    const wrapper = await mount({
      __typename: 'PredefinedQuestionResults',
      question,
      responseCount: 0,
      correctCount: 0,
      options: [],
    })

    expect(wrapper.text()).toContain('Ingen har svart på dette spørsmålet ennå')
    expect(wrapper.findAllComponents(AdminQuizResultBar)).toHaveLength(0)
  })

  it('offers no copy button when there is nothing to copy', async () => {
    const wrapper = await mount({
      __typename: 'JsonQuestionResults',
      question,
      responseCount: 8,
    })

    expect(wrapper.text()).not.toContain('Kopier')
  })

  describe('copy', () => {
    const writeText = vi.fn().mockResolvedValue(undefined)

    beforeEach(() => {
      writeText.mockClear()
      Object.defineProperty(navigator, 'clipboard', {
        value: { writeText },
        configurable: true,
      })
    })

    it('copies the rows as tab-separated text', async () => {
      const wrapper = await mount({
        __typename: 'PredefinedQuestionResults',
        question,
        responseCount: 3,
        correctCount: 2,
        options: [
          {
            answer: { id: 'QA1', answerText: 'Oslo' },
            count: 2,
            percentage: 66.7,
            isCorrect: true,
          },
        ],
      })

      const copy = wrapper.findAll('button').find((b) => b.text() === 'Kopier')
      await copy!.trigger('click')

      expect(writeText).toHaveBeenCalledWith(
        'Svar\tAntall\tProsent\nOslo\t2\t66,7',
      )
    })
  })
})
