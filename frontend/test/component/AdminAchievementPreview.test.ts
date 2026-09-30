// @vitest-environment nuxt
import { describe, it, expect } from 'vitest'
import { mountSuspended } from '@nuxt/test-utils/runtime'
import AdminAchievementPreview from '../../layers/admin/app/components/admin/achievement/AdminAchievementPreview.vue'
import AchievementDetails from '../../layers/user/app/components/achievements/AchievementDetails.vue'

const draft = {
  name: 'Leseplan uke 1',
  descriptionPending: 'Les hele uke 1',
  descriptionCompleted: 'Gratulerer!',
  imagePending: '/pending.png',
  imageCompleted: '/completed.png',
  points: 50,
}

const mount = (props: Record<string, unknown>) =>
  mountSuspended(AdminAchievementPreview, { props })

/**
 * The preview renders the participant app's own `AchievementDetails`, so the
 * two cannot drift. What lives here is the mapping from a draft.
 */
describe('AdminAchievementPreview', () => {
  it('shows the pending copy and image until it is earned', async () => {
    const wrapper = await mount({ achievement: draft })
    const shown = wrapper.findComponent(AchievementDetails).props('achievement')

    expect(shown.achievedAt).toBeNull()
    expect(wrapper.text()).toContain('Les hele uke 1')
    expect(wrapper.text()).not.toContain('Gratulerer!')
  })

  // A draft has never been earned, so the form's switcher decides the state.
  it('switches to the completed copy on request', async () => {
    const wrapper = await mount({ achievement: draft, achieved: true })
    const shown = wrapper.findComponent(AchievementDetails).props('achievement')

    expect(shown.achievedAt).toBeTruthy()
    expect(wrapper.text()).toContain('Gratulerer!')
  })

  // The form holds URLs; the app takes the image objects its query returns.
  it('hands over image objects, or nothing at all', async () => {
    const wrapper = await mount({ achievement: draft })
    const shown = wrapper.findComponent(AchievementDetails).props('achievement')

    expect(shown.imagePendingObject).toEqual({ url: '/pending.png' })
    expect(shown.imageCompletedObject).toEqual({ url: '/completed.png' })

    const bare = await mount({ achievement: { name: 'Uten bilde' } })
    expect(
      bare.findComponent(AchievementDetails).props('achievement')
        .imagePendingObject,
    ).toBeNull()
  })
})
