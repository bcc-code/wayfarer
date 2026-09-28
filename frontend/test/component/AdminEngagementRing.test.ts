// @vitest-environment nuxt
import { describe, it, expect } from 'vitest'
import { mountSuspended } from '@nuxt/test-utils/runtime'
import AdminEngagementRing from '../../layers/admin/app/components/admin/AdminEngagementRing.vue'

/** The arc circle is the second one; the first is the track. */
const arc = (wrapper: { findAll: (s: string) => unknown[] }) =>
  wrapper.findAll('circle')[1] as
    { attributes: (name: string) => string | undefined } | undefined

describe('AdminEngagementRing', () => {
  it('draws the arc as the share of participants', async () => {
    const wrapper = await mountSuspended(AdminEngagementRing, {
      props: { count: 22, total: 88, size: 44 },
    })

    const circumference = 2 * Math.PI * ((44 - 3) / 2)
    // A quarter earned it, so three quarters of the ring stays unpainted.
    expect(Number(arc(wrapper)!.attributes('stroke-dashoffset'))).toBeCloseTo(
      circumference * 0.75,
      5,
    )
    expect(wrapper.text()).toContain('22')
  })

  // Awards can outnumber current participants — someone who earned it may have
  // left the project since.
  it('caps the arc at a full ring', async () => {
    const wrapper = await mountSuspended(AdminEngagementRing, {
      props: { count: 120, total: 88 },
    })

    expect(Number(arc(wrapper)!.attributes('stroke-dashoffset'))).toBe(0)
  })

  it('leaves the ring empty at zero and without a total', async () => {
    const zero = await mountSuspended(AdminEngagementRing, {
      props: { count: 0, total: 88 },
    })
    expect(arc(zero)).toBeUndefined()

    const unknown = await mountSuspended(AdminEngagementRing, {
      props: { count: 5 },
    })
    expect(arc(unknown)).toBeUndefined()
    expect(unknown.text()).toContain('5')
  })

  it('renders no image when none is given', async () => {
    const wrapper = await mountSuspended(AdminEngagementRing, {
      props: { count: 1, total: 10 },
    })

    expect(wrapper.find('img').exists()).toBe(false)
  })
})
