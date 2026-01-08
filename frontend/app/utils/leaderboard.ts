import { LeaderboardEntryTag } from '~/api/generated'

/**
 * Pure functions for leaderboard data manipulation.
 * These functions are stateless and can be easily unit tested.
 */

export interface LeaderboardEntryLike {
  id?: string
  tags?: LeaderboardEntryTag[] | null
}

/**
 * Check if an entry represents the current user
 */
export function isCurrentUser(entry: LeaderboardEntryLike): boolean {
  return entry.tags?.includes(LeaderboardEntryTag.Me) ?? false
}

/**
 * Check if an entry is a team lead
 */
export function isTeamLead(entry: LeaderboardEntryLike): boolean {
  return entry.tags?.includes(LeaderboardEntryTag.TeamLead) ?? false
}

/**
 * Check if an entry has admin tag
 */
export function isAdmin(entry: LeaderboardEntryLike): boolean {
  return entry.tags?.includes(LeaderboardEntryTag.Admin) ?? false
}

/**
 * Find an entry by ID in a list
 */
export function findEntryById<T extends LeaderboardEntryLike>(
  entries: T[],
  id: string,
): T | undefined {
  return entries.find((entry) => entry.id === id)
}

/**
 * Check if an entry exists in a list by ID
 */
export function entryExistsInList<T extends LeaderboardEntryLike>(
  entries: T[],
  entry: T | null | undefined,
): boolean {
  if (!entry || !entry.id) return false
  return entries.some((e) => e.id === entry.id)
}

/**
 * Get extra items (entries not in the main list).
 * Useful for showing "me" when not in top N.
 */
export function getExtraItems<T extends LeaderboardEntryLike>(
  mainList: T[],
  me: T | null | undefined,
): T[] {
  if (!me) return []
  if (entryExistsInList(mainList, me)) return []
  return [me]
}

/**
 * Extract the current user entry from a list
 */
export function findCurrentUserEntry<T extends LeaderboardEntryLike>(
  entries: T[],
): T | undefined {
  return entries.find(isCurrentUser)
}

/**
 * Extract the team lead entry from a list
 */
export function findTeamLeadEntry<T extends LeaderboardEntryLike>(
  entries: T[],
): T | undefined {
  return entries.find(isTeamLead)
}
