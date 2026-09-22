// @vitest-environment nuxt
import { describe, it, expect, vi, beforeEach } from 'vitest'
import { mountSuspended, mockNuxtImport } from '@nuxt/test-utils/runtime'
import { ref } from 'vue'
import AdminScoreAdjustmentModal from '../../layers/admin/app/components/admin/score/AdminScoreAdjustmentModal.vue'

const createAdjustment = vi.fn(() => Promise.resolve({ error: undefined }))
const toastAdd = vi.fn()

mockNuxtImport('useCreateScoreAdjustmentMutation', () => () => ({
  executeMutation: createAdjustment,
  fetching: ref(false),
}))

mockNuxtImport('useToast', () => () => ({ add: toastAdd }))
mockNuxtImport('useAuthReady', () => () => ({ isAuthReady: ref(true) }))

// The picker owns a query of its own; this suite is about the form around it.
const AdminUserPicker = {
  props: ['modelValue', 'projectId'],
  emits: ['update:modelValue'],
  template: '<div />',
}

const mount = () =>
  mountSuspended(AdminScoreAdjustmentModal, {
    props: { open: true, projectId: 'PR1' },
    global: { stubs: { AdminUserPicker } },
  })

const submitButton = (wrapper: Awaited<ReturnType<typeof mount>>) =>
  wrapper
    .findAllComponents({ name: 'UButton' })
    .find((button) => button.text() === 'Opprett justering')!

describe('AdminScoreAdjustmentModal', () => {
  beforeEach(() => {
    createAdjustment.mockClear()
    toastAdd.mockClear()
  })

  // UModal teleports its body, so the DOM under the wrapper root is empty —
  // assert on the child component, which the lookup still reaches.
  it('scopes the picker to the project', async () => {
    const wrapper = await mount()

    expect(wrapper.findComponent(AdminUserPicker).props('projectId')).toBe(
      'PR1',
    )
  })

  // Zero points would be a journal row that changes nothing, and an
  // adjustment with no user has nowhere to go.
  it('cannot be submitted without a user and a non-zero amount', async () => {
    const wrapper = await mount()

    expect(submitButton(wrapper).props('disabled')).toBe(true)

    wrapper.findComponent(AdminUserPicker).vm.$emit('update:modelValue', 'US1')
    await wrapper.vm.$nextTick()
    expect(submitButton(wrapper).props('disabled')).toBe(true)

    await wrapper.findComponent({ name: 'UInput' }).setValue(50)
    expect(submitButton(wrapper).props('disabled')).toBe(false)
  })

  it('sends the project, user and points, and reports the new row', async () => {
    const wrapper = await mount()

    wrapper.findComponent(AdminUserPicker).vm.$emit('update:modelValue', 'US1')
    await wrapper.findComponent({ name: 'UInput' }).setValue(-25)
    await submitButton(wrapper).trigger('click')

    expect(createAdjustment).toHaveBeenCalledWith({
      input: {
        projectId: 'PR1',
        userId: 'US1',
        points: -25,
        reason: undefined,
      },
    })
    expect(wrapper.emitted('created')).toHaveLength(1)
  })

  it('keeps the dialog open and does not report a failed create', async () => {
    createAdjustment.mockResolvedValueOnce({
      error: { message: 'nope' },
    } as never)

    const wrapper = await mount()
    wrapper.findComponent(AdminUserPicker).vm.$emit('update:modelValue', 'US1')
    await wrapper.findComponent({ name: 'UInput' }).setValue(10)
    await submitButton(wrapper).trigger('click')

    expect(wrapper.emitted('created')).toBeUndefined()
    expect(toastAdd).toHaveBeenCalledWith(
      expect.objectContaining({ color: 'error' }),
    )
  })
})
