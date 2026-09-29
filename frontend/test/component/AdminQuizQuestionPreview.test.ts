// @vitest-environment nuxt
import { describe, it, expect } from 'vitest'
import { mountSuspended } from '@nuxt/test-utils/runtime'
import { QuizQuestionType } from '../../app/api/generated'
import AdminQuizQuestionPreview from '../../layers/admin/app/components/admin/quiz/AdminQuizQuestionPreview.vue'
import QuizPredefinedQuestion from '../../layers/user/app/components/challenges/quiz/questions/QuizPredefinedQuestion.vue'
import QuizOrderingQuestion from '../../layers/user/app/components/challenges/quiz/questions/QuizOrderingQuestion.vue'
import QuizFreeTextQuestion from '../../layers/user/app/components/challenges/quiz/questions/QuizFreeTextQuestion.vue'

const predefined = {
  questionType: QuizQuestionType.Predefined,
  questionText: 'Hvem skrev Romerbrevet?',
  questionOrder: 1,
  allowMultipleSelection: false,
  predefinedAnswers: [
    { answerText: 'Paulus', isCorrect: true, answerOrder: 1 },
    { answerText: 'Peter', isCorrect: false, answerOrder: 2 },
  ],
}

/**
 * The preview renders the participant app's own question components, so an
 * organiser sees the real thing rather than the admin's idea of it.
 */
describe('AdminQuizQuestionPreview', () => {
  it('shows the question and its alternatives as a participant sees them', async () => {
    const wrapper = await mountSuspended(AdminQuizQuestionPreview, {
      props: { question: predefined },
    })

    expect(wrapper.findComponent(QuizPredefinedQuestion).exists()).toBe(true)
    expect(wrapper.text()).toContain('Hvem skrev Romerbrevet?')
    expect(wrapper.text()).toContain('Paulus')
    expect(wrapper.text()).toContain('Peter')
  })

  it.each([
    ['ordering', QuizQuestionType.Ordering, QuizOrderingQuestion],
    ['free text', QuizQuestionType.FreeText, QuizFreeTextQuestion],
  ])(
    'renders a %s question with its own component',
    async (_name, questionType, component) => {
      const wrapper = await mountSuspended(AdminQuizQuestionPreview, {
        props: {
          question: {
            questionType,
            questionText: 'Sett i rekkefølge',
            questionOrder: 1,
            orderingItems: [
              { itemText: 'Først', correctOrder: 1 },
              { itemText: 'Så', correctOrder: 2 },
            ],
          },
        },
      })

      expect(wrapper.findComponent(component).exists()).toBe(true)
    },
  )

  // A draft has no ids until it is saved, and the components key on them.
  it('gives unsaved alternatives ids of their own', async () => {
    const wrapper = await mountSuspended(AdminQuizQuestionPreview, {
      props: { question: predefined },
    })

    const answers = wrapper
      .findComponent(QuizPredefinedQuestion)
      .props('question').predefinedAnswers as { id: string }[]

    expect(answers.map((answer) => answer.id)).toEqual(['answer-0', 'answer-1'])
  })

  // Nothing in a preview may be answered.
  it('renders every question read-only', async () => {
    const wrapper = await mountSuspended(AdminQuizQuestionPreview, {
      props: { question: predefined },
    })

    expect(
      wrapper.findComponent(QuizPredefinedQuestion).props('readonly'),
    ).toBe(true)
  })
})
