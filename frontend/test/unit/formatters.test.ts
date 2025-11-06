import { describe, it, expect } from 'vitest'
import {
  getInitials,
  formatDate,
  capitalizeFirst,
} from '../../app/utils/formatters'

describe('formatters', () => {
  describe('getInitials', () => {
    it('should return initials from a full name', () => {
      expect(getInitials('John Doe')).toBe('JD')
    })

    it('should handle names with multiple parts', () => {
      expect(getInitials('John Michael Doe')).toBe('JMD')
    })

    it('should handle single name', () => {
      expect(getInitials('John')).toBe('J')
    })

    it('should handle extra spaces', () => {
      expect(getInitials('John  Doe')).toBe('JD')
      expect(getInitials(' John Doe ')).toBe('JD')
    })

    it('should handle empty string', () => {
      expect(getInitials('')).toBe('')
    })

    it('should handle names with hyphens', () => {
      expect(getInitials('Mary-Jane Watson')).toBe('MW')
    })

    it('should handle lowercase names', () => {
      expect(getInitials('john doe')).toBe('jd')
    })

    it('should handle mixed case names', () => {
      expect(getInitials('jOhN DoE')).toBe('jD')
    })
  })

  describe('formatDate', () => {
    it('should format ISO date string', () => {
      const result = formatDate('2000-01-15')
      expect(result).toBe('January 15, 2000')
    })

    it('should format date at start of year', () => {
      const result = formatDate('2024-01-01')
      expect(result).toBe('January 1, 2024')
    })

    it('should format date at end of year', () => {
      const result = formatDate('2024-12-31')
      expect(result).toBe('December 31, 2024')
    })

    it('should handle dates with time component', () => {
      const result = formatDate('2000-01-15T10:30:00Z')
      expect(result).toBe('January 15, 2000')
    })

    it('should format leap year date', () => {
      const result = formatDate('2024-02-29')
      expect(result).toBe('February 29, 2024')
    })

    it('should format dates in the past', () => {
      const result = formatDate('1990-06-15')
      expect(result).toBe('June 15, 1990')
    })

    it('should format dates in the future', () => {
      const result = formatDate('2050-03-20')
      expect(result).toBe('March 20, 2050')
    })
  })

  describe('capitalizeFirst', () => {
    it('should capitalize first letter and lowercase rest', () => {
      expect(capitalizeFirst('hello')).toBe('Hello')
    })

    it('should handle all uppercase', () => {
      expect(capitalizeFirst('HELLO')).toBe('Hello')
    })

    it('should handle all lowercase', () => {
      expect(capitalizeFirst('hello')).toBe('Hello')
    })

    it('should handle mixed case', () => {
      expect(capitalizeFirst('hELLo')).toBe('Hello')
    })

    it('should handle single character', () => {
      expect(capitalizeFirst('h')).toBe('H')
      expect(capitalizeFirst('H')).toBe('H')
    })

    it('should handle empty string', () => {
      expect(capitalizeFirst('')).toBe('')
    })

    it('should handle string with numbers', () => {
      expect(capitalizeFirst('hello123')).toBe('Hello123')
    })

    it('should handle string with special characters', () => {
      expect(capitalizeFirst('hello-world')).toBe('Hello-world')
    })

    it('should handle gender values', () => {
      expect(capitalizeFirst('MALE')).toBe('Male')
      expect(capitalizeFirst('FEMALE')).toBe('Female')
    })
  })
})
