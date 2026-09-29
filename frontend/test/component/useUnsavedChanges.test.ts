// @vitest-environment nuxt
import { describe, it, expect, vi, beforeEach } from 'vitest'
import { mountSuspended, mockNuxtImport } from '@nuxt/test-utils/runtime'
import { defineComponent, reactive, h } from 'vue'

const confirm = vi.fn(() => Promise.resolve(true))
mockNuxtImport('useConfirm', () => () => ({ confirm }))

const leaveGuards: (() => unknown)[] = []
mockNuxtImport('onBeforeRouteLeave', () => (guard: () => unknown) => {
  leaveGuards.push(guard)
})

const host = (setup: (state: Record<string, unknown>) => unknown) =>
  defineComponent({
    setup() {
      const state = reactive({ name: 'Kickoff' })
      const api = setup(state) as { markSaved: () => void }
      return { state, ...api }
    },
    render: () => h('div'),
  })

beforeEach(() => {
  confirm.mockClear()
  leaveGuards.length = 0
})

describe('useUnsavedChanges', () => {
  it('lets an untouched form go without asking', async () => {
    const wrapper = await mountSuspended(
      host((state) => useUnsavedChanges(() => ({ ...state }))),
    )

    await expect(leaveGuards[0]!()).resolves.toBe(true)
    expect(confirm).not.toHaveBeenCalled()
    expect(wrapper.vm.isDirty).toBe(false)
  })

  it('asks once something has changed', async () => {
    const wrapper = await mountSuspended(
      host((state) => useUnsavedChanges(() => ({ ...state }))),
    )

    wrapper.vm.state.name = 'Kickoff 2026'
    expect(wrapper.vm.isDirty).toBe(true)

    await leaveGuards[0]!()
    expect(confirm).toHaveBeenCalled()
  })

  // The form fills itself in when its query resolves; that is not an edit.
  it('treats a new baseline as clean', async () => {
    const wrapper = await mountSuspended(
      host((state) => useUnsavedChanges(() => ({ ...state }))),
    )

    wrapper.vm.state.name = 'Loaded from the server'
    wrapper.vm.markSaved()

    expect(wrapper.vm.isDirty).toBe(false)
    await expect(leaveGuards[0]!()).resolves.toBe(true)
    expect(confirm).not.toHaveBeenCalled()
  })

  // Editing back to where you started is not unsaved work.
  it('forgets a change that was undone', async () => {
    const wrapper = await mountSuspended(
      host((state) => useUnsavedChanges(() => ({ ...state }))),
    )

    wrapper.vm.state.name = 'Changed'
    wrapper.vm.state.name = 'Kickoff'

    expect(wrapper.vm.isDirty).toBe(false)
  })
})
