/**
 * Pure functions for animation calculations.
 * These functions are stateless and can be easily unit tested.
 */

export interface StaggerTiming {
  duration: number
  stagger: number
}

/**
 * Calculate stagger timing for a list animation.
 * Per-element duration is 50% of total, stagger fills the rest.
 */
export function calculateStaggerTiming(
  totalDuration: number,
  elementCount: number,
): StaggerTiming {
  const duration = totalDuration * 0.5
  const stagger =
    elementCount > 1 ? (totalDuration - duration) / (elementCount - 1) : 0

  return {
    duration,
    stagger: Math.max(0, stagger),
  }
}

export interface ParticleTrajectory {
  x: number
  y: number
  rotation: number
  duration: number
}

/**
 * Calculate a random particle trajectory for confetti animations.
 * Returns pure data that can be used for animation.
 */
export function calculateParticleTrajectory(
  angle: number,
  velocity: number,
  upwardBias: number = 50,
): ParticleTrajectory {
  return {
    x: Math.cos(angle) * velocity,
    y: Math.sin(angle) * velocity - upwardBias,
    rotation: Math.random() * 720 - 360,
    duration: 1 + Math.random() * 0.5,
  }
}

/**
 * Generate random confetti particle properties
 */
export function generateConfettiParticle(colors: readonly string[]): {
  color: string
  isCircle: boolean
} {
  const color =
    colors[Math.floor(Math.random() * colors.length)] ?? colors[0] ?? '#FFD700'
  return {
    color,
    isCircle: Math.random() > 0.5,
  }
}

/**
 * Check if user prefers reduced motion.
 * Note: This must be called in browser context.
 */
export function prefersReducedMotion(): boolean {
  if (typeof window === 'undefined') return false
  return window.matchMedia('(prefers-reduced-motion: reduce)').matches
}

/**
 * Default confetti colors
 */
export const CONFETTI_COLORS = [
  '#FFD700',
  '#FF6B6B',
  '#4ECDC4',
  '#45B7D1',
  '#96CEB4',
  '#FFEAA7',
]
