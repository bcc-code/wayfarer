// @vitest-environment nuxt
import { describe, it, expect } from 'vitest'
import { mountSuspended } from '@nuxt/test-utils/runtime'
import { UCheckbox, UInput, UTextarea } from '#components'

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
