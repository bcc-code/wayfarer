import { describe, it, expect, beforeEach } from 'vitest'
import { useConfirm } from '../../app/composables/useConfirm'

describe('useConfirm', () => {
  beforeEach(() => {
    // The composable is a module-level singleton; make sure no dialog is left
    // open between tests.
    const { state, settle } = useConfirm()
    if (state.open) settle(false)
  })

  it('opens the dialog with the given options', () => {
    const { confirm, state } = useConfirm()

    confirm({ title: 'Slette laget?', description: 'Kan ikke angres.' })

    expect(state.open).toBe(true)
    expect(state.title).toBe('Slette laget?')
    expect(state.description).toBe('Kan ikke angres.')
  })

  it('defaults the confirm button to the error colour', () => {
    const { confirm, state } = useConfirm()

    confirm({ title: 'Slette?' })

    expect(state.color).toBe('error')
  })

  it('respects an explicit colour', () => {
    const { confirm, state } = useConfirm()

    confirm({ title: 'Publisere?', color: 'primary' })

    expect(state.color).toBe('primary')
  })

  it('resolves true and closes when confirmed', async () => {
    const { confirm, state, settle } = useConfirm()

    const pending = confirm({ title: 'Slette?' })
    settle(true)

    await expect(pending).resolves.toBe(true)
    expect(state.open).toBe(false)
  })

  it('resolves false and closes when dismissed', async () => {
    const { confirm, state, settle } = useConfirm()

    const pending = confirm({ title: 'Slette?' })
    settle(false)

    await expect(pending).resolves.toBe(false)
    expect(state.open).toBe(false)
  })

  it('never leaves a caller awaiting forever when a second confirm opens', async () => {
    const { confirm, settle } = useConfirm()

    const first = confirm({ title: 'Første' })
    const second = confirm({ title: 'Andre' })

    // The superseded dialog must resolve, otherwise its caller hangs.
    await expect(first).resolves.toBe(false)

    settle(true)
    await expect(second).resolves.toBe(true)
  })

  it('carries options over from a previous confirm without leaking them', () => {
    const { confirm, state, settle } = useConfirm()

    confirm({ title: 'Med beskrivelse', description: 'Detaljer' })
    settle(false)
    confirm({ title: 'Uten beskrivelse' })

    expect(state.title).toBe('Uten beskrivelse')
    expect(state.description).toBeUndefined()
  })
})
