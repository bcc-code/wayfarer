// @vitest-environment nuxt
import { describe, it, expect } from 'vitest'
import { mountSuspended } from '@nuxt/test-utils/runtime'
import AdminPushNotificationPreview from '../../layers/admin/app/components/admin/AdminPushNotificationPreview.vue'

const mount = (props: Record<string, unknown>) =>
  mountSuspended(AdminPushNotificationPreview, { props })

describe('AdminPushNotificationPreview', () => {
  it('shows the name as the title and the text as the body', async () => {
    const wrapper = await mount({
      title: 'Leseplan uke 1',
      body: 'Du er meldt på – kom i gang!',
    })

    expect(wrapper.text()).toContain('Leseplan uke 1')
    expect(wrapper.text()).toContain('Du er meldt på – kom i gang!')
  })

  // An empty field is not an empty notification; it is no notification.
  it('says that an empty text sends nothing', async () => {
    const wrapper = await mount({ title: 'Leseplan uke 1' })

    expect(wrapper.text()).toContain('sendes ingen varsling')
  })

  // A phone shows two lines and cuts the rest, so a longer message is not a
  // longer notification.
  it('clamps the body the way a phone does', async () => {
    const wrapper = await mount({ title: 'Kort', body: 'x'.repeat(400) })

    expect(wrapper.html()).toContain('line-clamp-2')
  })
})
