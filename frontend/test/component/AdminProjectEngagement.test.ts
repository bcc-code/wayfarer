// @vitest-environment nuxt
import { describe, it, expect } from 'vitest'
import { mountSuspended, mockNuxtImport } from '@nuxt/test-utils/runtime'
import { ref } from 'vue'
import AdminProjectEngagement from '../../layers/admin/app/components/admin/project/AdminProjectEngagement.vue'

const engagement = {
  challenges: {
    edges: [
      { node: { id: 'CL1', name: 'Les Matteus 1', completionCount: 4 } },
      { node: { id: 'CL2', name: 'Kveldssamling', completionCount: 61 } },
      { node: { id: 'CL3', name: 'Morgenbønn', completionCount: 22 } },
    ],
  },
  achievements: {
    edges: [
      {
        node: {
          id: 'AC1',
          name: 'Første steg',
          awardedUserCount: 12,
          imageCompletedObject: { url: 'https://example.test/first.png' },
        },
      },
      {
        node: {
          id: 'AC2',
          name: 'Halvveis',
          awardedUserCount: 70,
          imageCompletedObject: null,
        },
      },
    ],
  },
}

const data = ref<typeof engagement | undefined>(engagement)
const fetching = ref(false)

mockNuxtImport('useAdminProjectEngagementQuery', () => () => ({
  data,
  fetching,
  error: ref(undefined),
}))

mockNuxtImport('useAuthReady', () => () => ({ isAuthReady: ref(true) }))

// UTooltip teleports its content and needs UApp's provider, so it is stubbed —
// but its `text` is where each row's name and share live, so the stub keeps it
// readable.
const UTooltip = {
  props: ['text'],
  template: '<div :data-tooltip="text"><slot /></div>',
}

const mount = () =>
  mountSuspended(AdminProjectEngagement, {
    props: { projectId: 'PR01ARZ3NDEKTSV4RRFFQ69G5FAV', participants: 88 },
    global: { stubs: { UTooltip } },
  })

describe('AdminProjectEngagement', () => {
  // Achievements first, then challenges. Most-completed first within each, so
  // the tail answers "what is nobody doing".
  it('orders both lists by count, descending', async () => {
    const wrapper = await mount()

    const lists = wrapper.findAll('ul')
    expect(lists).toHaveLength(2)
    expect(lists[0]!.findAll('li').map((li) => li.text())).toEqual([
      expect.stringContaining('Halvveis'),
      expect.stringContaining('Første steg'),
    ])
    expect(lists[1]!.findAll('li').map((li) => li.text())).toEqual([
      expect.stringContaining('Kveldssamling'),
      expect.stringContaining('Morgenbønn'),
      expect.stringContaining('Les Matteus 1'),
    ])
  })

  // The denominator is stated once for the panel, not repeated per row.
  it('names the participant total once', async () => {
    const wrapper = await mount()

    expect(wrapper.text()).toContain('Andel av 88 deltakere')
    expect(wrapper.findAll('[data-tooltip]')).toHaveLength(5)
  })

  it('puts the count and share in each hover label', async () => {
    const wrapper = await mount()

    const labels = wrapper
      .findAll('[data-tooltip]')
      .map((el) => el.attributes('data-tooltip'))
    expect(labels).toContain('Kveldssamling: 61 av 88 · 69 %')
    expect(labels).toContain('Halvveis: 70 av 88 · 80 %')
  })

  it('links each row to the item it counts', async () => {
    const wrapper = await mount()

    expect(wrapper.findAll('ul')[0]!.find('a').attributes('href')).toContain(
      'AC2',
    )
    expect(wrapper.findAll('ul')[1]!.find('a').attributes('href')).toContain(
      'CL2',
    )
  })

  // The placeholder PNG is a white disc, which swallowed the ring. An
  // achievement without an image gets the ring's own icon fallback instead.
  it('passes only real images to the ring', async () => {
    const wrapper = await mount()

    const images = wrapper.findAll('ul')[0]!.findAll('img')
    expect(images.map((img) => img.attributes('src'))).toEqual([
      'https://example.test/first.png',
    ])
  })

  it('reports an empty project per list', async () => {
    data.value = {
      challenges: { edges: [] },
      achievements: engagement.achievements,
    }
    const wrapper = await mount()

    expect(wrapper.text()).toContain('Ingen utfordringer ennå')
    expect(wrapper.text()).not.toContain('Ingen utmerkelser ennå')

    data.value = engagement
  })
})
