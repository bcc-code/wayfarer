import { gsap } from 'gsap'
import {
  prefersReducedMotion,
  calculateStaggerTiming,
  calculateParticleTrajectory,
  generateConfettiParticle,
  CONFETTI_COLORS,
} from '~/utils/animations'

/**
 * Shake animation for error feedback
 */
export function useShake() {
  function shake(element: HTMLElement | null) {
    if (!element || prefersReducedMotion()) return

    gsap.fromTo(
      element,
      { x: 0 },
      {
        x: 8,
        duration: 0.08,
        repeat: 5,
        yoyo: true,
        ease: 'power2.inOut',
        onComplete: () => {
          gsap.set(element, { x: 0 })
        },
      },
    )
  }

  return { shake }
}

/**
 * Pulse/scale animation for success feedback
 */
export function usePulse() {
  function pulse(element: HTMLElement | null) {
    if (!element || prefersReducedMotion()) return

    gsap.fromTo(
      element,
      { scale: 1 },
      {
        scale: 1.05,
        duration: 0.15,
        repeat: 1,
        yoyo: true,
        ease: 'power2.out',
      },
    )
  }

  return { pulse }
}

/**
 * Count up animation for numbers
 */
export function useCountUp() {
  function countUp(
    element: HTMLElement | null,
    targetValue: number,
    duration = 0.8,
  ) {
    if (!element || prefersReducedMotion()) {
      if (element) element.textContent = String(targetValue)
      return
    }

    const obj = { value: 0 }
    gsap.to(obj, {
      value: targetValue,
      duration,
      ease: 'power2.out',
      onUpdate: () => {
        element.textContent = Math.round(obj.value).toString()
      },
    })
  }

  return { countUp }
}

interface StaggeredEntranceOptions {
  /** Total animation duration in seconds (default: 0.8) */
  totalDuration?: number
}

/**
 * Staggered entrance animation for lists of elements.
 * Duration and stagger are calculated automatically based on totalDuration and element count.
 */
export function useStaggeredEntrance(options?: StaggeredEntranceOptions) {
  const totalDuration = options?.totalDuration ?? 0.8

  let ctx: gsap.Context | null = null

  function animate(elements: HTMLElement[] | NodeListOf<Element>) {
    if (!elements || elements.length === 0) return
    if (prefersReducedMotion()) return

    // Clean up previous context
    ctx?.revert()

    const { duration, stagger } = calculateStaggerTiming(
      totalDuration,
      elements.length,
    )

    ctx = gsap.context(() => {
      gsap.fromTo(
        elements,
        {
          opacity: 0,
          y: 20,
        },
        {
          opacity: 1,
          y: 0,
          duration,
          stagger,
          ease: 'power2.out',
        },
      )
    })
  }

  function cleanup() {
    ctx?.revert()
  }

  onUnmounted(cleanup)

  return { animate, cleanup }
}

/**
 * Confetti burst animation for celebrations
 */
export function useConfetti() {
  let ctx: gsap.Context | null = null

  function burst(container: HTMLElement | null) {
    if (!container || prefersReducedMotion()) return

    // Clean up previous
    ctx?.revert()

    const particleCount = 50
    const particles: HTMLElement[] = []

    ctx = gsap.context(() => {
      // Create particles
      for (let i = 0; i < particleCount; i++) {
        const { color, isCircle } = generateConfettiParticle(CONFETTI_COLORS)
        const particle = document.createElement('div')
        particle.style.cssText = `
          position: absolute;
          width: 10px;
          height: 10px;
          background: ${color};
          border-radius: ${isCircle ? '50%' : '2px'};
          pointer-events: none;
          left: 50%;
          top: 50%;
          transform: translate(-50%, -50%);
        `
        container.appendChild(particle)
        particles.push(particle)
      }

      // Animate particles
      particles.forEach((particle) => {
        const angle = Math.random() * Math.PI * 2
        const velocity = 100 + Math.random() * 150
        const trajectory = calculateParticleTrajectory(angle, velocity)

        gsap.to(particle, {
          x: trajectory.x,
          y: trajectory.y,
          rotation: trajectory.rotation,
          opacity: 0,
          duration: trajectory.duration,
          ease: 'power2.out',
          onComplete: () => {
            particle.remove()
          },
        })
      })
    }, container)
  }

  function cleanup() {
    ctx?.revert()
  }

  onUnmounted(cleanup)

  return { burst, cleanup }
}

/**
 * Button press animation - scale down on press, spring back on release
 */
export function useButtonPress() {
  function onPressStart(element: HTMLElement | null) {
    if (!element || prefersReducedMotion()) return

    gsap.to(element, {
      scale: 0.97,
      duration: 0.1,
      ease: 'power2.out',
    })
  }

  function onPressEnd(element: HTMLElement | null) {
    if (!element || prefersReducedMotion()) return

    gsap.to(element, {
      scale: 1,
      duration: 0.2,
      ease: 'back.out(2)',
    })
  }

  return { onPressStart, onPressEnd }
}
