// @vitest-environment nuxt
import { describe, it, expect } from 'vitest'
import { ref } from 'vue'
import { mountSuspended, mockNuxtImport } from '@nuxt/test-utils/runtime'
import OfflineNotice from '../../layers/user/app/components/OfflineNotice.vue'
import InfoBanner from '../../layers/user/app/components/InfoBanner.vue'
import DesignIconButton from '../../layers/user/app/components/design/DesignIconButton.vue'

const isOnline = ref(true)
mockNuxtImport('useOnline', () => () => isOnline)

describe('OfflineNotice', () => {
  it('stays out of the way while online', async () => {
    isOnline.value = true
    const wrapper = await mountSuspended(OfflineNotice)

    expect(wrapper.findComponent(InfoBanner).exists()).toBe(false)
  })

  it('appears when the connection drops', async () => {
    isOnline.value = false
    const wrapper = await mountSuspended(OfflineNotice)

    expect(wrapper.findComponent(InfoBanner).exists()).toBe(true)
    expect(wrapper.text()).toContain('offline')
  })

  // Connectivity is not something to dismiss — it resolves itself.
  it('cannot be dismissed', async () => {
    isOnline.value = false
    const wrapper = await mountSuspended(OfflineNotice)

    expect(wrapper.findComponent(DesignIconButton).exists()).toBe(false)
  })
})

describe('InfoBanner', () => {
  it('offers a dismiss control only when asked, and reports it', async () => {
    const plain = await mountSuspended(InfoBanner, {
      slots: { default: () => 'Hei' },
    })
    expect(plain.findComponent(DesignIconButton).exists()).toBe(false)
    expect(plain.text()).toContain('Hei')

    const wrapper = await mountSuspended(InfoBanner, {
      props: { dismissible: true },
      slots: { default: () => 'Hei' },
    })
    await wrapper.findComponent(DesignIconButton).trigger('click')

    expect(wrapper.emitted('dismiss')).toHaveLength(1)
  })
})
