// @vitest-environment nuxt
import { describe, it, expect } from 'vitest'
import { mountSuspended } from '@nuxt/test-utils/runtime'
import EmptyState from '~/components/EmptyState.vue'

describe('EmptyState', () => {
  it('renders the title text', async () => {
    const wrapper = await mountSuspended(EmptyState, {
      props: { title: 'Nothing here yet' },
    })

    expect(wrapper.text()).toContain('Nothing here yet')
  })

  it('does not render the paragraph when title is empty', async () => {
    const wrapper = await mountSuspended(EmptyState, {
      props: { title: '' },
    })

    // v-if="title" — falsy title should omit the <p>
    expect(wrapper.find('p').exists()).toBe(false)
  })
})
