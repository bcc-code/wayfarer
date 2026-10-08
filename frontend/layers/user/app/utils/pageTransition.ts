export type PageTransitionDirection = 1 | 0 | -1

/** Siblings despite their differing segment counts, so they share a depth. */
const TAB_PATHS = new Set(['/', '/standings', '/challenges'])

function normalise(path: string): string {
  const [withoutQuery = ''] = path.split(/[?#]/)
  if (withoutQuery.length > 1 && withoutQuery.endsWith('/')) {
    return withoutQuery.slice(0, -1)
  }
  return withoutQuery || '/'
}

/** Tab routes are the root; everything else counts its segments. */
export function routeDepth(path: string): number {
  const normalised = normalise(path)
  if (TAB_PATHS.has(normalised)) return 0
  return normalised.split('/').filter(Boolean).length
}

/**
 * `1` pushes deeper, `-1` pops back out, `0` cross-fades without travel, and
 * `null` leaves the navigation alone — the admin panel should not animate.
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
