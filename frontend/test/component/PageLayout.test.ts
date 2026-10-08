// @vitest-environment nuxt
import { describe, it, expect, beforeEach } from 'vitest'
import { ref, h, nextTick } from 'vue'
import { mountSuspended, mockNuxtImport } from '@nuxt/test-utils/runtime'
import PageLayout from '../../layers/user/app/components/PageLayout.vue'
import TitleBar from '../../layers/user/app/components/TitleBar.vue'

const scrollY = ref(0)

mockNuxtImport('useWindowScroll', () => () => ({ x: ref(0), y: scrollY }))

describe('PageLayout', () => {
  beforeEach(() => {
    scrollY.value = 0
  })

  it('keeps the bar in flow, so content scrolls past it rather than under it', async () => {
    const wrapper = await mountSuspended(PageLayout, {
      props: { title: 'Innstillinger' },
    })

    expect(
      wrapper.findComponent(TitleBar).element.parentElement?.className,
    ).not.toContain('sticky')
  })

  it('does not collapse the title on scroll while the bar is not sticky', async () => {
    const wrapper = await mountSuspended(PageLayout, {
      props: { title: 'Innstillinger' },
    })
    const bar = wrapper.findComponent(TitleBar)

    scrollY.value = 5000
    await nextTick()
    expect(bar.props('titleOpacity')).toBe(0)
  })

  it('renders the title', async () => {
    const wrapper = await mountSuspended(PageLayout, {
      props: { title: 'Innstillinger' },
    })

    expect(wrapper.find('h1').text()).toBe('Innstillinger')
  })

  it('shows only the large title', async () => {
    const wrapper = await mountSuspended(PageLayout, {
      props: { title: 'Innstillinger' },
    })

    expect(wrapper.findComponent(TitleBar).props('titleOpacity')).toBe(0)
  })

  it('keeps the bar title visible when there is no large title to hand over from', async () => {
    const wrapper = await mountSuspended(PageLayout)

    expect(wrapper.find('h1').exists()).toBe(false)
    expect(wrapper.findComponent(TitleBar).props('titleOpacity')).toBe(1)
  })

  it('passes a #title slot through as the large title', async () => {
    const wrapper = await mountSuspended(PageLayout, {
      slots: { title: () => h('h1', 'Ola Nordmann') },
    })

    expect(wrapper.find('h1').text()).toBe('Ola Nordmann')
    expect(wrapper.findComponent(TitleBar).props('titleOpacity')).toBe(0)
  })
})
