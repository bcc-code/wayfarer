/**
 * Direction-aware page transitions for the user-facing app.
 *
 * Direction is derived from path depth rather than from history state, which
 * keeps it a pure function of the two paths — unit testable, and identical
 * whether the navigation came from a tap, the iOS edge-swipe gesture or the
 * Android back button.
 *
 * The direction is published as a CSS variable on a stable element rather than
 * stored per route in `meta.pageTransition`, and that is load-bearing. Under
 * `mode: 'out-in'` NuxtPage resolves transition props from
 * `routeProps.route` — the route RouterView is *currently rendering* — so the
 * leaving page is governed by the route being left, not the one being entered.
 * A per-route transition name therefore animates the two halves of one
 * navigation in opposite directions. One shared variable cannot disagree with
 * itself.
 */

/** Sign of the horizontal travel: deeper, level, or back out. */
export type PageTransitionDirection = 1 | 0 | -1

/**
 * The bottom navigation routes. They are siblings even though their segment
 * counts differ, so they all share a depth and cross-fade between each other
 * instead of pushing.
 */
const TAB_PATHS = new Set(['/', '/standings', '/challenges'])

function normalise(path: string): string {
  const [withoutQuery = ''] = path.split(/[?#]/)
  if (withoutQuery.length > 1 && withoutQuery.endsWith('/')) {
    return withoutQuery.slice(0, -1)
  }
  return withoutQuery || '/'
}

/**
 * How deep a path sits in the navigation hierarchy. Tab routes are the root;
 * everything else counts its segments, so `/challenges/:id` and
 * `/settings/archive` both sit two levels in.
 */
export function routeDepth(path: string): number {
  const normalised = normalise(path)
  if (TAB_PATHS.has(normalised)) return 0
  return normalised.split('/').filter(Boolean).length
}

/**
 * Which way the pages should travel when moving between two paths: `1` pushes
 * deeper, `-1` pops back out, `0` cross-fades without travel.
 *
 * `null` means the navigation is none of this layer's business — the admin
 * panel navigates like a desktop app and should not animate at all.
 */
export function pageTransitionDirection(
  fromPath: string,
  toPath: string,
): PageTransitionDirection | null {
  const from = normalise(fromPath)
  const to = normalise(toPath)

  if (from.startsWith('/admin') || to.startsWith('/admin')) return null

  const fromDepth = routeDepth(from)
  const toDepth = routeDepth(to)

  if (toDepth > fromDepth) return 1
  if (toDepth < fromDepth) return -1
  return 0
}
