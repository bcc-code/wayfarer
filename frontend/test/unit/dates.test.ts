import { describe, it, expect, beforeEach, vi, afterEach } from 'vitest'
import { formatDate, formatDateRange } from '../../app/utils/formatters'
import {
  calculateDuration,
  getProjectStatus,
  isWithinRange,
  extractDateOnly,
  formatDateWithTimezone,
} from '../../app/utils/dates'

describe('dates utilities', () => {
  describe('formatDate', () => {
    it('should format ISO date string to readable format', () => {
      expect(formatDate('2024-01-15')).toBe('January 15, 2024')
    })

    it('should handle different months', () => {
      expect(formatDate('2024-06-30')).toBe('June 30, 2024')
      expect(formatDate('2024-12-25')).toBe('December 25, 2024')
    })

    it('should format single digit days', () => {
      expect(formatDate('2024-03-05')).toBe('March 5, 2024')
    })
  })

  describe('formatDateRange', () => {
    it('should format same month and year dates', () => {
      const result = formatDateRange('2024-03-15', '2024-03-20')
      expect(result).toBe('Mar 15\u2009\u2013\u200920, 2024')
    })

    it('should format different months same year', () => {
      const result = formatDateRange('2024-03-15', '2024-05-20')
      expect(result).toBe('Mar 15\u2009\u2013\u2009May 20, 2024')
    })

    it('should format different years', () => {
      const result = formatDateRange('2024-12-15', '2025-01-20')
      expect(result).toBe('Dec 15, 2024\u2009\u2013\u2009Jan 20, 2025')
    })

    it('should handle year boundaries', () => {
      const result = formatDateRange('2023-11-28', '2024-02-14')
      expect(result).toBe('Nov 28, 2023\u2009\u2013\u2009Feb 14, 2024')
    })

    it('should handle single day ranges (same start and end)', () => {
      const result = formatDateRange('2024-05-15', '2024-05-15')
      expect(result).toBe('May 15, 2024')
    })
  })

  describe('calculateDuration', () => {
    it('should calculate single day duration', () => {
      expect(calculateDuration('2024-01-01', '2024-01-02')).toBe('1 day')
    })

    it('should calculate multiple days', () => {
      expect(calculateDuration('2024-01-01', '2024-01-05')).toBe('4 days')
    })

    it('should calculate weeks', () => {
      expect(calculateDuration('2024-01-01', '2024-01-08')).toBe('1 week')
      expect(calculateDuration('2024-01-01', '2024-01-15')).toBe('2 weeks')
    })

    it('should calculate months', () => {
      expect(calculateDuration('2024-01-01', '2024-02-01')).toBe('1 month')
      expect(calculateDuration('2024-01-01', '2024-03-05')).toBe('2 months')
    })

    it('should prefer weeks over days for 7+ days under a month', () => {
      expect(calculateDuration('2024-01-01', '2024-01-22')).toBe('3 weeks')
    })

    it('should handle zero duration', () => {
      // Same date gives 0 days
      expect(calculateDuration('2024-01-01', '2024-01-01')).toBe('0 days')
    })
  })

  describe('getProjectStatus', () => {
    beforeEach(() => {
      // Mock the current date to 2024-06-15
      vi.useFakeTimers()
      vi.setSystemTime(new Date('2024-06-15'))
    })

    afterEach(() => {
      vi.useRealTimers()
    })

    it('should return "Upcoming" for future projects', () => {
      const status = getProjectStatus('2024-07-01', '2024-08-01')
      expect(status).toBe('Upcoming')
    })

    it('should return "In Progress" for current projects', () => {
      const status = getProjectStatus('2024-06-01', '2024-06-30')
      expect(status).toBe('In Progress')
    })

    it('should return "Completed" for past projects', () => {
      const status = getProjectStatus('2024-05-01', '2024-05-31')
      expect(status).toBe('Completed')
    })

    it('should return "In Progress" when today is start date', () => {
      const status = getProjectStatus('2024-06-15', '2024-07-01')
      expect(status).toBe('In Progress')
    })

    it('should return "In Progress" when today is end date', () => {
      // End date is inclusive (project is still in progress on end date)
      const status = getProjectStatus('2024-05-01', '2024-06-15')
      expect(status).toBe('In Progress')
    })
  })

  describe('isWithinRange', () => {
    it('should return true for date within range', () => {
      const result = isWithinRange('2024-06-15', '2024-06-01', '2024-06-30')
      expect(result).toBe(true)
    })

    it('should return true for date at start of range', () => {
      const result = isWithinRange('2024-06-01', '2024-06-01', '2024-06-30')
      expect(result).toBe(true)
    })

    it('should return true for date at end of range', () => {
      const result = isWithinRange('2024-06-30', '2024-06-01', '2024-06-30')
      expect(result).toBe(true)
    })

    it('should return false for date before range', () => {
      const result = isWithinRange('2024-05-31', '2024-06-01', '2024-06-30')
      expect(result).toBe(false)
    })

    it('should return false for date after range', () => {
      const result = isWithinRange('2024-07-01', '2024-06-01', '2024-06-30')
      expect(result).toBe(false)
    })

    it('should handle Date objects', () => {
      const date = new Date('2024-06-15')
      const start = new Date('2024-06-01')
      const end = new Date('2024-06-30')
      expect(isWithinRange(date, start, end)).toBe(true)
    })

    it('should handle timestamps', () => {
      const date = new Date('2024-06-15').getTime()
      const start = new Date('2024-06-01').getTime()
      const end = new Date('2024-06-30').getTime()
      expect(isWithinRange(date, start, end)).toBe(true)
    })

    it('should handle mixed types', () => {
      const result = isWithinRange(
        new Date('2024-06-15'),
        '2024-06-01',
        new Date('2024-06-30').getTime(),
      )
      expect(result).toBe(true)
    })
  })

  describe('extractDateOnly', () => {
    it('should extract date from ISO timestamp', () => {
      expect(extractDateOnly('2024-01-15T10:30:00Z')).toBe('2024-01-15')
    })

    it('should return date-only string as is', () => {
      expect(extractDateOnly('2024-01-15')).toBe('2024-01-15')
    })

    it('should handle timestamp with timezone offset', () => {
      expect(extractDateOnly('2024-01-15T10:30:00+02:00')).toBe('2024-01-15')
    })

    it('should return undefined for undefined input', () => {
      expect(extractDateOnly(undefined)).toBeUndefined()
    })

    it('should return undefined for empty string', () => {
      expect(extractDateOnly('')).toBeUndefined()
    })

    it('should return undefined for whitespace-only string', () => {
      expect(extractDateOnly('   ')).toBeUndefined()
    })
  })

  describe('formatDateWithTimezone', () => {
    it('should format date with timezone offset', () => {
      const result = formatDateWithTimezone(2024, 1, 15)

      // Should be in format: YYYY-MM-DDTHH:MM:SS+HH:MM
      expect(result).toMatch(/^2024-01-15T01:00:00[+-]\d{2}:\d{2}$/)
    })

    it('should set time to 01:00:00', () => {
      const result = formatDateWithTimezone(2024, 6, 20)

      expect(result).toContain('T01:00:00')
    })

    it('should handle single digit months correctly', () => {
      const result = formatDateWithTimezone(2024, 3, 15)

      expect(result).toContain('2024-03-15')
    })

    it('should handle single digit days correctly', () => {
      const result = formatDateWithTimezone(2024, 10, 5)

      expect(result).toContain('2024-10-05')
    })

    it('should handle end of year', () => {
      const result = formatDateWithTimezone(2024, 12, 31)

      expect(result).toContain('2024-12-31')
    })

    it('should handle start of year', () => {
      const result = formatDateWithTimezone(2024, 1, 1)

      expect(result).toContain('2024-01-01')
    })

    it('should produce a parseable ISO string', () => {
      const result = formatDateWithTimezone(2024, 6, 15)
      const parsed = new Date(result)

      expect(parsed.getFullYear()).toBe(2024)
      expect(parsed.getMonth()).toBe(5) // 0-indexed
      expect(parsed.getDate()).toBe(15)
    })
  })
})
