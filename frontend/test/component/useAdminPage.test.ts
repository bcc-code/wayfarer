// @vitest-environment nuxt
import { describe, it, expect, afterEach } from 'vitest'
import { mountSuspended, mockNuxtImport } from '@nuxt/test-utils/runtime'
import { computed, defineComponent, h, ref, nextTick } from 'vue'
import { useAdminPage } from '../../layers/admin/app/composables/useAdminPage'

mockNuxtImport('useCurrentProject', () => () => ({
  project: computed(() => undefined),
  projectId: computed(() => undefined),
  fetching: ref(false),
}))

mockNuxtImport('useRoute', () => () => ({
  name: 'admin-users-userId',
  params: { userId: 'US1' },
  query: {},
}))

/**
 * A getter that throws until it is "initialized", standing in for a `const`
 * declared below the `useAdminPage` call. Reading such a binding during setup
 * hits the temporal dead zone and throws exactly this error — which Nuxt turns
 * into a 500 page, not a missing breadcrumb.
 */
function tdzGetter() {
  const state = { initialized: false }
  const getter = () => {
    if (!state.initialized) {
      throw new ReferenceError(
        "can't access lexical declaration 'data' before initialization",
      )
    }
    return 'Sigve Hansen'
  }
  return { state, getter }
}

describe('useAdminPage', () => {
  /**
   * `pageCrumbs` is module-level by design — the layout renders the chrome and
   * is an *ancestor* of the page, so provide/inject cannot carry the label
   * upward. It is cleared on scope dispose, which means a test must unmount or
   * its label leaks into the next one.
   */
  const mounted: Array<{ unmount: () => void }> = []
  afterEach(() => {
    while (mounted.length) mounted.pop()!.unmount()
  })

  // Every user detail view 500'd on this: `watchEffect` runs its effect
  // synchronously on creation, and both user pages declare the query below the
  // breadcrumb call because the query needs an id computed above it.
  it('does not evaluate the page label during setup', async () => {
    const { state, getter } = tdzGetter()

    const Page = defineComponent({
      setup() {
        useAdminPage(getter)
        // Stands in for `const { data } = useSomeQuery(...)` further down.
        state.initialized = true
        return () => h('div', 'body')
      },
    })

    const wrapper = await mountSuspended(Page)
    mounted.push(wrapper)
    expect(wrapper.text()).toBe('body')
  })

  it('picks the label up once the binding exists', async () => {
    const { state, getter } = tdzGetter()
    let title: ReturnType<typeof useAdminPage>['title'] | undefined

    const Page = defineComponent({
      setup() {
        const page = useAdminPage(getter)
        title = page.title
        state.initialized = true
        return () => h('div', 'body')
      },
    })

    mounted.push(await mountSuspended(Page))
    await nextTick()

    expect(title?.value).toBe('Sigve Hansen')
  })

  it('still reacts to the label changing afterwards', async () => {
    const name = ref('Første')
    let title: ReturnType<typeof useAdminPage>['title'] | undefined

    const Page = defineComponent({
      setup() {
        const page = useAdminPage(() => name.value)
        title = page.title
        return () => h('div', 'body')
      },
    })

    mounted.push(await mountSuspended(Page))
    await nextTick()
    expect(title?.value).toBe('Første')

    name.value = 'Andre'
    await nextTick()
    expect(title?.value).toBe('Andre')
  })

  it('falls back to the nav section when a page sets no label', async () => {
    const Page = defineComponent({
      setup() {
        const page = useAdminPage()
        return () => h('div', page.title.value)
      },
    })

    const wrapper = await mountSuspended(Page)
    mounted.push(wrapper)
    expect(wrapper.text()).toBe('Brukere')
  })
})
