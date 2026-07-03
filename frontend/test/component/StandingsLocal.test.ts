// @vitest-environment nuxt
import { describe, it, expect, beforeEach, vi } from 'vitest'
import { ref } from 'vue'
import { mountSuspended, mockNuxtImport } from '@nuxt/test-utils/runtime'
import StandingsLocal from '~/components/standings/StandingsLocal.vue'
import StandingsListSkeleton from '~/components/standings/StandingsListSkeleton.vue'
import LeaderboardList from '~/components/leaderboard/LeaderboardList.vue'
import ErrorState from '~/components/ErrorState.vue'
import EmptyState from '~/components/EmptyState.vue'
import { LeaderboardEntityType } from '~/api/generated'

const { queryMock, authMock, authReadyMock, localStorageMock } = vi.hoisted(
  () => ({
    queryMock: vi.fn(),
    authMock: vi.fn(),
    authReadyMock: vi.fn(),
    localStorageMock: vi.fn(),
  }),
)
mockNuxtImport('useStandingsLocalPageQuery', () => queryMock)
mockNuxtImport('useAuth', () => authMock)
mockNuxtImport('useAuthReady', () => authReadyMock)
// entityType is persisted with useLocalStorage; return a ref we control.
mockNuxtImport('useLocalStorage', () => (_key: string, def: unknown) => {
  return localStorageMock(_key, def)
})

type Entry = { id: string; name: string }
type QueryState = {
  data?: unknown
  error?: unknown
  fetching?: boolean
  entityType?: LeaderboardEntityType
}

function makeData(opts: {
  churchName?: string | null
  persons?: Entry[]
  units?: Entry[]
  personsTotal?: number
  personsMe?: unknown
}) {
  return {
    me: { church: opts.churchName ? { name: opts.churchName } : null },
    myCurrentProject: {
      personLeaderboard: {
        edges: (opts.persons ?? []).map((node) => ({ node })),
        me: opts.personsMe ?? null,
        totalCount: opts.personsTotal ?? (opts.persons ?? []).length,
      },
      unitLeaderboard: {
        edges: (opts.units ?? []).map((node) => ({ node })),
        me: null,
      },
    },
  }
}

function mountWith(state: QueryState) {
  authMock.mockReturnValue({ me: ref({ church: { id: 'CH01' } }) })
  authReadyMock.mockReturnValue({ isAuthReady: ref(true) })
  localStorageMock.mockReturnValue(
    ref(state.entityType ?? LeaderboardEntityType.Persons),
  )
  queryMock.mockReturnValue({
    data: ref(state.data ?? null),
    error: ref(state.error ?? null),
    fetching: ref(state.fetching ?? false),
  })
  // All child components render for real; only the data composables are mocked.
  return mountSuspended(StandingsLocal)
}

describe('StandingsLocal', () => {
  beforeEach(() => {
    vi.clearAllMocks()
  })

  it('shows the loading skeleton while fetching', async () => {
    const wrapper = await mountWith({ fetching: true, data: null })

    expect(wrapper.findComponent(StandingsListSkeleton).exists()).toBe(true)
  })

  it('shows the error state when the query errors', async () => {
    const wrapper = await mountWith({ error: new Error('boom') })

    expect(wrapper.findComponent(ErrorState).exists()).toBe(true)
  })

  it('shows the empty state when there is no data', async () => {
    const wrapper = await mountWith({ data: null, fetching: false })

    expect(wrapper.findComponent(EmptyState).exists()).toBe(true)
  })

  it('renders the church name heading', async () => {
    const wrapper = await mountWith({
      data: makeData({ churchName: 'Oslo/Follo', persons: [] }),
    })

    expect(wrapper.text()).toContain('Oslo/Follo')
  })

  it('shows the person leaderboard when the Persons tab is active', async () => {
    const wrapper = await mountWith({
      entityType: LeaderboardEntityType.Persons,
      data: makeData({
        persons: [{ id: 'US01', name: 'Alice' }],
        units: [{ id: 'TM01', name: 'Team A' }],
      }),
    })

    const lists = wrapper.findAllComponents(LeaderboardList)
    expect(lists).toHaveLength(1)
    expect(lists[0]!.props('leaderboard')[0]).toMatchObject({ name: 'Alice' })
  })

  it('shows the unit leaderboard when the Teams tab is active', async () => {
    const wrapper = await mountWith({
      entityType: LeaderboardEntityType.Teams,
      data: makeData({
        persons: [{ id: 'US01', name: 'Alice' }],
        units: [{ id: 'TM01', name: 'Team A' }],
      }),
    })

    const lists = wrapper.findAllComponents(LeaderboardList)
    expect(lists).toHaveLength(1)
    expect(lists[0]!.props('leaderboard')[0]).toMatchObject({ name: 'Team A' })
  })

  it.each([
    { total: 0, expected: 20 },
    { total: 10, expected: 3 },
    { total: 30, expected: 10 },
    { total: 100, expected: 20 },
  ])(
    'computes the persons tab count as $expected for totalCount $total',
    async ({ total, expected }) => {
      const wrapper = await mountWith({
        data: makeData({
          persons: [{ id: 'US01', name: 'Alice' }],
          units: [{ id: 'TM01', name: 'Team A' }],
          personsTotal: total,
        }),
      })

      const tabs = wrapper.findComponent({ name: 'DesignTabs' }).props('tabs')
      // Persons tab label is $t('standings.top', { amount: totalPersons })
      expect(tabs[0].label).toContain(String(expected))
    },
  )
})
