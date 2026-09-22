import { describeProjectTiming } from './dates'

/** Beyond this, the rest collapse into an "N more" link. */
export const MAX_ACTIVE_PROJECT_SECTIONS = 3

export interface TimedProject {
  id: string
  startDate: string
  endDate: string
}

export interface HomeProjectSelection<T extends TimedProject> {
  active: T[]
  hiddenActiveCount: number
  /** Only set when nothing is active. */
  next: T | null
}

// Active projects end-soonest first, with start date as a stable tiebreak.
// `next` is the empty-state fallback, so it is skipped when anything is live.
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
