import { describe, it, expect } from 'vitest'
import {
  sparklineBars,
  dailyAverage,
} from '../../layers/admin/app/utils/sparkline'

describe('sparklineBars', () => {
  const opts = { width: 100, height: 40 }

  it('returns one bar per value, left to right', () => {
    const bars = sparklineBars([1, 2, 3, 4], opts)

    expect(bars).toHaveLength(4)
    expect(bars.map((bar) => bar.index)).toEqual([0, 1, 2, 3])
    const xs = bars.map((bar) => bar.x)
    expect([...xs].sort((a, b) => a - b)).toEqual(xs)
  })

  it('scales the tallest bar to the full height', () => {
    const bars = sparklineBars([1, 5], opts)

    expect(bars[1]!.height).toBe(40)
    expect(bars[1]!.y).toBe(0)
  })

  it('anchors every bar to the baseline', () => {
    const bars = sparklineBars([3, 7, 1], opts)

    for (const bar of bars) {
      expect(bar.y + bar.height).toBeCloseTo(40)
    }
  })

  it('leaves a gap between adjacent bars', () => {
    const bars = sparklineBars([1, 1, 1, 1], { ...opts, gap: 2 })

    const slot = 100 / 4
    expect(bars[0]!.width).toBe(slot - 2)
    // The right edge of one bar must not reach the left edge of the next.
    expect(bars[0]!.x + bars[0]!.width).toBeLessThan(bars[1]!.x)
  })

  // An all-zero window is the normal state for a quiet project. Dividing by the
  // max would be a divide-by-zero and produce NaN geometry, which renders as
  // nothing at all with no console error.
  it('renders a flat series for all zeros rather than NaN', () => {
    const bars = sparklineBars([0, 0, 0], opts)

    expect(bars).toHaveLength(3)
    for (const bar of bars) {
      expect(bar.height).toBe(0)
      expect(Number.isNaN(bar.y)).toBe(false)
      expect(bar.y).toBe(40)
    }
  })

  // A day with a small value next to a large one must not round away to an
  // invisible sliver — otherwise "some activity" looks identical to "none".
  it('gives a non-zero value a minimum visible height', () => {
    const bars = sparklineBars([1, 100000], { ...opts, minBarHeight: 2 })

    expect(bars[0]!.height).toBeGreaterThanOrEqual(2)
  })

  it('keeps a zero value at zero height', () => {
    const bars = sparklineBars([0, 50], { ...opts, minBarHeight: 2 })

    expect(bars[0]!.height).toBe(0)
  })

  it('handles a single-day window', () => {
    const bars = sparklineBars([5], opts)

    expect(bars).toHaveLength(1)
    expect(bars[0]!.height).toBe(40)
    expect(bars[0]!.width).toBeGreaterThan(0)
  })

  it('returns nothing for an empty series or a zero-size box', () => {
    expect(sparklineBars([], opts)).toEqual([])
    expect(sparklineBars([1, 2], { width: 0, height: 40 })).toEqual([])
    expect(sparklineBars([1, 2], { width: 100, height: 0 })).toEqual([])
  })

  it('never produces a negative bar from bad data', () => {
    const bars = sparklineBars([-5, 10], opts)

    for (const bar of bars) {
      expect(bar.height).toBeGreaterThanOrEqual(0)
      expect(bar.width).toBeGreaterThan(0)
    }
  })
})

describe('dailyAverage', () => {
  // This exists because a distinct-users-per-day count is not additive.
  it('averages rather than sums', () => {
    expect(dailyAverage([10, 10, 10])).toBe(10)
    expect(dailyAverage([0, 0, 30])).toBe(10)
  })

  it('rounds to a whole number', () => {
    expect(dailyAverage([1, 2])).toBe(2)
    expect(dailyAverage([1, 1, 2])).toBe(1)
  })

  it('is 0 for an empty series rather than NaN', () => {
    expect(dailyAverage([])).toBe(0)
  })
})
