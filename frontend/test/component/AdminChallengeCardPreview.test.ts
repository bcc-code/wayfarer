// @vitest-environment nuxt
import { describe, it, expect } from 'vitest'
import { mountSuspended } from '@nuxt/test-utils/runtime'
import { ChallengeType } from '../../app/api/generated'
import AdminChallengeCardPreview from '../../layers/admin/app/components/admin/challenge/AdminChallengeCardPreview.vue'
import ChallengeCard from '../../layers/user/app/components/challenges/ChallengeCard.vue'

const draft = {
  type: ChallengeType.Simple,
  name: 'Les Matteus 5',
  description: '<p>Bergprekenen</p>',
  buttonText: 'Start',
}

/**
 * The preview renders the participant app's own card. A replica would drift
 * from it, which is what this file is here to prevent.
 */
describe('AdminChallengeCardPreview', () => {
  it('renders the real challenge card', async () => {
    const wrapper = await mountSuspended(AdminChallengeCardPreview, {
      props: { challenge: draft },
    })

    expect(wrapper.findComponent(ChallengeCard).exists()).toBe(true)
    expect(wrapper.text()).toContain('Les Matteus 5')
    expect(wrapper.text()).toContain('Bergprekenen')
    expect(wrapper.text()).toContain('Start')
  })

  // The form holds a URL; the card takes the image object the query returns.
  it('hands the card an image object, or nothing at all', async () => {
    const withImage = await mountSuspended(AdminChallengeCardPreview, {
      props: { challenge: { ...draft, image: '/kart.png' } },
    })
    expect(
      withImage.findComponent(ChallengeCard).props('challenge').imageObject,
    ).toEqual({ url: '/kart.png' })

    const without = await mountSuspended(AdminChallengeCardPreview, {
      props: { challenge: draft },
    })
    expect(
      without.findComponent(ChallengeCard).props('challenge').imageObject,
    ).toBeNull()
  })

  // The card branches on `__typename`; an external challenge links out.
  it('tells the card which kind of challenge it is', async () => {
    const wrapper = await mountSuspended(AdminChallengeCardPreview, {
      props: {
        challenge: {
          ...draft,
          type: ChallengeType.External,
          url: 'https://bcc.media',
        },
      },
    })

    expect(
      wrapper.findComponent(ChallengeCard).props('challenge').__typename,
    ).toBe('ExternalChallenge')
  })
})
