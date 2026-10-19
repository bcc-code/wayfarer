import { describe, it, expect } from 'vitest'
import {
  visibilityFromVisibleAt,
  visibleAtForVisibility,
} from '../../layers/admin/app/utils/challengeVisibility'

const now = new Date('2026-06-15T12:00')

describe('visibilityFromVisibleAt', () => {
  // The case the form exists to make visible: no timestamp is a real setting,
  // not an unset field.
  it('reads an empty value as enrolled-only', () => {
    expect(visibilityFromVisibleAt(undefined, now)).toBe('enrolled')
    expect(visibilityFromVisibleAt('', now)).toBe('enrolled')
  })

  it('reads a past timestamp as visible to everyone', () => {
    expect(visibilityFromVisibleAt('2026-06-15T11:59', now)).toBe('everyone')
    expect(visibilityFromVisibleAt('2020-01-01T00:00', now)).toBe('everyone')
  })

  it('reads a future timestamp as scheduled', () => {
    expect(visibilityFromVisibleAt('2026-06-15T12:01', now)).toBe('scheduled')
  })

  it('falls back to enrolled-only for an unparseable value', () => {
    expect(visibilityFromVisibleAt('i går', now)).toBe('enrolled')
  })
})

describe('visibleAtForVisibility', () => {
  it('stores nothing for enrolled-only', () => {
    expect(
      visibleAtForVisibility('enrolled', '2026-07-01T09:00', undefined, now),
    ).toBeUndefined()
  })

  it('stores the picked time when scheduled', () => {
    expect(
      visibleAtForVisibility('scheduled', '2026-07-01T09:00', undefined, now),
    ).toBe('2026-07-01T09:00')
  })

  it('stores now when a hidden challenge becomes visible to everyone', () => {
    expect(visibleAtForVisibility('everyone', undefined, undefined, now)).toBe(
      '2026-06-15T12:00',
    )
  })

  // Saving an unrelated edit must not move the date a challenge went live.
  it('keeps the original timestamp of an already visible challenge', () => {
    expect(
      visibleAtForVisibility('everyone', undefined, '2026-01-02T08:30', now),
    ).toBe('2026-01-02T08:30')
  })

  // A scheduled time that has passed means the challenge is live now; the
  // stored timestamp is still the one that let it through.
  it('keeps a past scheduled timestamp when switching to everyone', () => {
    expect(
      visibleAtForVisibility(
        'everyone',
        '2026-07-01T09:00',
        '2026-06-15T11:00',
        now,
      ),
    ).toBe('2026-06-15T11:00')
  })

  // A future timestamp is not "visible to everyone" yet, so it is replaced.
  it('replaces a future timestamp when switching to everyone', () => {
    expect(
      visibleAtForVisibility('everyone', undefined, '2026-08-01T09:00', now),
    ).toBe('2026-06-15T12:00')
  })
})
