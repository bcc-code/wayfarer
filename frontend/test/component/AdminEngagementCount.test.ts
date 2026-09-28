// @vitest-environment nuxt
import { describe, it, expect } from 'vitest'
import { mountSuspended } from '@nuxt/test-utils/runtime'
import AdminEngagementCount from '../../layers/admin/app/components/admin/AdminEngagementCount.vue'

describe('AdminEngagementCount', () => {
  it('shows the count and its share of participants', async () => {
    const wrapper = await mountSuspended(AdminEngagementCount, {
      props: { count: 42, total: 88 },
    })

    expect(wrapper.text()).toContain('42')
    expect(wrapper.text()).toContain('av 88')
    expect(wrapper.text()).toContain('48 %')
  })

  // The count is the fact; the share is context that may not have loaded yet.
  it('shows the count alone without a total', async () => {
    const wrapper = await mountSuspended(AdminEngagementCount, {
      props: { count: 42 },
    })

    expect(wrapper.text()).toContain('42')
    expect(wrapper.text()).not.toContain('av')
  })

  it('omits the share when there are no participants', async () => {
    const wrapper = await mountSuspended(AdminEngagementCount, {
      props: { count: 0, total: 0, bar: true },
    })

    expect(wrapper.text()).not.toContain('%')
    expect(wrapper.find('.bg-primary').exists()).toBe(false)
  })

  // Awards can outnumber current participants — someone who earned it may have
  // left the project since.
  it('caps the bar at 100%', async () => {
    const wrapper = await mountSuspended(AdminEngagementCount, {
      props: { count: 120, total: 88, bar: true },
    })

    expect(wrapper.find('.bg-primary').attributes('style')).toContain(
      'width: 100%',
    )
  })

  it('draws no bar unless asked', async () => {
    const wrapper = await mountSuspended(AdminEngagementCount, {
      props: { count: 42, total: 88 },
    })

    expect(wrapper.find('.bg-primary').exists()).toBe(false)
  })
})
