import { describe, it, expect } from 'vitest'
import {
  findTeamLeader,
  findMemberById,
  isMemberTeamLead,
  getNonLeaderMembers,
} from '../../app/utils/teams'
import { LeaderboardEntryTag } from '../../app/api/generated'
import type { LeaderboardEntryLike } from '../../app/utils/leaderboard'

describe('teams', () => {
  const createMember = (
    id: string,
    tags?: LeaderboardEntryTag[],
  ): LeaderboardEntryLike => ({
    id,
    tags: tags ?? null,
  })

  describe('findTeamLeader', () => {
    it('should find member with TeamLead tag', () => {
      const leader = createMember('2', [LeaderboardEntryTag.TeamLead])
      const members = [createMember('1'), leader, createMember('3')]

      expect(findTeamLeader(members)).toBe(leader)
    })

    it('should return undefined when no team leader exists', () => {
      const members = [createMember('1'), createMember('2')]

      expect(findTeamLeader(members)).toBeUndefined()
    })

    it('should return undefined for empty array', () => {
      expect(findTeamLeader([])).toBeUndefined()
    })

    it('should find team leader with multiple tags', () => {
      const leader = createMember('1', [
        LeaderboardEntryTag.Me,
        LeaderboardEntryTag.TeamLead,
      ])
      const members = [leader, createMember('2')]

      expect(findTeamLeader(members)).toBe(leader)
    })
  })

  describe('findMemberById', () => {
    it('should find member by id', () => {
      const member = createMember('2')
      const members = [createMember('1'), member, createMember('3')]

      expect(findMemberById(members, '2')).toBe(member)
    })

    it('should return undefined when not found', () => {
      const members = [createMember('1'), createMember('2')]

      expect(findMemberById(members, '3')).toBeUndefined()
    })

    it('should return undefined for empty array', () => {
      expect(findMemberById([], '1')).toBeUndefined()
    })
  })

  describe('isMemberTeamLead', () => {
    it('should return true when member has TeamLead tag', () => {
      const member = createMember('1', [LeaderboardEntryTag.TeamLead])

      expect(isMemberTeamLead(member)).toBe(true)
    })

    it('should return false when member does not have TeamLead tag', () => {
      const member = createMember('1', [LeaderboardEntryTag.Me])

      expect(isMemberTeamLead(member)).toBe(false)
    })

    it('should return false when tags is null', () => {
      const member = createMember('1')

      expect(isMemberTeamLead(member)).toBe(false)
    })

    it('should return false when tags is empty', () => {
      const member = createMember('1', [])

      expect(isMemberTeamLead(member)).toBe(false)
    })

    it('should return true when TeamLead is among other tags', () => {
      const member = createMember('1', [
        LeaderboardEntryTag.Admin,
        LeaderboardEntryTag.TeamLead,
      ])

      expect(isMemberTeamLead(member)).toBe(true)
    })
  })

  describe('getNonLeaderMembers', () => {
    it('should filter out team leaders', () => {
      const leader = createMember('1', [LeaderboardEntryTag.TeamLead])
      const regular1 = createMember('2')
      const regular2 = createMember('3', [LeaderboardEntryTag.Me])
      const members = [leader, regular1, regular2]

      const result = getNonLeaderMembers(members)

      expect(result).toEqual([regular1, regular2])
    })

    it('should return all members when no leader exists', () => {
      const members = [createMember('1'), createMember('2')]

      const result = getNonLeaderMembers(members)

      expect(result).toEqual(members)
    })

    it('should return empty array when all are leaders', () => {
      const members = [
        createMember('1', [LeaderboardEntryTag.TeamLead]),
        createMember('2', [LeaderboardEntryTag.TeamLead]),
      ]

      expect(getNonLeaderMembers(members)).toEqual([])
    })

    it('should return empty array for empty input', () => {
      expect(getNonLeaderMembers([])).toEqual([])
    })
  })
})
