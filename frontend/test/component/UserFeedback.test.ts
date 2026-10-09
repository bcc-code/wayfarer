// @vitest-environment nuxt
import { describe, it, expect, beforeEach, vi } from 'vitest'
import { flushPromises } from '@vue/test-utils'
import { mountSuspended, mockNuxtImport } from '@nuxt/test-utils/runtime'
import UserFeedback from '../../layers/user/app/components/UserFeedback.vue'

const { submitMutationMock, analyticsMock, routeMock } = vi.hoisted(() => ({
  submitMutationMock: vi.fn(),
  analyticsMock: vi.fn(),
  routeMock: vi.fn(),
}))
mockNuxtImport('useSubmitFeedbackMutation', () => submitMutationMock)
mockNuxtImport('useAnalytics', () => analyticsMock)
mockNuxtImport('useRoute', () => routeMock)

const submitFeedback = vi.fn()

// DesignDrawer wraps @nuxt/ui's UDrawer, which teleports its content out of the
// wrapper and only renders it while open. This stub renders both slots inline.
const stubs = {
  DesignDrawer: {
    props: ['open', 'title'],
    emits: ['update:open'],
    template: '<div><slot /><slot name="content" :close="() => {}" /></div>',
  },
}

async function mountFeedback(props: { projectId?: string } = {}) {
  submitFeedback.mockResolvedValue({ error: undefined })
  submitMutationMock.mockReturnValue({ executeMutation: submitFeedback })
  analyticsMock.mockReturnValue({ track: vi.fn() })
  routeMock.mockReturnValue({ fullPath: '/challenges?tab=open' })

  return mountSuspended(UserFeedback, { props, global: { stubs } })
}

async function send(
  wrapper: Awaited<ReturnType<typeof mountFeedback>>,
  message = 'Appen henger på utfordringer',
) {
  await wrapper.find('textarea').setValue(message)
  await wrapper.find('form').trigger('submit')
  await flushPromises()
}

const submittedInput = () => submitFeedback.mock.calls[0]![0].input

describe('UserFeedback', () => {
  beforeEach(() => {
    vi.clearAllMocks()
  })

  it('sends the page the user was on as context', async () => {
    const wrapper = await mountFeedback()

    await send(wrapper)

    expect(submittedInput().device.contextUrl).toBe('/challenges?tab=open')
  })

  it('sends the project when one is provided', async () => {
    const wrapper = await mountFeedback({ projectId: 'PR01' })

    await send(wrapper)

    expect(submittedInput().projectId).toBe('PR01')
  })

  it('submits the trimmed message', async () => {
    const wrapper = await mountFeedback()

    await send(wrapper, '  noe er galt  ')

    expect(submittedInput().message).toBe('noe er galt')
  })

  it('does not submit an empty message', async () => {
    const wrapper = await mountFeedback()

    await send(wrapper, '   ')

    expect(submitFeedback).not.toHaveBeenCalled()
  })
})
