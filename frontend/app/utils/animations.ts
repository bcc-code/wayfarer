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

/**
 * Check if user prefers reduced motion.
 * Note: This must be called in browser context.
 */
export function prefersReducedMotion(): boolean {
  if (typeof window === 'undefined') return false
  return window.matchMedia('(prefers-reduced-motion: reduce)').matches
}
