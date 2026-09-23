// @vitest-environment nuxt
import { describe, it, expect } from 'vitest'
import { mountSuspended } from '@nuxt/test-utils/runtime'
import StandingsBoard from '../../layers/user/app/components/standings/StandingsBoard.vue'
import LeaderboardList from '../../layers/user/app/components/leaderboard/LeaderboardList.vue'
import EmptyState from '../../layers/user/app/components/EmptyState.vue'
import { LeaderboardEntryTag } from '../../app/api/generated'

const entry = (id: string, name: string, rank: number, tags = []) => ({
  id,
  name,
  description: '',
  score: 100 - rank,
  rank,
  tags,
})

const mount = (board: Record<string, unknown>) =>
  mountSuspended(StandingsBoard, { props: { board } })

const boardWith = (
  nodes: ReturnType<typeof entry>[],
  me: ReturnType<typeof entry> | null = null,
) => ({
  id: 'LC1',
  name: 'Topp 20',
  leaderboard: {
    totalCount: nodes.length,
    edges: nodes.map((node) => ({ node })),
    me,
  },
})

describe('StandingsBoard', () => {
  it('names the board from its config', async () => {
    const wrapper = await mount(boardWith([entry('US1', 'Ada', 1)]))

    expect(wrapper.find('h2').text()).toBe('Topp 20')
  })

  it('renders the entries the server ranked', async () => {
    const wrapper = await mount(
      boardWith([entry('US1', 'Ada', 1), entry('US2', 'Linus', 2)]),
    )

    expect(
      wrapper.findComponent(LeaderboardList).props('leaderboard'),
    ).toHaveLength(2)
    expect(wrapper.findComponent(EmptyState).exists()).toBe(false)
  })

  // `me` is appended below the cut only when the user is not already in the
  // list — a board can legitimately rank the viewer outside the top N.
  it('appends the viewer when they fall outside the board', async () => {
    const me = entry('US9', 'Deg', 42, [LeaderboardEntryTag.Me] as never)
    const wrapper = await mount(boardWith([entry('US1', 'Ada', 1)], me))

    expect(wrapper.findComponent(LeaderboardList).props('extraItems')).toEqual([
      me,
    ])
  })

  it('does not repeat the viewer when they are already listed', async () => {
    const me = entry('US1', 'Ada', 1, [LeaderboardEntryTag.Me] as never)
    const wrapper = await mount(boardWith([me], me))

    expect(wrapper.findComponent(LeaderboardList).props('extraItems')).toEqual(
      [],
    )
  })

  it('shows an empty state for a board with no entries', async () => {
    const wrapper = await mount(boardWith([]))

    expect(wrapper.findComponent(EmptyState).exists()).toBe(true)
    expect(wrapper.findComponent(LeaderboardList).exists()).toBe(false)
  })
})
