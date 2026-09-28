// @vitest-environment nuxt
import { describe, it, expect, vi } from 'vitest'
import { mountSuspended, mockNuxtImport } from '@nuxt/test-utils/runtime'
import { ref } from 'vue'
import AdminUserPicker from '../../layers/admin/app/components/admin/AdminUserPicker.vue'

const node = (id: string, name: string) => ({
  id,
  name,
  image: null,
  church: { id: 'CH1', name: 'Oslo' },
})

const searchVariables = ref<unknown>()
const searchData = ref<{
  users: { edges: { node: ReturnType<typeof node> }[] }
}>({
  users: {
    edges: [
      { node: node('US1', 'Dino Dooley') },
      { node: node('US2', 'Margarett Lesch') },
    ],
  },
})

mockNuxtImport('useAdminUserPickerSearchQuery', () => (options: never) => {
  searchVariables.value = (
    options as { variables: { value: unknown } }
  ).variables.value
  return { data: searchData, fetching: ref(false), error: ref(undefined) }
})

const selectedData = ref<
  { users: { edges: { node: ReturnType<typeof node> }[] } } | undefined
>()

mockNuxtImport('useAdminUserPickerSelectedQuery', () => () => ({
  data: selectedData,
  fetching: ref(false),
  error: ref(undefined),
}))

mockNuxtImport('useAuthReady', () => () => ({ isAuthReady: ref(true) }))

describe('AdminUserPicker', () => {
  // A score adjustment belongs to one project, so its participants are the
  // only valid targets.
  it('scopes the search to the given project', async () => {
    await mountSuspended(AdminUserPicker, {
      props: { modelValue: '', projectId: 'PR1' },
    })

    expect(searchVariables.value).toMatchObject({
      projectId: 'PR1',
      query: undefined,
    })
  })

  it('emits the selected user id, not the user', async () => {
    const onUpdate = vi.fn()
    const wrapper = await mountSuspended(AdminUserPicker, {
      props: { modelValue: '', 'onUpdate:modelValue': onUpdate },
    })

    const menu = wrapper.findComponent({ name: 'USelectMenu' })
    menu.vm.$emit('update:modelValue', node('US2', 'Margarett Lesch'))
    await wrapper.vm.$nextTick()

    expect(onUpdate).toHaveBeenCalledWith('US2')
  })

  it('emits an empty string when the selection is cleared', async () => {
    const onUpdate = vi.fn()
    const wrapper = await mountSuspended(AdminUserPicker, {
      props: { modelValue: 'US2', 'onUpdate:modelValue': onUpdate },
    })

    wrapper
      .findComponent({ name: 'USelectMenu' })
      .vm.$emit('update:modelValue', undefined)
    await wrapper.vm.$nextTick()

    expect(onUpdate).toHaveBeenCalledWith('')
  })

  // An id handed in from outside has no label until it is looked up.
  it('labels a value that arrived as a prop', async () => {
    selectedData.value = { users: { edges: [{ node: node('US9', 'Sigve') }] } }

    const wrapper = await mountSuspended(AdminUserPicker, {
      props: { modelValue: 'US9' },
    })

    expect(
      wrapper.findComponent({ name: 'USelectMenu' }).props('modelValue'),
    ).toMatchObject({ id: 'US9', name: 'Sigve' })

    selectedData.value = undefined
  })
})
