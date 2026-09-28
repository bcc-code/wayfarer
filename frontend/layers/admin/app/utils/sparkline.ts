// Geometry for a bar sparkline. Bars, not a line: one discrete bucket per
// day, and a bar is its own hit target.

export interface SparklineBar {
  x: number
  y: number
  width: number
  height: number
  /** Index in the input series, so callers can style the last bar. */
  index: number
  /**
   * Full-height hit column for this day, spanning the whole slot including the
   * gap. The painted bar is a poor hit target — a quiet day is a 2px sliver,
   * which is unhoverable in practice — so pointer handling attaches here
   * instead, and the reader only has to be over the right column.
   */
  hitX: number
  hitWidth: number
  /** Slot centre as a fraction of total width, for positioning a tooltip. */
  centerRatio: number
}

export interface SparklineOptions {
  width: number
  height: number
  gap?: number
  /** So a non-zero day is never invisible next to a large one. */
  minBarHeight?: number
}

export function sparklineBars(
  values: number[],
  { width, height, gap = 2, minBarHeight = 2 }: SparklineOptions,
): SparklineBar[] {
  if (values.length === 0 || width <= 0 || height <= 0) return []

  const slot = width / values.length
  const barWidth = Math.max(1, slot - gap)

  // Scaled to the largest value: a fixed ceiling would flatten quiet projects.
  const max = Math.max(...values)

  return values.map((value, index) => {
    // Guarded: an all-zero series would divide by zero into NaN geometry.
    const ratio = max > 0 ? Math.max(0, value) / max : 0
    const barHeight = ratio === 0 ? 0 : Math.max(minBarHeight, ratio * height)

    const hitX = index * slot

    return {
      x: hitX + gap / 2,
      y: height - barHeight,
      width: barWidth,
      height: barHeight,
      index,
      hitX,
      hitWidth: slot,
      centerRatio: (hitX + slot / 2) / width,
    }
  })
}

// For series that are not additive: summing distinct users per day would count
// the same person once per day they appeared.
export function dailyAverage(values: number[]): number {
  if (values.length === 0) return 0
  const total = values.reduce((sum, value) => sum + value, 0)
  return Math.round(total / values.length)
}
