import { describe, it, expect } from 'vitest'
import {
  describeProjectTiming,
  formatProjectCountdown,
} from '../../layers/admin/app/utils/dates'
import {
  selectHomeProjects,
  MAX_ACTIVE_PROJECT_SECTIONS,
} from '../../layers/admin/app/utils/homeProjects'

// Noon everywhere, per the repo convention — midnight comparisons land on
// timezone boundaries and make day arithmetic flaky.
const NOW = new Date('2026-09-21T12:00:00')

const project = (
  id: string,
  startDate: string,
  endDate: string,
): { id: string; startDate: string; endDate: string } => ({
  id,
  startDate,
  endDate,
})

describe('describeProjectTiming', () => {
  it('reports a project that has not started as upcoming', () => {
    expect(
      describeProjectTiming('2026-09-25T10:00:00', '2026-09-30T10:00:00', NOW),
    ).toEqual({ state: 'upcoming', days: 4 })
  })

  it('reports a running project with days remaining', () => {
    expect(
      describeProjectTiming('2026-09-18T10:00:00', '2026-09-24T10:00:00', NOW),
    ).toEqual({ state: 'running', days: 3 })
  })

  it('reports an ended project with days since', () => {
    expect(
      describeProjectTiming('2026-09-01T10:00:00', '2026-09-19T10:00:00', NOW),
    ).toEqual({ state: 'ended', days: 2 })
  })

  // The reason the comparison is at day granularity rather than on timestamps.
  it('counts the final day as running, not ended', () => {
    expect(
      describeProjectTiming('2026-09-18T10:00:00', '2026-09-21T09:00:00', NOW),
    ).toEqual({ state: 'running', days: 0 })
  })

  it('counts the first day as running, not upcoming', () => {
    expect(
      describeProjectTiming('2026-09-21T18:00:00', '2026-09-25T10:00:00', NOW),
    ).toEqual({ state: 'running', days: 4 })
  })
})

describe('formatProjectCountdown', () => {
  it('labels upcoming projects', () => {
    expect(formatProjectCountdown({ state: 'upcoming', days: 0 })).toBe(
      'Starter i dag',
    )
    expect(formatProjectCountdown({ state: 'upcoming', days: 1 })).toBe(
      'Starter i morgen',
    )
    expect(formatProjectCountdown({ state: 'upcoming', days: 5 })).toBe(
      'Starter om 5 dager',
    )
  })

  it('labels running projects', () => {
    expect(formatProjectCountdown({ state: 'running', days: 0 })).toBe(
      'Siste dag',
    )
    expect(formatProjectCountdown({ state: 'running', days: 1 })).toBe(
      'Én dag igjen',
    )
    expect(formatProjectCountdown({ state: 'running', days: 9 })).toBe(
      '9 dager igjen',
    )
  })

  it('labels ended projects', () => {
    expect(formatProjectCountdown({ state: 'ended', days: 0 })).toBe(
      'Avsluttet i dag',
    )
    expect(formatProjectCountdown({ state: 'ended', days: 1 })).toBe(
      'Avsluttet i går',
    )
    expect(formatProjectCountdown({ state: 'ended', days: 3 })).toBe(
      'Avsluttet for 3 dager siden',
    )
  })
})

describe('selectHomeProjects', () => {
  it('returns nothing for an empty or missing list', () => {
    expect(selectHomeProjects(undefined, NOW)).toEqual({
      active: [],
      hiddenActiveCount: 0,
      next: null,
    })
    expect(selectHomeProjects([], NOW)).toEqual({
      active: [],
      hiddenActiveCount: 0,
      next: null,
    })
  })

  it('orders active projects by ending soonest', () => {
    const result = selectHomeProjects(
      [
        project('late', '2026-09-01T10:00:00', '2026-12-01T10:00:00'),
        project('soon', '2026-09-01T10:00:00', '2026-09-23T10:00:00'),
        project('mid', '2026-09-01T10:00:00', '2026-10-15T10:00:00'),
      ],
      NOW,
    )
    expect(result.active.map((p) => p.id)).toEqual(['soon', 'mid', 'late'])
  })

  it('breaks ties on end date by start date, so the order is stable', () => {
    const result = selectHomeProjects(
      [
        project('startedLater', '2026-09-10T10:00:00', '2026-09-30T10:00:00'),
        project('startedEarlier', '2026-09-01T10:00:00', '2026-09-30T10:00:00'),
      ],
      NOW,
    )
    expect(result.active.map((p) => p.id)).toEqual([
      'startedEarlier',
      'startedLater',
    ])
  })

  // Multiple projects may run at once — the whole reason the page is built for N.
  it('keeps several active projects rather than picking one', () => {
    const result = selectHomeProjects(
      [
        project('a', '2026-09-01T10:00:00', '2026-09-25T10:00:00'),
        project('b', '2026-09-01T10:00:00', '2026-09-26T10:00:00'),
      ],
      NOW,
    )
    expect(result.active).toHaveLength(2)
    expect(result.hiddenActiveCount).toBe(0)
  })

  it('caps the cards and reports the overflow', () => {
    const many = Array.from(
      { length: MAX_ACTIVE_PROJECT_SECTIONS + 2 },
      (_, i) =>
        project(`p${i}`, '2026-09-01T10:00:00', `2026-10-0${i + 1}T10:00:00`),
    )
    const result = selectHomeProjects(many, NOW)
    expect(result.active).toHaveLength(MAX_ACTIVE_PROJECT_SECTIONS)
    expect(result.hiddenActiveCount).toBe(2)
  })

  it('falls back to the soonest upcoming project when none are active', () => {
    const result = selectHomeProjects(
      [
        project('later', '2026-11-01T10:00:00', '2026-11-10T10:00:00'),
        project('sooner', '2026-10-01T10:00:00', '2026-10-10T10:00:00'),
      ],
      NOW,
    )
    expect(result.active).toEqual([])
    expect(result.next?.id).toBe('sooner')
  })

  // The bug this page had: upcoming projects were fetched and then discarded,
  // so an empty state showed while the next camp sat in the same response.
  it('never hides an upcoming project behind an empty state', () => {
    const result = selectHomeProjects(
      [project('next', '2026-10-01T10:00:00', '2026-10-10T10:00:00')],
      NOW,
    )
    expect(result.next).not.toBeNull()
  })

  it('does not offer a next project while one is active', () => {
    const result = selectHomeProjects(
      [
        project('running', '2026-09-01T10:00:00', '2026-09-25T10:00:00'),
        project('upcoming', '2026-10-01T10:00:00', '2026-10-10T10:00:00'),
      ],
      NOW,
    )
    expect(result.next).toBeNull()
  })

  it('ignores ended projects entirely', () => {
    const result = selectHomeProjects(
      [project('done', '2026-08-01T10:00:00', '2026-09-01T10:00:00')],
      NOW,
    )
    expect(result).toEqual({ active: [], hiddenActiveCount: 0, next: null })
  })
})
