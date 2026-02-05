import { describe, it, expect } from 'vitest'
import {
  calculateStaggerTiming,
  calculateParticleTrajectory,
  generateConfettiParticle,
  CONFETTI_COLORS,
} from '../../app/utils/animations'

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

  describe('calculateParticleTrajectory', () => {
    it('should calculate x and y based on angle and velocity', () => {
      // Test with angle = 0 (right direction)
      const result = calculateParticleTrajectory(0, 100, 0)

      expect(result.x).toBeCloseTo(100) // cos(0) * 100
      expect(result.y).toBeCloseTo(0) // sin(0) * 100 - 0
    })

    it('should apply upward bias to y', () => {
      const result = calculateParticleTrajectory(0, 100, 50)

      expect(result.y).toBeCloseTo(-50) // sin(0) * 100 - 50
    })

    it('should calculate trajectory at 90 degrees (up)', () => {
      const result = calculateParticleTrajectory(Math.PI / 2, 100, 0)

      expect(result.x).toBeCloseTo(0) // cos(90°) * 100
      expect(result.y).toBeCloseTo(100) // sin(90°) * 100
    })

    it('should return rotation between -360 and 360', () => {
      // Run multiple times to test randomness
      for (let i = 0; i < 10; i++) {
        const result = calculateParticleTrajectory(0, 100)

        expect(result.rotation).toBeGreaterThanOrEqual(-360)
        expect(result.rotation).toBeLessThanOrEqual(360)
      }
    })

    it('should return duration between 1 and 1.5', () => {
      // Run multiple times to test randomness
      for (let i = 0; i < 10; i++) {
        const result = calculateParticleTrajectory(0, 100)

        expect(result.duration).toBeGreaterThanOrEqual(1)
        expect(result.duration).toBeLessThanOrEqual(1.5)
      }
    })

    it('should use default upward bias of 50', () => {
      const result = calculateParticleTrajectory(Math.PI / 2, 100) // straight up

      expect(result.y).toBeCloseTo(50) // sin(90°) * 100 - 50
    })
  })

  describe('generateConfettiParticle', () => {
    it('should return a color from the provided array', () => {
      const colors = ['#FF0000', '#00FF00', '#0000FF']
      const result = generateConfettiParticle(colors)

      expect(colors).toContain(result.color)
    })

    it('should return isCircle as boolean', () => {
      const result = generateConfettiParticle(CONFETTI_COLORS)

      expect(typeof result.isCircle).toBe('boolean')
    })

    it('should handle single color array', () => {
      const colors = ['#FF0000']
      const result = generateConfettiParticle(colors)

      expect(result.color).toBe('#FF0000')
    })

    it('should return fallback color for empty array', () => {
      const result = generateConfettiParticle([])

      expect(result.color).toBe('#FFD700') // fallback
    })
  })

  describe('CONFETTI_COLORS', () => {
    it('should be an array of hex colors', () => {
      expect(Array.isArray(CONFETTI_COLORS)).toBe(true)
      expect(CONFETTI_COLORS.length).toBeGreaterThan(0)

      for (const color of CONFETTI_COLORS) {
        expect(color).toMatch(/^#[0-9A-Fa-f]{6}$/)
      }
    })
  })
})
