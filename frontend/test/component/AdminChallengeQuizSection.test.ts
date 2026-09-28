// @vitest-environment nuxt
import { describe, it, expect } from 'vitest'
import { mountSuspended } from '@nuxt/test-utils/runtime'
import AdminChallengeQuizSection from '../../layers/admin/app/components/admin/challenge/AdminChallengeQuizSection.vue'

const quiz = {
  id: 'QZ1',
  questions: [{ id: 'Q1' }, { id: 'Q2' }, { id: 'Q3' }],
  completionPoints: 50,
  allowRetakes: true,
  randomizeQuestions: false,
  revealCorrectAnswers: true,
  timeoutSeconds: 30,
}

const ids = { projectId: 'PR1', challengeId: 'CL1' }

describe('AdminChallengeQuizSection', () => {
  // Labelled pairs, so each value says what it is a value of.
  it('summarises the quiz as labelled values', async () => {
    const wrapper = await mountSuspended(AdminChallengeQuizSection, {
      props: { ...ids, quiz },
    })

    const pairs = wrapper
      .findAll('dl > div')
      .map((row) => [row.find('dt').text(), row.find('dd').text()])

    expect(pairs).toEqual([
      ['Spørsmål', '3'],
      ['Poeng for fullføring', '50'],
      ['Forsøk', 'Flere tillatt'],
      ['Rekkefølge', 'Fast'],
      ['Riktige svar', 'Vises'],
      ['Tidsgrense', '30 sek per spørsmål'],
    ])
  })

  it('names the opposite of each setting when it is off', async () => {
    const wrapper = await mountSuspended(AdminChallengeQuizSection, {
      props: {
        ...ids,
        quiz: {
          ...quiz,
          allowRetakes: false,
          randomizeQuestions: true,
          revealCorrectAnswers: false,
          timeoutSeconds: null,
        },
      },
    })

    const values = wrapper.findAll('dl > div dd').map((dd) => dd.text())
    expect(values).toEqual(['3', '50', 'Ett', 'Tilfeldig', 'Skjules', 'Ingen'])
  })

  // A live challenge with nothing to answer is the state worth flagging.
  it('warns when the quiz has no questions', async () => {
    const wrapper = await mountSuspended(AdminChallengeQuizSection, {
      props: { ...ids, quiz: { ...quiz, questions: [] } },
    })

    expect(wrapper.text()).toContain('ingen spørsmål ennå')
  })

  // The quiz cannot exist before the challenge does.
  it('explains itself on a challenge that does not exist yet', async () => {
    const wrapper = await mountSuspended(AdminChallengeQuizSection, {
      props: {},
    })

    expect(wrapper.text()).toContain('etter at utfordringen er opprettet')
    expect(wrapper.find('a').exists()).toBe(false)
  })
})
