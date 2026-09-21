/**
 * Geometry for a bar sparkline.
 *
 * Bars rather than a line because the data is one discrete bucket per day, and
 * because a bar is its own hit target — a 2px polyline needs a crosshair layer
 * to be hoverable, which is more machinery than a tile-sized chart warrants.
 *
 * Pure, so the layout is unit-testable without rendering: the awkward cases
 * here are all-zero series and single-day windows, and both are easier to pin
 * as numbers than as pixels.
 */

export interface SparklineBar {
  x: number
  y: number
  width: number
  height: number
  /** Index in the input series, so callers can style the last bar. */
  index: number
}

export interface SparklineOptions {
  width: number
  height: number
  /** Surface gap between bars. The spec asks for 2px between adjacent fills. */
  gap?: number
  /** Bars shorter than this still render, so a non-zero day is never invisible. */
  minBarHeight?: number
}

export function sparklineBars(
  values: number[],
  { width, height, gap = 2, minBarHeight = 2 }: SparklineOptions,
): SparklineBar[] {
  if (values.length === 0 || width <= 0 || height <= 0) return []

  const slot = width / values.length
  const barWidth = Math.max(1, slot - gap)

  // Scale to the largest value, not to a fixed ceiling: a sparkline shows
  // shape, and a fixed max would flatten every quiet project to nothing.
  const max = Math.max(...values)

  return values.map((value, index) => {
    // An all-zero series is flat by definition — guard the divide rather than
    // letting it produce NaN geometry.
    const ratio = max > 0 ? Math.max(0, value) / max : 0
    const barHeight = ratio === 0 ? 0 : Math.max(minBarHeight, ratio * height)

    return {
      x: index * slot + gap / 2,
      y: height - barHeight,
      width: barWidth,
      height: barHeight,
      index,
    }
  })
}

/**
 * Mean of a daily series, rounded.
 *
 * Exists because a distinct-user-per-day count is **not additive** — summing
 * "active users" across days counts the same person once per day they showed
 * up. Points are additive and should just be summed; this is for the ones that
 * are not.
 */
export function dailyAverage(values: number[]): number {
  if (values.length === 0) return 0
  const total = values.reduce((sum, value) => sum + value, 0)
  return Math.round(total / values.length)
}
