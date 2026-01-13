import { describe, it, expect } from 'vitest'
import {
  buildNextPageVariables,
  buildPreviousPageVariables,
  buildFirstPageVariables,
  isFirstPage,
  isLastPage,
  validatePageSize,
} from '../../app/utils/pagination'
import type { CursorPageInfo } from '../../app/utils/pagination'

describe('pagination utils', () => {
  describe('buildNextPageVariables', () => {
    it('should return variables for next page when hasNextPage and endCursor exist', () => {
      const pageInfo: CursorPageInfo = {
        hasNextPage: true,
        hasPreviousPage: false,
        startCursor: 'start',
        endCursor: 'end',
      }

      const result = buildNextPageVariables(pageInfo, 20)

      expect(result).toEqual({
        first: 20,
        after: 'end',
        last: null,
        before: null,
      })
    })

    it('should return null when hasNextPage is false', () => {
      const pageInfo: CursorPageInfo = {
        hasNextPage: false,
        hasPreviousPage: false,
        startCursor: 'start',
        endCursor: 'end',
      }

      expect(buildNextPageVariables(pageInfo, 20)).toBeNull()
    })

    it('should return null when endCursor is missing', () => {
      const pageInfo: CursorPageInfo = {
        hasNextPage: true,
        hasPreviousPage: false,
        startCursor: 'start',
      }

      expect(buildNextPageVariables(pageInfo, 20)).toBeNull()
    })

    it('should return null when endCursor is null', () => {
      const pageInfo: CursorPageInfo = {
        hasNextPage: true,
        hasPreviousPage: false,
        startCursor: 'start',
        endCursor: null,
      }

      expect(buildNextPageVariables(pageInfo, 20)).toBeNull()
    })

    it('should use the provided page size', () => {
      const pageInfo: CursorPageInfo = {
        hasNextPage: true,
        hasPreviousPage: false,
        endCursor: 'end',
      }

      const result = buildNextPageVariables(pageInfo, 50)

      expect(result?.first).toBe(50)
    })
  })

  describe('buildPreviousPageVariables', () => {
    it('should return variables for previous page when hasPreviousPage and startCursor exist', () => {
      const pageInfo: CursorPageInfo = {
        hasNextPage: false,
        hasPreviousPage: true,
        startCursor: 'start',
        endCursor: 'end',
      }

      const result = buildPreviousPageVariables(pageInfo, 20)

      expect(result).toEqual({
        first: null,
        after: null,
        last: 20,
        before: 'start',
      })
    })

    it('should return null when hasPreviousPage is false', () => {
      const pageInfo: CursorPageInfo = {
        hasNextPage: true,
        hasPreviousPage: false,
        startCursor: 'start',
        endCursor: 'end',
      }

      expect(buildPreviousPageVariables(pageInfo, 20)).toBeNull()
    })

    it('should return null when startCursor is missing', () => {
      const pageInfo: CursorPageInfo = {
        hasNextPage: false,
        hasPreviousPage: true,
        endCursor: 'end',
      }

      expect(buildPreviousPageVariables(pageInfo, 20)).toBeNull()
    })

    it('should return null when startCursor is null', () => {
      const pageInfo: CursorPageInfo = {
        hasNextPage: false,
        hasPreviousPage: true,
        startCursor: null,
        endCursor: 'end',
      }

      expect(buildPreviousPageVariables(pageInfo, 20)).toBeNull()
    })

    it('should use the provided page size', () => {
      const pageInfo: CursorPageInfo = {
        hasNextPage: false,
        hasPreviousPage: true,
        startCursor: 'start',
      }

      const result = buildPreviousPageVariables(pageInfo, 100)

      expect(result?.last).toBe(100)
    })
  })

  describe('buildFirstPageVariables', () => {
    it('should return first page variables with page size', () => {
      const result = buildFirstPageVariables(20)

      expect(result).toEqual({
        first: 20,
        after: null,
        last: null,
        before: null,
      })
    })

    it('should include initial cursor when provided', () => {
      const result = buildFirstPageVariables(20, 'initial-cursor')

      expect(result).toEqual({
        first: 20,
        after: 'initial-cursor',
        last: null,
        before: null,
      })
    })

    it('should handle null initial cursor', () => {
      const result = buildFirstPageVariables(20, null)

      expect(result.after).toBeNull()
    })

    it('should handle undefined initial cursor', () => {
      const result = buildFirstPageVariables(20, undefined)

      expect(result.after).toBeNull()
    })
  })

  describe('isFirstPage', () => {
    it('should return true when pageInfo is null', () => {
      expect(isFirstPage(null)).toBe(true)
    })

    it('should return true when hasPreviousPage is false', () => {
      const pageInfo: CursorPageInfo = {
        hasNextPage: true,
        hasPreviousPage: false,
      }

      expect(isFirstPage(pageInfo)).toBe(true)
    })

    it('should return false when hasPreviousPage is true', () => {
      const pageInfo: CursorPageInfo = {
        hasNextPage: true,
        hasPreviousPage: true,
      }

      expect(isFirstPage(pageInfo)).toBe(false)
    })
  })

  describe('isLastPage', () => {
    it('should return true when pageInfo is null', () => {
      expect(isLastPage(null)).toBe(true)
    })

    it('should return true when hasNextPage is false', () => {
      const pageInfo: CursorPageInfo = {
        hasNextPage: false,
        hasPreviousPage: true,
      }

      expect(isLastPage(pageInfo)).toBe(true)
    })

    it('should return false when hasNextPage is true', () => {
      const pageInfo: CursorPageInfo = {
        hasNextPage: true,
        hasPreviousPage: false,
      }

      expect(isLastPage(pageInfo)).toBe(false)
    })
  })

  describe('validatePageSize', () => {
    it('should return true for positive numbers', () => {
      expect(validatePageSize(1)).toBe(true)
      expect(validatePageSize(20)).toBe(true)
      expect(validatePageSize(100)).toBe(true)
    })

    it('should return false for zero', () => {
      expect(validatePageSize(0)).toBe(false)
    })

    it('should return false for negative numbers', () => {
      expect(validatePageSize(-1)).toBe(false)
      expect(validatePageSize(-100)).toBe(false)
    })
  })
})
