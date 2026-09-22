import { describe, it, expect } from 'vitest'
import {
  nextOffset,
  previousOffset,
  pageRange,
} from '../../layers/admin/app/utils/pagination'

describe('offset stepping', () => {
  it('advances by a page', () => {
    expect(nextOffset(0, 15)).toBe(15)
    expect(nextOffset(15, 15)).toBe(30)
  })

  it('steps back by a page', () => {
    expect(previousOffset(30, 15)).toBe(15)
    expect(previousOffset(15, 15)).toBe(0)
  })

  // Going back from the first page must not produce a negative position, which
  // would render as "Viser -14–0 av …".
  it('never goes below zero', () => {
    expect(previousOffset(0, 15)).toBe(0)
    expect(previousOffset(10, 15)).toBe(0)
  })
})

describe('pageRange', () => {
  it('describes the first page', () => {
    expect(pageRange(0, 15, 13477)).toEqual({ start: 1, end: 15, total: 13477 })
  })

  it('describes a later page', () => {
    expect(pageRange(30, 15, 13477)).toEqual({
      start: 31,
      end: 45,
      total: 13477,
    })
  })

  // `end` comes from the rows actually returned, not from the page size, so a
  // short final page does not overshoot the total.
  it('does not overshoot on a short final page', () => {
    expect(pageRange(13470, 7, 13477)).toEqual({
      start: 13471,
      end: 13477,
      total: 13477,
    })
  })

  it('has nothing to describe when the list is empty', () => {
    expect(pageRange(0, 0, 0)).toBeNull()
    expect(pageRange(0, 0, null)).toBeNull()
    expect(pageRange(0, 0, undefined)).toBeNull()
  })

  it('has nothing to describe when a page returned no rows', () => {
    expect(pageRange(30, 0, 100)).toBeNull()
  })

  // Data can shrink under a held position — a filter applied elsewhere, a
  // deletion. Printing "Viser 501–515 av 12" would be worse than printing
  // nothing.
  it('returns null when the offset is past the end', () => {
    expect(pageRange(500, 15, 12)).toBeNull()
  })

  it('clamps the end to the total', () => {
    expect(pageRange(0, 50, 20)).toEqual({ start: 1, end: 20, total: 20 })
  })
})
