// @vitest-environment nuxt
import { describe, it, expect } from 'vitest'
import { mountSuspended } from '@nuxt/test-utils/runtime'
import { UCheckbox, UInput, UModal, UTextarea } from '#components'

/**
 * A form control has to have a visible edge. The panel's glass hairline
 * (`ring-black/8`) was applied to the checkbox and the input, which left them
 * hard to make out — and inconsistent with the textarea and select, which
 * never had the override.
 */
describe('form control edges', () => {
  it.each([
    ['checkbox', UCheckbox],
    ['input', UInput],
    ['textarea', UTextarea],
  ])('draws %s with the accented hairline', async (_name, component) => {
    const wrapper = await mountSuspended(component)

    expect(wrapper.html()).toContain('ring-accented')
    expect(wrapper.html()).not.toContain('ring-black/8')
  })
})

/**
 * A dialog is read against whatever it happens to be covering. The panel used
 * to frost its floating surfaces; they are opaque now.
 */
describe('floating surfaces', () => {
  it('gives a dialog a solid background', async () => {
    const wrapper = await mountSuspended(UModal, {
      props: { open: true, title: 'Forlate siden?' },
    })

    const dialog = document.querySelector('[role="dialog"]')
    expect(dialog?.className).toContain('bg-default')
    expect(dialog?.className).not.toContain('backdrop-blur')
    wrapper.unmount()
  })
})
