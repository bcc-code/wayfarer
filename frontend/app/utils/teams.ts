import { LeaderboardEntryTag } from '~/api/generated'
import type { LeaderboardEntryLike } from './leaderboard'

/**
 * Pure functions for team-related data manipulation.
 * These functions are stateless and can be easily unit tested.
 */

/**
 * Find the team leader from a list of members
 */
export function findTeamLeader<T extends LeaderboardEntryLike>(
  members: T[],
): T | undefined {
  return members.find((member) =>
    member.tags?.includes(LeaderboardEntryTag.TeamLead),
  )
}

/**
 * Find a team member by ID
 */
export function findMemberById<T extends LeaderboardEntryLike>(
  members: T[],
  id: string,
): T | undefined {
  return members.find((member) => member.id === id)
}

/**
 * Check if a member is the team leader
 */
export function isMemberTeamLead<T extends LeaderboardEntryLike>(
  member: T,
): boolean {
  return member.tags?.includes(LeaderboardEntryTag.TeamLead) ?? false
}

/**
 * Get all non-leader members
 */
export function getNonLeaderMembers<T extends LeaderboardEntryLike>(
  members: T[],
): T[] {
  return members.filter(
    (member) => !member.tags?.includes(LeaderboardEntryTag.TeamLead),
  )
}
