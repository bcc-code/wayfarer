// @vitest-environment nuxt
import { describe, it, expect } from 'vitest'
import { mountSuspended } from '@nuxt/test-utils/runtime'
import AdminAchievementBadges from '../../layers/admin/app/components/admin/AdminAchievementBadges.vue'

const achievements = [
  {
    id: 'AC1',
    name: 'Første steg',
    awardedUserCount: 12,
    imageCompletedObject: { url: 'https://example.test/first.png' },
  },
  { id: 'AC2', name: 'Halvveis', awardedUserCount: 70 },
  { id: 'AC3', name: 'Streakmaster', awardedUserCount: 0 },
]

const mount = (participants?: number) =>
  mountSuspended(AdminAchievementBadges, {
    props: {
      achievements,
      projectId: 'PR01ARZ3NDEKTSV4RRFFQ69G5FAV',
      participants,
    },
  })

describe('AdminAchievementBadges', () => {
  // The project's own order, as the achievements list defines it by drag.
  it('keeps the order it is given', async () => {
    const wrapper = await mount(88)

    expect(
      wrapper.findAll('li').map((li) => li.find('a').attributes('href')),
    ).toEqual([
      expect.stringContaining('AC1'),
      expect.stringContaining('AC2'),
      expect.stringContaining('AC3'),
    ])
  })

  it('puts the name, count and share in the hover label', async () => {
    const wrapper = await mount(88)

    expect(
      wrapper
        .findAll('[data-tooltip]')
        .map((el) => el.attributes('data-tooltip')),
    ).toEqual([
      'Første steg: 12 av 88 · 14 %',
      'Halvveis: 70 av 88 · 80 %',
      'Streakmaster: 0 av 88 · 0 %',
    ])
  })

  // The count is still a fact without a denominator.
  it('drops the share when participants are unknown', async () => {
    const wrapper = await mount()

    expect(wrapper.find('[data-tooltip]').attributes('data-tooltip')).toBe(
      'Første steg: 12',
    )
  })

  it('names each badge for screen readers', async () => {
    const wrapper = await mount(88)

    expect(wrapper.find('.sr-only').text()).toBe('Første steg')
  })
})
