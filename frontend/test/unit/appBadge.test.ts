import { describe, it, expect, vi } from 'vitest'
import {
  syncAppBadge,
  type BadgeCapableNavigator,
} from '../../layers/user/app/utils/appBadge'

function fakeNavigator(
  overrides: Partial<BadgeCapableNavigator> = {},
): BadgeCapableNavigator {
  return {
    setAppBadge: vi.fn().mockResolvedValue(undefined),
    clearAppBadge: vi.fn().mockResolvedValue(undefined),
    ...overrides,
  }
}

describe('syncAppBadge', () => {
  it('shows the count when there is something to show', () => {
    const nav = fakeNavigator()

    syncAppBadge(nav, 3)

    expect(nav.setAppBadge).toHaveBeenCalledWith(3)
    expect(nav.clearAppBadge).not.toHaveBeenCalled()
  })

  it.each([[0], [null], [undefined]])('clears the badge for %s', (count) => {
    const nav = fakeNavigator()

    syncAppBadge(nav, count)

    expect(nav.clearAppBadge).toHaveBeenCalled()
    expect(nav.setAppBadge).not.toHaveBeenCalled()
  })

  it('does nothing where the Badging API is absent', () => {
    expect(() => syncAppBadge({}, 3)).not.toThrow()
    expect(() => syncAppBadge({}, 0)).not.toThrow()
  })

  // The platform rejects while the app is only open in a browser tab.
  it('swallows a rejection rather than leaving it unhandled', async () => {
    const nav = fakeNavigator({
      setAppBadge: vi.fn().mockRejectedValue(new Error('not installed')),
    })

    expect(() => syncAppBadge(nav, 3)).not.toThrow()
    await Promise.resolve()
  })
})
