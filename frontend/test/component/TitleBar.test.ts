// @vitest-environment nuxt
import { describe, it, expect } from 'vitest'
import { mountSuspended } from '@nuxt/test-utils/runtime'
import TitleBar from '../../layers/user/app/components/TitleBar.vue'
import DesignDrawer from '../../layers/user/app/components/design/DesignDrawer.vue'

const SHADOW_CLASS = 'to-shadow-default'

describe('TitleBar', () => {
  it('paints the scroll shadow when asked', async () => {
    const wrapper = await mountSuspended(TitleBar, {
      props: { title: 'Innstillinger', shadow: true },
    })

    expect(wrapper.html()).toContain(SHADOW_CLASS)
  })

  it('omits the scroll shadow when told not to', async () => {
    const wrapper = await mountSuspended(TitleBar, {
      props: { title: 'Innstillinger', shadow: false },
    })

    expect(wrapper.html()).not.toContain(SHADOW_CLASS)
  })

  it('renders both titles so they can be cross-faded', async () => {
    const wrapper = await mountSuspended(TitleBar, {
      props: { title: 'Innstillinger', titleOpacity: 0.25 },
    })

    expect(wrapper.find('h1').text()).toBe('Innstillinger')
    expect(wrapper.find('p').text()).toBe('Innstillinger')
    expect(wrapper.find('p').attributes('style')).toContain('opacity: 0.25')
  })

  it('shows a sheet heading only once, with no compact title to fade in', async () => {
    const wrapper = await mountSuspended(TitleBar, {
      props: { title: 'Rediger lag', size: 'small' },
    })

    expect(wrapper.find('h1').text()).toBe('Rediger lag')
    expect(wrapper.find('p').exists()).toBe(false)
  })
})

describe('DesignDrawer', () => {
  it('turns off both bar effects so the sheet keeps its rounded top', async () => {
    const wrapper = await mountSuspended(DesignDrawer, {
      props: { title: 'Rediger lag', open: true },
    })

    const bar = wrapper.findComponent(TitleBar)
    expect(bar.props('shadow')).toBe(false)
    expect(bar.props('blurred')).toBe(false)
  })
})
