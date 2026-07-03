// @vitest-environment nuxt
import { describe, it, expect, beforeEach, vi } from 'vitest'
import { ref } from 'vue'
import { mountSuspended, mockNuxtImport } from '@nuxt/test-utils/runtime'
import StandingsGlobal from '~/components/standings/StandingsGlobal.vue'
import StandingsListSkeleton from '~/components/standings/StandingsListSkeleton.vue'
import LeaderboardList from '~/components/leaderboard/LeaderboardList.vue'
import ErrorState from '~/components/ErrorState.vue'
import EmptyState from '~/components/EmptyState.vue'

// These composables are Nuxt auto-imports; replace them so we can drive the
// component through each of its render states without a real GraphQL backend.
const { queryMock, authMock, authReadyMock } = vi.hoisted(() => ({
  queryMock: vi.fn(),
  authMock: vi.fn(),
  authReadyMock: vi.fn(),
}))
mockNuxtImport('useStandingsGlobalPageQuery', () => queryMock)
mockNuxtImport('useAuth', () => authMock)
mockNuxtImport('useAuthReady', () => authReadyMock)

type QueryState = { data?: unknown; error?: unknown; fetching?: boolean }

function makeData(nodes: Array<{ id: string; name: string }>, me = null) {
  return {
    myCurrentProject: {
      leaderboard: {
        edges: nodes.map((node) => ({ node })),
        me,
      },
    },
  }
}

function mountWith(state: QueryState) {
  authMock.mockReturnValue({ me: ref({ age: 15 }) })
  authReadyMock.mockReturnValue({ isAuthReady: ref(true) })
  queryMock.mockReturnValue({
    data: ref(state.data ?? null),
    error: ref(state.error ?? null),
    fetching: ref(state.fetching ?? false),
  })
  // All child components render for real; only the data composables are mocked.
  return mountSuspended(StandingsGlobal)
}

describe('StandingsGlobal', () => {
  beforeEach(() => {
    vi.clearAllMocks()
  })

  it('shows the loading skeleton on initial load (fetching, no data)', async () => {
    const wrapper = await mountWith({ fetching: true, data: null })

    expect(wrapper.findComponent(StandingsListSkeleton).exists()).toBe(true)
    expect(wrapper.findComponent(LeaderboardList).exists()).toBe(false)
  })

  it('shows the error state when the query errors', async () => {
    const wrapper = await mountWith({ error: new Error('boom') })

    expect(wrapper.findComponent(ErrorState).exists()).toBe(true)
    expect(wrapper.findComponent(StandingsListSkeleton).exists()).toBe(false)
  })

  it('shows the empty state when there are no entries', async () => {
    const wrapper = await mountWith({ data: makeData([]) })

    expect(wrapper.findComponent(EmptyState).exists()).toBe(true)
    expect(wrapper.findComponent(LeaderboardList).exists()).toBe(false)
  })

  it('renders the leaderboard with the mapped entry nodes', async () => {
    const wrapper = await mountWith({
      data: makeData([
        { id: 'US01', name: 'Alice' },
        { id: 'US02', name: 'Bob' },
      ]),
    })

    const list = wrapper.findComponent(LeaderboardList)
    expect(list.exists()).toBe(true)
    expect(list.props('leaderboard')).toHaveLength(2)
    expect(list.props('leaderboard')[0]).toMatchObject({ name: 'Alice' })
  })

  it('passes the current user as an extra item when not in the top list', async () => {
    const me = { id: 'US99', name: 'Me' }
    const wrapper = await mountWith({
      data: makeData([{ id: 'US01', name: 'Alice' }], me),
    })

    const list = wrapper.findComponent(LeaderboardList)
    expect(list.props('extraItems')).toEqual([me])
  })
})
