import { describe, it, expect } from 'vitest'
import {
  isCurrentUser,
  isTeamLead,
  isAdmin,
  findEntryById,
  entryExistsInList,
  getExtraItems,
  getExtraItemsWithRivals,
  findCurrentUserEntry,
  findTeamLeadEntry,
  type LeaderboardEntryLike,
} from '../../app/utils/leaderboard'
import { LeaderboardEntryTag } from '../../app/api/generated'

describe('leaderboard', () => {
  const createEntry = (
    id: string,
    tags?: LeaderboardEntryTag[],
  ): LeaderboardEntryLike => ({
    id,
    tags: tags ?? null,
  })

  describe('isCurrentUser', () => {
    it('should return true when entry has Me tag', () => {
      const entry = createEntry('1', [LeaderboardEntryTag.Me])

      expect(isCurrentUser(entry)).toBe(true)
    })

    it('should return false when entry does not have Me tag', () => {
      const entry = createEntry('1', [LeaderboardEntryTag.Admin])

      expect(isCurrentUser(entry)).toBe(false)
    })

    it('should return false when tags is null', () => {
      const entry = createEntry('1')

      expect(isCurrentUser(entry)).toBe(false)
    })

    it('should return false when tags is empty', () => {
      const entry = createEntry('1', [])

      expect(isCurrentUser(entry)).toBe(false)
    })

    it('should return true when Me tag is among other tags', () => {
      const entry = createEntry('1', [
        LeaderboardEntryTag.Admin,
        LeaderboardEntryTag.Me,
      ])

      expect(isCurrentUser(entry)).toBe(true)
    })
  })

  describe('isTeamLead', () => {
    it('should return true when entry has TeamLead tag', () => {
      const entry = createEntry('1', [LeaderboardEntryTag.TeamLead])

      expect(isTeamLead(entry)).toBe(true)
    })

    it('should return false when entry does not have TeamLead tag', () => {
      const entry = createEntry('1', [LeaderboardEntryTag.Me])

      expect(isTeamLead(entry)).toBe(false)
    })

    it('should return false when tags is null', () => {
      const entry = createEntry('1')

      expect(isTeamLead(entry)).toBe(false)
    })
  })

  describe('isAdmin', () => {
    it('should return true when entry has Admin tag', () => {
      const entry = createEntry('1', [LeaderboardEntryTag.Admin])

      expect(isAdmin(entry)).toBe(true)
    })

    it('should return false when entry does not have Admin tag', () => {
      const entry = createEntry('1', [LeaderboardEntryTag.Me])

      expect(isAdmin(entry)).toBe(false)
    })

    it('should return false when tags is null', () => {
      const entry = createEntry('1')

      expect(isAdmin(entry)).toBe(false)
    })
  })

  describe('findEntryById', () => {
    it('should find entry by id', () => {
      const entries = [createEntry('1'), createEntry('2'), createEntry('3')]

      const result = findEntryById(entries, '2')

      expect(result).toBe(entries[1])
    })

    it('should return undefined when not found', () => {
      const entries = [createEntry('1'), createEntry('2')]

      expect(findEntryById(entries, '3')).toBeUndefined()
    })

    it('should return undefined for empty array', () => {
      expect(findEntryById([], '1')).toBeUndefined()
    })
  })

  describe('entryExistsInList', () => {
    it('should return true when entry exists in list', () => {
      const entries = [createEntry('1'), createEntry('2')]
      const entry = createEntry('2')

      expect(entryExistsInList(entries, entry)).toBe(true)
    })

    it('should return false when entry does not exist in list', () => {
      const entries = [createEntry('1'), createEntry('2')]
      const entry = createEntry('3')

      expect(entryExistsInList(entries, entry)).toBe(false)
    })

    it('should return false for null entry', () => {
      const entries = [createEntry('1')]

      expect(entryExistsInList(entries, null)).toBe(false)
    })

    it('should return false for undefined entry', () => {
      const entries = [createEntry('1')]

      expect(entryExistsInList(entries, undefined)).toBe(false)
    })

    it('should return false for entry without id', () => {
      const entries = [createEntry('1')]
      const entry: LeaderboardEntryLike = { tags: [] }

      expect(entryExistsInList(entries, entry)).toBe(false)
    })
  })

  describe('getExtraItems', () => {
    it('should return array with me when not in main list', () => {
      const mainList = [createEntry('1'), createEntry('2')]
      const me = createEntry('3', [LeaderboardEntryTag.Me])

      const result = getExtraItems(mainList, me)

      expect(result).toEqual([me])
    })

    it('should return empty array when me is in main list', () => {
      const me = createEntry('2', [LeaderboardEntryTag.Me])
      const mainList = [createEntry('1'), me]

      const result = getExtraItems(mainList, me)

      expect(result).toEqual([])
    })

    it('should return empty array when me is null', () => {
      const mainList = [createEntry('1')]

      expect(getExtraItems(mainList, null)).toEqual([])
    })

    it('should return empty array when me is undefined', () => {
      const mainList = [createEntry('1')]

      expect(getExtraItems(mainList, undefined)).toEqual([])
    })

    it('should work with empty main list', () => {
      const me = createEntry('1', [LeaderboardEntryTag.Me])

      const result = getExtraItems([], me)

      expect(result).toEqual([me])
    })
  })

  describe('findCurrentUserEntry', () => {
    it('should find entry with Me tag', () => {
      const meEntry = createEntry('2', [LeaderboardEntryTag.Me])
      const entries = [createEntry('1'), meEntry, createEntry('3')]

      expect(findCurrentUserEntry(entries)).toBe(meEntry)
    })

    it('should return undefined when no Me entry exists', () => {
      const entries = [createEntry('1'), createEntry('2')]

      expect(findCurrentUserEntry(entries)).toBeUndefined()
    })

    it('should return undefined for empty array', () => {
      expect(findCurrentUserEntry([])).toBeUndefined()
    })
  })

  describe('findTeamLeadEntry', () => {
    it('should find entry with TeamLead tag', () => {
      const leadEntry = createEntry('2', [LeaderboardEntryTag.TeamLead])
      const entries = [createEntry('1'), leadEntry, createEntry('3')]

      expect(findTeamLeadEntry(entries)).toBe(leadEntry)
    })

    it('should return undefined when no TeamLead entry exists', () => {
      const entries = [createEntry('1'), createEntry('2')]

      expect(findTeamLeadEntry(entries)).toBeUndefined()
    })

    it('should return undefined for empty array', () => {
      expect(findTeamLeadEntry([])).toBeUndefined()
    })
  })
})

describe('getExtraItemsWithRivals', () => {
  const ranked = (id: string, rank: number, tags?: LeaderboardEntryTag[]) =>
    ({ id, rank, tags }) as LeaderboardEntryLike & { rank: number }

  const top = [ranked('US1', 1), ranked('US2', 2)]

  it('is empty when there is no viewer entry', () => {
    expect(getExtraItemsWithRivals(top, null, [ranked('US8', 40)])).toEqual([])
  })

  // Rivals rank above the viewer, so if the viewer made the cut they did too —
  // appending them would repeat rows already on the board.
  it('is empty when the viewer is already on the board', () => {
    const me = ranked('US1', 1, [LeaderboardEntryTag.Me])

    expect(
      getExtraItemsWithRivals([me, top[1]!], me, [ranked('US0', 0)]),
    ).toEqual([])
  })

  it('appends the viewer alone when there are no rivals', () => {
    const me = ranked('US9', 41, [LeaderboardEntryTag.Me])

    expect(getExtraItemsWithRivals(top, me)).toEqual([me])
  })

  // The server walks backward from the viewer, so rivals arrive nearest-first.
  it('orders rivals by rank, with the viewer last', () => {
    const me = ranked('US9', 41, [LeaderboardEntryTag.Me])
    const nearest = ranked('US8', 40)
    const middle = ranked('US7', 35)
    const furthest = ranked('US6', 20)

    expect(
      getExtraItemsWithRivals(top, me, [nearest, middle, furthest]),
    ).toEqual([furthest, middle, nearest, me])
  })

  it('drops a rival that is already on the board', () => {
    const me = ranked('US9', 41, [LeaderboardEntryTag.Me])
    const onBoard = top[1]!

    expect(
      getExtraItemsWithRivals(top, me, [onBoard, ranked('US8', 40)]),
    ).toEqual([ranked('US8', 40), me])
  })

  it('puts an unranked rival last among the rivals', () => {
    const me = ranked('US9', 41, [LeaderboardEntryTag.Me])
    const unranked = { id: 'US5' } as LeaderboardEntryLike & { rank?: number }

    expect(
      getExtraItemsWithRivals(top, me, [unranked, ranked('US8', 40)]),
    ).toEqual([ranked('US8', 40), unranked, me])
  })
})
