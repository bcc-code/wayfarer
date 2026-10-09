/** The Badging API surface, absent on browsers and on uninstalled PWAs. */
export interface BadgeCapableNavigator {
  setAppBadge?: (count?: number) => Promise<unknown>
  clearAppBadge?: () => Promise<unknown>
}

/**
 * Mirrors a count onto the home-screen icon. Rejections are swallowed: the
 * platform refuses while the app is merely open in a browser tab, which is not
 * a failure worth surfacing.
 */
export function syncAppBadge(
  navigator: BadgeCapableNavigator,
  count: number | null | undefined,
): void {
  const badge = count ?? 0

  if (badge > 0) {
    navigator.setAppBadge?.(badge)?.catch(() => {})
    return
  }

  navigator.clearAppBadge?.()?.catch(() => {})
}
