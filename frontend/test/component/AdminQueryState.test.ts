// @vitest-environment nuxt
import { describe, it, expect } from 'vitest'
import { mountSuspended } from '@nuxt/test-utils/runtime'
import AdminQueryState from '../../layers/admin/app/components/admin/AdminQueryState.vue'

const CONTENT = 'THE PAGE BODY'

describe('AdminQueryState', () => {
  it('shows the body once the query has resolved', async () => {
    const wrapper = await mountSuspended(AdminQueryState, {
      props: { fetching: false },
      slots: { default: () => CONTENT },
    })

    expect(wrapper.text()).toContain(CONTENT)
  })

  it('shows a loading state on the first fetch', async () => {
    const wrapper = await mountSuspended(AdminQueryState, {
      props: { fetching: true },
      slots: { default: () => CONTENT },
    })

    expect(wrapper.text()).not.toContain(CONTENT)
  })

  // The reason this component exists. `v-if="fetching"` unmounts the page body
  // on every refetch — after a mutation, a cache refresh, a param change — so
  // the page flashes back to a spinner despite having content to show.
  it('keeps the body visible across a refetch', async () => {
    const wrapper = await mountSuspended(AdminQueryState, {
      props: { fetching: true },
      slots: { default: () => CONTENT },
    })
    expect(wrapper.text()).not.toContain(CONTENT)

    // First load resolves.
    await wrapper.setProps({ fetching: false })
    expect(wrapper.text()).toContain(CONTENT)

    // Refetch: the body must stay.
    await wrapper.setProps({ fetching: true })
    expect(wrapper.text()).toContain(CONTENT)
  })

  // Every admin query is paused until auth is ready, so `fetching` is false
  // before it ever runs. Treating that as "settled" would suppress the first
  // spinner entirely.
  it('still shows the first spinner for a query that starts paused', async () => {
    const wrapper = await mountSuspended(AdminQueryState, {
      props: { fetching: false },
      slots: { default: () => CONTENT },
    })

    await wrapper.setProps({ fetching: true })

    expect(wrapper.text()).not.toContain(CONTENT)
  })

  it('shows the error instead of the body', async () => {
    const wrapper = await mountSuspended(AdminQueryState, {
      props: { fetching: false, error: new Error('boom') },
      slots: { default: () => CONTENT },
    })

    expect(wrapper.text()).not.toContain(CONTENT)
  })

  // A failed refetch must not leave the page spinning forever.
  it('prefers the error over a stale fetch once settled', async () => {
    const wrapper = await mountSuspended(AdminQueryState, {
      props: { fetching: true },
      slots: { default: () => CONTENT },
    })
    await wrapper.setProps({ fetching: false })
    await wrapper.setProps({ fetching: true, error: new Error('boom') })

    expect(wrapper.text()).not.toContain(CONTENT)
    expect(wrapper.html()).not.toContain('animate-spin')
  })

  it('passes an error class through for panels that reserve height', async () => {
    const wrapper = await mountSuspended(AdminQueryState, {
      props: { fetching: false, error: new Error('boom'), errorClass: 'h-150' },
      slots: { default: () => CONTENT },
    })

    expect(wrapper.html()).toContain('h-150')
  })
})
