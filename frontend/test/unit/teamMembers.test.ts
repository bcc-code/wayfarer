import { describe, it, expect } from 'vitest'
import { rankTeamMembers } from '../../layers/admin/app/utils/teamMembers'

const member = (id: string, isTeamLead = false) => ({
  id: `TM-${id}`,
  name: id,
  isTeamLead,
  user: { id },
})

describe('rankTeamMembers', () => {
  it('puts the team lead first, then orders by points', () => {
    const ranked = rankTeamMembers(
      [member('a'), member('b'), member('lead', true), member('c')],
      [
        { id: 'a', score: 10, rank: 3 },
        { id: 'b', score: 120, rank: 1 },
        { id: 'c', score: 40, rank: 2 },
        { id: 'lead', score: 0, rank: 4 },
      ],
    )

    expect(ranked.map((m) => m.user.id)).toEqual(['lead', 'b', 'c', 'a'])
  })

  // A member who has scored nothing has no journal rows, so the leaderboard
  // may not mention them at all.
  it('scores a member missing from the leaderboard as zero', () => {
    const ranked = rankTeamMembers(
      [member('a'), member('b')],
      [{ id: 'b', score: 5 }],
    )

    expect(ranked.map((m) => [m.user.id, m.score, m.rank])).toEqual([
      ['b', 5, null],
      ['a', 0, null],
    ])
  })

  it('leaves the input array untouched', () => {
    const members = [member('a'), member('lead', true)]

    rankTeamMembers(members, [])

    expect(members.map((m) => m.user.id)).toEqual(['a', 'lead'])
  })

  it('handles an empty team', () => {
    expect(rankTeamMembers([], [{ id: 'a', score: 1 }])).toEqual([])
  })
})
