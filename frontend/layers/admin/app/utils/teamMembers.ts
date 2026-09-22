interface HasUser {
  isTeamLead: boolean
  user: { id: string }
}

interface ScoreEntry {
  id: string
  score: number
  rank?: number | null
}

/**
 * Joins each member to their row in `Team.memberLeaderboard` and orders them
 * team lead first, then most points: the first row answers "who runs this
 * team", the order below it answers "who is contributing".
 */
export function rankTeamMembers<T extends HasUser>(
  members: readonly T[],
  leaderboard: readonly ScoreEntry[],
): Array<T & { score: number; rank: number | null }> {
  const byUser = new Map(leaderboard.map((entry) => [entry.id, entry]))

  return members
    .map((member) => {
      const entry = byUser.get(member.user.id)
      return {
        ...member,
        score: entry?.score ?? 0,
        rank: entry?.rank ?? null,
      }
    })
    .sort((a, b) => {
      if (a.isTeamLead !== b.isTeamLead) return a.isTeamLead ? -1 : 1
      return b.score - a.score
    })
}
