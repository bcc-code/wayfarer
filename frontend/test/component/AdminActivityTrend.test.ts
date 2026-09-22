// @vitest-environment nuxt
import { describe, it, expect } from 'vitest'
import { mountSuspended } from '@nuxt/test-utils/runtime'
import AdminActivityTrend from '../../layers/admin/app/components/admin/AdminActivityTrend.vue'

const trend = [
  { date: '2026-09-15', points: 0, activeUsers: 0 },
  { date: '2026-09-16', points: 120, activeUsers: 12 },
  { date: '2026-09-17', points: 0, activeUsers: 0 },
]

describe('AdminActivityTrend', () => {
  // Points are additive; a distinct-users-per-day count is not. Summing the
  // second would count the same person once per day they appeared.
  it('sums points but averages active users', async () => {
    const wrapper = await mountSuspended(AdminActivityTrend, {
      props: { trend, days: 14 },
    })

    const tiles = wrapper.findAll('.grid > div')
    expect(
      tiles
        .find((tile) => tile.text().includes('Poeng siste 14 dager'))
        ?.text(),
    ).toContain('120')
    // (0 + 12 + 0) / 3 = 4, not 12.
    expect(
      tiles
        .find((tile) => tile.text().includes('Aktive deltakere per dag'))
        ?.text(),
    ).toContain('4')
  })

  it('labels the window from the days prop', async () => {
    const wrapper = await mountSuspended(AdminActivityTrend, {
      props: { trend, days: 30 },
    })

    expect(wrapper.text()).toContain('Poeng siste 30 dager')
  })

  // The caller decides how to present "no data"; an empty trend renders as
  // nothing rather than as two zeroed tiles.
  it('renders nothing without trend data', async () => {
    const wrapper = await mountSuspended(AdminActivityTrend, {
      props: { trend: [] },
    })

    expect(wrapper.find('.grid').exists()).toBe(false)
  })
})
