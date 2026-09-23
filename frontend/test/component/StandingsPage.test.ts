// @vitest-environment nuxt
import { describe, it, expect, beforeEach, vi } from 'vitest'
import { ref } from 'vue'
import { mountSuspended, mockNuxtImport } from '@nuxt/test-utils/runtime'
import StandingsPage from '../../layers/user/app/pages/standings.vue'
import StandingsBoard from '../../layers/user/app/components/standings/StandingsBoard.vue'
import StandingsUnit from '../../layers/user/app/components/standings/StandingsUnit.vue'
import StandingsListSkeleton from '../../layers/user/app/components/standings/StandingsListSkeleton.vue'
import DesignTabs from '../../layers/user/app/components/design/DesignTabs.vue'
import ErrorState from '../../layers/user/app/components/ErrorState.vue'
import EmptyState from '../../layers/user/app/components/EmptyState.vue'

// These composables are Nuxt auto-imports; replace them so we can drive the
// page through each of its render states without a real GraphQL backend.
const { queryMock, unitQueryMock, authReadyMock, trackMock } = vi.hoisted(
  () => ({
    queryMock: vi.fn(),
    unitQueryMock: vi.fn(),
    authReadyMock: vi.fn(),
    trackMock: vi.fn(),
  }),
)
mockNuxtImport('useStandingsPageQuery', () => queryMock)
// StandingsUnit keeps its own query — it is the one tab that is not a config.
mockNuxtImport('useStandingsUnitPageQuery', () => unitQueryMock)
mockNuxtImport('useAuthReady', () => authReadyMock)
mockNuxtImport('useAnalytics', () => () => ({ track: trackMock }))

const board = (id: string, name: string, names: string[] = []) => ({
  id,
  name,
  leaderboard: {
    totalCount: names.length,
    edges: names.map((entryName, index) => ({
      node: {
        id: `US${index}`,
        name: entryName,
        description: '',
        score: 100 - index,
        rank: index + 1,
        tags: [],
      },
    })),
    me: null,
  },
})

function makeData(
  leaderboards: ReturnType<typeof board>[],
  myTeam: { id: string } | null = null,
) {
  return { myCurrentProject: { id: 'PR1', myTeam, leaderboards } }
}

function mountWith(state: {
  data?: unknown
  error?: unknown
  fetching?: boolean
}) {
  authReadyMock.mockReturnValue({ isAuthReady: ref(true) })
  queryMock.mockReturnValue({
    data: ref(state.data ?? null),
    error: ref(state.error ?? null),
    fetching: ref(state.fetching ?? false),
  })
  unitQueryMock.mockReturnValue({
    data: ref(null),
    error: ref(null),
    fetching: ref(false),
  })
  return mountSuspended(StandingsPage)
}

describe('standings page', () => {
  beforeEach(() => {
    vi.clearAllMocks()
    localStorage.clear()
  })

  it('shows the loading skeleton on initial load', async () => {
    const wrapper = await mountWith({ fetching: true, data: null })

    expect(wrapper.findComponent(StandingsListSkeleton).exists()).toBe(true)
    expect(wrapper.findComponent(StandingsBoard).exists()).toBe(false)
  })

  it('shows the error state when the query errors', async () => {
    const wrapper = await mountWith({ error: new Error('boom') })

    expect(wrapper.findComponent(ErrorState).exists()).toBe(true)
  })

  // Config-driven only: there is no fallback to the old hardcoded boards.
  it('shows an empty state for a project with no configs and no team', async () => {
    const wrapper = await mountWith({ data: makeData([]) })

    expect(wrapper.findComponent(EmptyState).exists()).toBe(true)
    expect(wrapper.findComponent(StandingsBoard).exists()).toBe(false)
    expect(wrapper.findComponent(DesignTabs).exists()).toBe(false)
  })

  it('renders one tab per config, in the order the server returns them', async () => {
    const wrapper = await mountWith({
      data: makeData([
        board('LC1', 'Topp 20'),
        board('LC2', 'Lokalt'),
        board('LC3', 'Units'),
      ]),
    })

    expect(
      wrapper.findComponent(DesignTabs).props('tabs') as { label: string }[],
    ).toEqual([
      { key: 'LC1', label: 'Topp 20', value: 'LC1' },
      { key: 'LC2', label: 'Lokalt', value: 'LC2' },
      { key: 'LC3', label: 'Units', value: 'LC3' },
    ])
  })

  // The unit tab reads myTeam.memberLeaderboard, so no config describes it.
  it('appends the unit tab last, only when the user has a team', async () => {
    const withoutTeam = await mountWith({
      data: makeData([board('LC1', 'Topp 20'), board('LC2', 'Lokalt')]),
    })
    expect(
      (
        withoutTeam.findComponent(DesignTabs).props('tabs') as {
          value: string
        }[]
      ).map((t) => t.value),
    ).toEqual(['LC1', 'LC2'])

    const withTeam = await mountWith({
      data: makeData([board('LC1', 'Topp 20')], { id: 'TM1' }),
    })
    expect(
      (
        withTeam.findComponent(DesignTabs).props('tabs') as { value: string }[]
      ).map((t) => t.value),
    ).toEqual(['LC1', 'unit'])
  })

  it('hides the tab bar when there is only one board', async () => {
    const wrapper = await mountWith({ data: makeData([board('LC1', 'Topp')]) })

    expect(wrapper.findComponent(DesignTabs).exists()).toBe(false)
    expect(wrapper.findComponent(StandingsBoard).exists()).toBe(true)
  })

  it('opens the first board by default', async () => {
    const wrapper = await mountWith({
      data: makeData([
        board('LC1', 'Topp 20', ['Ada']),
        board('LC2', 'Lokalt', ['Linus']),
      ]),
    })

    expect(wrapper.findComponent(StandingsBoard).props('board').id).toBe('LC1')
  })

  // A stored tab outlives the board it names — a deleted or deactivated config,
  // or a different project entirely.
  it('falls back to the first board when the stored tab no longer exists', async () => {
    localStorage.setItem('standings-tab', 'LC-GONE')

    const wrapper = await mountWith({
      data: makeData([board('LC1', 'Topp 20'), board('LC2', 'Lokalt')]),
    })

    expect(wrapper.findComponent(StandingsBoard).props('board').id).toBe('LC1')
  })

  it('opens the stored tab when it still exists', async () => {
    localStorage.setItem('standings-tab', 'LC2')

    const wrapper = await mountWith({
      data: makeData([board('LC1', 'Topp 20'), board('LC2', 'Lokalt')]),
    })

    expect(wrapper.findComponent(StandingsBoard).props('board').id).toBe('LC2')
  })

  it('switches to the unit component on the unit tab', async () => {
    localStorage.setItem('standings-tab', 'unit')

    const wrapper = await mountWith({
      data: makeData([board('LC1', 'Topp 20')], { id: 'TM1' }),
    })

    expect(wrapper.findComponent(StandingsUnit).exists()).toBe(true)
    expect(wrapper.findComponent(StandingsBoard).exists()).toBe(false)
  })
})
