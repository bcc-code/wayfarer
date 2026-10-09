import { describe, it, expect, vi, beforeEach } from 'vitest'

const gsapTo = vi.fn()
const reducedMotion = vi.fn(() => false)

vi.mock('gsap', () => ({ gsap: { to: gsapTo, set: vi.fn(), fromTo: vi.fn() } }))
vi.mock('../../layers/user/app/utils/animations', () => ({
  prefersReducedMotion: () => reducedMotion(),
  calculateStaggerTiming: vi.fn(),
}))

const { useButtonPress } =
  await import('../../layers/user/app/composables/useGsap')

function pressEvent(element: unknown) {
  return { currentTarget: element } as unknown as PointerEvent
}

describe('useButtonPress', () => {
  beforeEach(() => {
    gsapTo.mockClear()
    reducedMotion.mockReturnValue(false)
  })

  it('presses the element the event was delivered to', () => {
    const { pressListeners } = useButtonPress()
    const element = {} as HTMLElement

    pressListeners.pointerdown(pressEvent(element))

    expect(gsapTo).toHaveBeenCalledWith(
      element,
      expect.objectContaining({ scale: 0.97 }),
    )
  })

  it.each(['pointerup', 'pointerleave'] as const)(
    'releases the press on %s, so dragging off cancels it',
    (event) => {
      const { pressListeners } = useButtonPress()
      const element = {} as HTMLElement

      pressListeners[event](pressEvent(element))

      expect(gsapTo).toHaveBeenCalledWith(
        element,
        expect.objectContaining({ scale: 1 }),
      )
    },
  )

  it('does nothing when the event has no target element', () => {
    const { pressListeners } = useButtonPress()

    pressListeners.pointerdown(pressEvent(null))

    expect(gsapTo).not.toHaveBeenCalled()
  })

  it('does not animate under reduced motion', () => {
    reducedMotion.mockReturnValue(true)
    const { pressListeners } = useButtonPress()

    pressListeners.pointerdown(pressEvent({} as HTMLElement))
    pressListeners.pointerup(pressEvent({} as HTMLElement))

    expect(gsapTo).not.toHaveBeenCalled()
  })
})
