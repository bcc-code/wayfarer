// @vitest-environment nuxt
import { describe, it, expect } from 'vitest'
import { mountSuspended } from '@nuxt/test-utils/runtime'
import AdminProjectHomePreview from '../../layers/admin/app/components/admin/project/AdminProjectHomePreview.vue'
import ProfileProjectCard from '../../layers/user/app/components/profile/ProfileProjectCard.vue'
import ProjectInfoBanner from '../../layers/user/app/components/project/ProjectInfoBanner.vue'

const mount = (props: Record<string, unknown> = {}) =>
  mountSuspended(AdminProjectHomePreview, { props })

describe('AdminProjectHomePreview', () => {
  it('renders the app’s own project card', async () => {
    const wrapper = await mount({ projectName: 'Spring Revival 2025' })

    expect(wrapper.findComponent(ProfileProjectCard).exists()).toBe(true)
  })

  it('shows the banner image once one is chosen', async () => {
    const withBanner = await mount({ banner: '/banner.png' })
    expect(
      withBanner.findComponent(ProfileProjectCard).props('banner'),
    ).toEqual({ url: '/banner.png' })

    const without = await mount({})
    expect(without.findComponent(ProfileProjectCard).props('banner')).toBeNull()
  })

  it('leaves the info banner out when there is no message', async () => {
    const wrapper = await mount({})

    expect(wrapper.findComponent(ProjectInfoBanner).exists()).toBe(false)
  })

  // The real banner hides itself once dismissed, keyed by project id. A
  // preview that renders nothing would answer no question.
  it('shows the info message under an id of its own', async () => {
    const wrapper = await mount({
      infoMessage: { markdown: 'Husk leiren!', html: '<p>Husk leiren!</p>' },
    })

    const banner = wrapper.findComponent(ProjectInfoBanner)
    expect(banner.props('projectId')).toBe('preview')
    expect(banner.props('infoMessageStart')).toBeUndefined()
    expect(wrapper.text()).toContain('Husk leiren!')
  })
})
