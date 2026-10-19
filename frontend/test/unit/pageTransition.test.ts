import { describe, it, expect } from 'vitest'
import {
  routeDepth,
  pageTransitionDirection,
  pageTransitionMeta,
} from '../../layers/user/app/utils/pageTransition'

describe('pageTransition', () => {
  describe('routeDepth', () => {
    it('puts every tab route at the root', () => {
      expect(routeDepth('/')).toBe(0)
      expect(routeDepth('/standings')).toBe(0)
      expect(routeDepth('/challenges')).toBe(0)
    })

    it('counts segments for non-tab routes', () => {
      expect(routeDepth('/settings')).toBe(1)
      expect(routeDepth('/settings/archive')).toBe(2)
      expect(routeDepth('/challenges/CL01ARZ3NDEKTSV4RRFFQ69G5FAV')).toBe(2)
    })

    it('ignores a trailing slash, query string and hash', () => {
      expect(routeDepth('/standings/')).toBe(0)
      expect(routeDepth('/settings/archive/')).toBe(2)
      expect(routeDepth('/standings?tab=unit')).toBe(0)
      expect(routeDepth('/settings#top')).toBe(1)
    })
  })

  describe('pageTransitionDirection', () => {
    it('does not travel between tabs', () => {
      expect(pageTransitionDirection('/', '/standings')).toBe(0)
      expect(pageTransitionDirection('/challenges', '/')).toBe(0)
      expect(pageTransitionDirection('/standings', '/challenges')).toBe(0)
    })

    it('pushes when moving deeper', () => {
      expect(pageTransitionDirection('/', '/settings')).toBe(1)
      expect(pageTransitionDirection('/settings', '/settings/archive')).toBe(1)
      expect(
        pageTransitionDirection(
          '/challenges',
          '/challenges/CL01ARZ3NDEKTSV4RRFF',
        ),
      ).toBe(1)
    })

    it('pops when moving back out', () => {
      expect(pageTransitionDirection('/settings', '/')).toBe(-1)
      expect(pageTransitionDirection('/settings/archive', '/settings')).toBe(-1)
      expect(
        pageTransitionDirection(
          '/challenges/CL01ARZ3NDEKTSV4RRFF',
          '/challenges',
        ),
      ).toBe(-1)
    })

    // A push and its matching pop must be exact opposites, or the two halves
    // of one navigation slide against each other.
    it.each([
      ['/', '/settings'],
      ['/settings', '/settings/archive'],
      ['/challenges', '/challenges/CL01ARZ3NDEKTSV4RRFF'],
    ])('is symmetric between %s and %s', (shallow, deep) => {
      const out = pageTransitionDirection(shallow, deep)
      const back = pageTransitionDirection(deep, shallow)
      expect(out).toBe(1)
      expect(back).toBe(-1)
      expect(out).toBe(-(back as number))
    })

    it('treats a sibling at the same depth as a cross-fade', () => {
      expect(
        pageTransitionDirection('/settings/archive', '/settings/consent'),
      ).toBe(0)
    })

    it('leaves the admin panel alone in both directions', () => {
      expect(pageTransitionDirection('/', '/admin/projects')).toBeNull()
      expect(pageTransitionDirection('/admin/projects', '/')).toBeNull()
      expect(
        pageTransitionDirection('/admin/projects', '/admin/projects/PR01'),
      ).toBeNull()
    })

    it('does not travel when only the query changes', () => {
      expect(pageTransitionDirection('/standings', '/standings')).toBe(0)
      expect(pageTransitionDirection('/standings', '/standings?tab=unit')).toBe(
        0,
      )
    })
  })
})

describe('pageTransitionMeta', () => {
  it('gives siblings no transition, the way a native tab bar swaps', () => {
    expect(pageTransitionMeta(0)).toBe(false)
  })

  it('animates a change of depth in both directions', () => {
    expect(pageTransitionMeta(1)).toEqual({ name: 'page', mode: 'out-in' })
    expect(pageTransitionMeta(-1)).toEqual({ name: 'page', mode: 'out-in' })
  })

  it('names one transition for both directions, so the halves cannot disagree', () => {
    expect(pageTransitionMeta(1)).toEqual(pageTransitionMeta(-1))
  })

  it('returns a value for every sibling pair rather than leaving meta unset', () => {
    // Route meta lives on the record and persists, so a route reached once by
    // a push must be actively reset when it is later reached from a sibling.
    const tabs = ['/', '/standings', '/challenges']

    for (const from of tabs) {
      for (const to of tabs) {
        const direction = pageTransitionDirection(from, to)
        expect(direction).not.toBeNull()
        expect(pageTransitionMeta(direction!)).toBe(false)
      }
    }
  })
})
