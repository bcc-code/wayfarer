import { describe, it, expect } from 'vitest'
import {
  LEADERBOARD_ENTITY_TYPE_LABELS,
  leaderboardFilterViewToInput,
  summarizeLeaderboardFilter,
} from '../../layers/admin/app/utils/leaderboardConfig'
import {
  ChurchCategory,
  Gender,
  LeaderboardEntityType,
} from '../../app/api/generated'

describe('leaderboardFilterViewToInput', () => {
  it('returns null for a config with no filter', () => {
    expect(leaderboardFilterViewToInput(null)).toBeNull()
    expect(leaderboardFilterViewToInput(undefined)).toBeNull()
  })

  it('returns null when every field is unset', () => {
    expect(
      leaderboardFilterViewToInput({
        __typename: 'LeaderboardFilterView',
        minScore: null,
        maxScore: null,
        churchId: null,
        country: null,
        churchCategory: null,
        gender: null,
        ageRange: null,
        teamId: null,
        superTeamId: null,
      }),
    ).toBeNull()
  })

  it('copies every field across', () => {
    expect(
      leaderboardFilterViewToInput({
        __typename: 'LeaderboardFilterView',
        minScore: 10,
        maxScore: 100,
        churchId: 'CH01',
        country: 'NO',
        churchCategory: ChurchCategory.Xl,
        gender: Gender.Female,
        ageRange: { __typename: 'AgeRange', min: 13, max: 18 },
        teamId: 'TM01',
        superTeamId: 'ST01',
      }),
    ).toEqual({
      minScore: 10,
      maxScore: 100,
      churchId: 'CH01',
      country: 'NO',
      churchCategory: ChurchCategory.Xl,
      gender: Gender.Female,
      ageRange: { min: 13, max: 18 },
      teamId: 'TM01',
      superTeamId: 'ST01',
    })
  })

  it('drops the __typename urql adds — the server rejects it on an input', () => {
    const input = leaderboardFilterViewToInput({
      __typename: 'LeaderboardFilterView',
      ageRange: { __typename: 'AgeRange', min: 13, max: 18 },
    })

    expect(input).not.toHaveProperty('__typename')
    expect(input?.ageRange).not.toHaveProperty('__typename')
  })

  // A reorder resends the whole config, so a zero that round-trips as "unset"
  // would silently widen the board it is reordering.
  it('keeps a zero score bound', () => {
    expect(
      leaderboardFilterViewToInput({
        __typename: 'LeaderboardFilterView',
        minScore: 0,
      }),
    ).toEqual({ minScore: 0 })
  })
})

describe('summarizeLeaderboardFilter', () => {
  it('is empty for no filter', () => {
    expect(summarizeLeaderboardFilter(null)).toEqual([])
  })

  it('describes the set fields only', () => {
    expect(
      summarizeLeaderboardFilter({
        __typename: 'LeaderboardFilterView',
        ageRange: { __typename: 'AgeRange', min: 13, max: 18 },
        gender: Gender.Male,
        churchCategory: null,
        teamId: 'TM01',
      }),
    ).toEqual(['13–18 år', 'Gutt', 'Ett lag'])
  })

  it('names the dimension for ID-valued fields rather than the thing', () => {
    expect(
      summarizeLeaderboardFilter({
        __typename: 'LeaderboardFilterView',
        churchId: 'CH01',
        superTeamId: 'ST01',
      }),
    ).toEqual(['Én menighet', 'Ett superlag'])
  })

  it('shows a zero score bound', () => {
    expect(
      summarizeLeaderboardFilter({
        __typename: 'LeaderboardFilterView',
        minScore: 0,
      }),
    ).toEqual(['Fra 0 p'])
  })
})

describe('LEADERBOARD_ENTITY_TYPE_LABELS', () => {
  it('labels every entity type the schema defines', () => {
    for (const type of Object.values(LeaderboardEntityType)) {
      expect(LEADERBOARD_ENTITY_TYPE_LABELS[type]).toBeTruthy()
    }
  })
})
