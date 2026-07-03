// @vitest-environment nuxt
import { describe, it, expect, beforeEach, vi } from 'vitest'
import { ref } from 'vue'
import { mountSuspended, mockNuxtImport } from '@nuxt/test-utils/runtime'
import StandingsUnit from '~/components/standings/StandingsUnit.vue'
import StandingsListSkeleton from '~/components/standings/StandingsListSkeleton.vue'
import LeaderboardList from '~/components/leaderboard/LeaderboardList.vue'
import ErrorState from '~/components/ErrorState.vue'
import { LeaderboardEntryTag } from '~/api/generated'

const {
  queryMock,
  authMock,
  authReadyMock,
  updateTeamMock,
  assignLeadMock,
  nowMock,
} = vi.hoisted(() => ({
  queryMock: vi.fn(),
  authMock: vi.fn(),
  authReadyMock: vi.fn(),
  updateTeamMock: vi.fn(),
  assignLeadMock: vi.fn(),
  nowMock: vi.fn(),
}))
mockNuxtImport('useStandingsUnitPageQuery', () => queryMock)
mockNuxtImport('useAuth', () => authMock)
mockNuxtImport('useAuthReady', () => authReadyMock)
mockNuxtImport('useUpdateTeamMutation', () => updateTeamMock)
mockNuxtImport('useAssignTeamLeadMutation', () => assignLeadMock)
mockNuxtImport('useNow', () => nowMock)

type Member = { id: string; name: string; rank?: number; tags?: string[] }
type QueryState = {
  data?: unknown
  error?: unknown
  fetching?: boolean
  isTeamLead?: boolean
  meId?: string
}

function makeData(team: { name?: string; members?: Member[] } | null) {
  return {
    myCurrentProject: {
      myTeam: team
        ? {
            id: 'TM01',
            name: team.name ?? 'Team A',
            superTeam: null,
            memberLeaderboard: team.members ?? [],
          }
        : null,
    },
  }
}

function mountWith(state: QueryState) {
  authMock.mockReturnValue({
    isTeamLead: ref(state.isTeamLead ?? false),
    me: ref({ id: state.meId ?? 'US99' }),
  })
  authReadyMock.mockReturnValue({ isAuthReady: ref(true) })
  queryMock.mockReturnValue({
    data: ref(state.data ?? null),
    error: ref(state.error ?? null),
    fetching: ref(state.fetching ?? false),
    executeQuery: vi.fn(),
  })
  updateTeamMock.mockReturnValue({ executeMutation: vi.fn() })
  assignLeadMock.mockReturnValue({ executeMutation: vi.fn() })
  nowMock.mockReturnValue(ref(new Date('2020-01-01T00:00:00Z')))
  // All child components render for real; only the data composables are mocked.
  return mountSuspended(StandingsUnit)
}

describe('StandingsUnit', () => {
  beforeEach(() => {
    vi.clearAllMocks()
  })

  it('shows the loading skeleton on initial load', async () => {
    const wrapper = await mountWith({ fetching: true, data: null })

    expect(wrapper.findComponent(StandingsListSkeleton).exists()).toBe(true)
  })

  it('shows the error state when the query errors', async () => {
    const wrapper = await mountWith({ error: new Error('boom') })

    expect(wrapper.findComponent(ErrorState).exists()).toBe(true)
  })

  it('renders the team name and member leaderboard', async () => {
    const wrapper = await mountWith({
      data: makeData({
        name: 'The Eagles',
        members: [{ id: 'US01', name: 'Alice', rank: 1 }],
      }),
    })

    expect(wrapper.text()).toContain('The Eagles')
    const list = wrapper.findComponent(LeaderboardList)
    expect(list.exists()).toBe(true)
    expect(list.props('leaderboard')[0]).toMatchObject({ name: 'Alice' })
  })

  describe('badge prop', () => {
    it('labels team-lead members and leaves others unbadged', async () => {
      const wrapper = await mountWith({
        data: makeData({
          members: [
            { id: 'US01', name: 'Lead', tags: [LeaderboardEntryTag.TeamLead] },
            { id: 'US02', name: 'Member', tags: [] },
          ],
        }),
      })

      const badge = wrapper.findComponent(LeaderboardList).props('badge')
      expect(badge({ tags: [LeaderboardEntryTag.TeamLead] })).toBeTruthy()
      expect(badge({ tags: [] })).toBeUndefined()
    })
  })

  describe('shouldHideScore prop (non team-lead viewer)', () => {
    it('hides scores below the top 3 for other members, but not your own', async () => {
      const wrapper = await mountWith({
        meId: 'US99',
        isTeamLead: false,
        data: makeData({ members: [{ id: 'US01', name: 'Alice' }] }),
      })

      const hide = wrapper
        .findComponent(LeaderboardList)
        .props('shouldHideScore')

      // rank > 3 and not the viewer -> hidden
      expect(hide({ id: 'US01', rank: 5 })).toBe(true)
      // top 3 -> visible
      expect(hide({ id: 'US01', rank: 2 })).toBe(false)
      // the viewer's own entry -> always visible
      expect(hide({ id: 'US99', rank: 10 })).toBe(false)
    })

    it('never hides scores when the viewer is a team lead', async () => {
      const wrapper = await mountWith({
        isTeamLead: true,
        data: makeData({ members: [{ id: 'US01', name: 'Alice' }] }),
      })

      const hide = wrapper
        .findComponent(LeaderboardList)
        .props('shouldHideScore')
      expect(hide({ id: 'US01', rank: 10 })).toBe(false)
    })
  })
})
