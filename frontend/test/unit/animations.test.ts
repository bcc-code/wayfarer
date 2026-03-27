import { describe, it, expect } from 'vitest'
import { calculateStaggerTiming } from '../../app/utils/animations'

describe('animations', () => {
  describe('calculateStaggerTiming', () => {
    it('should calculate correct timing for multiple elements', () => {
      const result = calculateStaggerTiming(1.0, 5)

      expect(result.duration).toBe(0.5) // 50% of total
      expect(result.stagger).toBeCloseTo(0.125) // (1.0 - 0.5) / (5 - 1)
    })

    it('should return zero stagger for single element', () => {
      const result = calculateStaggerTiming(1.0, 1)

      expect(result.duration).toBe(0.5)
      expect(result.stagger).toBe(0)
    })

    it('should calculate correct timing for two elements', () => {
      const result = calculateStaggerTiming(0.8, 2)

      expect(result.duration).toBe(0.4) // 50% of 0.8
      expect(result.stagger).toBeCloseTo(0.4) // (0.8 - 0.4) / (2 - 1)
    })

    it('should handle different total durations', () => {
      const result = calculateStaggerTiming(2.0, 4)

      expect(result.duration).toBe(1.0) // 50% of 2.0
      expect(result.stagger).toBeCloseTo(0.333, 2) // (2.0 - 1.0) / (4 - 1)
    })

    it('should never return negative stagger', () => {
      const result = calculateStaggerTiming(0.1, 100)

      expect(result.stagger).toBeGreaterThanOrEqual(0)
    })
  })

})
