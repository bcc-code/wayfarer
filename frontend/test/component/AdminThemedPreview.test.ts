// @vitest-environment nuxt
import { describe, it, expect } from 'vitest'
import { mountSuspended } from '@nuxt/test-utils/runtime'
import { h } from 'vue'
import AdminThemedPreview from '../../layers/admin/app/components/admin/AdminThemedPreview.vue'

const colors = {
  accent: '#111111',
  accentContrast: '#111111',
  onAccent: '#ffffff',
  backgroundDefault: '#efefef',
  backgroundRaised: '#ffffff',
  backgroundIndent: '#dddddd',
  textDefault: '#222222',
  textMuted: '#666666',
  textHint: '#999999',
  shadowDefault: '#000000',
  shadowBlank: '#000000',
  borderDefault: '#cccccc',
}

const mount = (props: Record<string, unknown> = {}) =>
  mountSuspended(AdminThemedPreview, {
    props,
    slots: { default: () => h('a', { href: '/challenges' }, 'Start') },
  })

describe('AdminThemedPreview', () => {
  // The app is a phone app; a card judged at panel width is judged at a width
  // no participant will ever see.
  // Width, not a device: the frame caps how wide the content is judged and
  // otherwise stays out of the way.
  it('frames every preview at one mobile width', async () => {
    const wrapper = await mount()

    expect(wrapper.html()).toContain('w-[390px]')
    expect(wrapper.html()).not.toContain('aspect-')
  })

  // The previews render the real components, links included. A preview's own
  // controls belong outside the frame, which is where the achievement form
  // keeps its state switcher.
  it('is never interactive', async () => {
    const wrapper = await mount()

    expect(wrapper.find('[inert]').exists()).toBe(true)
  })

  it('paints the project palette onto the screen', async () => {
    const wrapper = await mount({ colors: { light: colors, dark: colors } })

    expect(wrapper.html()).toContain('--color-background-default: #efefef')
  })
})
