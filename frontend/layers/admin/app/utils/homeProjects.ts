import { describeProjectTiming } from './dates'

/**
 * How many active projects get their own section before the rest collapse into
 * an "N more" link. Multiple projects can run at once, so the home page is
 * built for N — but in practice there is usually one, and an unbounded number
 * of full sections stops the page being answerable at a glance.
 */
export const MAX_ACTIVE_PROJECT_SECTIONS = 3

export interface TimedProject {
  id: string
  startDate: string
  endDate: string
}

export interface HomeProjectSelection<T extends TimedProject> {
  /** Running now, soonest to end first, capped at MAX_ACTIVE_PROJECT_SECTIONS. */
  active: T[]
  /** Active projects beyond the cap, for an "N more" link. */
  hiddenActiveCount: number
  /** The next one to start. Only set when nothing is active. */
  next: T | null
}

/**
 * Pick what the home page shows from the set of not-yet-ended projects.
 *
 * Active projects are ordered by **ending soonest**: a camp that ends tomorrow
 * needs attention before one that runs for another three months. Start date is
 * the tiebreak so the order is stable rather than dependent on server ordering.
 *
 * `next` is deliberately only populated when nothing is active — it is the
 * fallback for the empty state, not a permanent section, and showing both at
 * once buries the thing that is actually live.
 */
export function selectHomeProjects<T extends TimedProject>(
  projects: T[] | undefined,
  now: Date = new Date(),
): HomeProjectSelection<T> {
  const all = projects ?? []

  const timed = all.map((project) => ({
    project,
    timing: describeProjectTiming(project.startDate, project.endDate, now),
  }))

  const active = timed
    .filter((entry) => entry.timing.state === 'running')
    .sort(
      (a, b) =>
        Date.parse(a.project.endDate) - Date.parse(b.project.endDate) ||
        Date.parse(a.project.startDate) - Date.parse(b.project.startDate),
    )
    .map((entry) => entry.project)

  const upcoming = timed
    .filter((entry) => entry.timing.state === 'upcoming')
    .sort(
      (a, b) =>
        Date.parse(a.project.startDate) - Date.parse(b.project.startDate),
    )
    .map((entry) => entry.project)

  return {
    active: active.slice(0, MAX_ACTIVE_PROJECT_SECTIONS),
    hiddenActiveCount: Math.max(0, active.length - MAX_ACTIVE_PROJECT_SECTIONS),
    next: active.length === 0 ? (upcoming[0] ?? null) : null,
  }
}
