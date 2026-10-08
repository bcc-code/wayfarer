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
 * `1` pushes deeper, `-1` pops back out, `0` is a swap between siblings, and
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

export type PageTransitionMeta = false | { name: 'page'; mode: 'out-in' }

/**
 * Siblings swap instantly, the way a native tab bar does — `UITabBarController`
 * animates nothing between tabs, so a cross-fade there is not a weak version of
 * the platform behaviour, it is behaviour the platform does not have. Only a
 * change of depth is worth animating, which is also where iOS animates.
 *
 * `false` is Nuxt's "no transition", and is returned rather than left unset
 * because route meta lives on the record and persists: a route reached once by
 * a push would keep that transition when reached from a sibling later.
 */
export function pageTransitionMeta(
  direction: Exclude<PageTransitionDirection, null>,
): PageTransitionMeta {
  if (direction === 0) return false
  return { name: 'page', mode: 'out-in' }
}
