// @vitest-environment nuxt
import { describe, it, expect } from 'vitest'
import { mountSuspended } from '@nuxt/test-utils/runtime'
import DesignSlider from '../../layers/user/app/components/design/DesignSlider.vue'

const mount = (props: Record<string, unknown>) =>
  mountSuspended(DesignSlider, { props })

/** Whatever the machine wrote into the thumb's position. */
const thumbStyle = (html: string) => {
  const styles = html.match(/style="[^"]*"/g) ?? []
  return styles.join(' ')
}

describe('DesignSlider', () => {
  it('places the thumb within a normal range', async () => {
    const wrapper = await mount({ min: 0, max: 100, modelValue: 50 })
    expect(thumbStyle(wrapper.html())).not.toContain('NaN')
  })

  /**
   * `(value - min) / (max - min)` is 0/0 when the bounds meet, and the NaN
   * reached the DOM as `calc(NaN% - 32px)`, leaving the thumb beside the
   * track instead of on it.
   */
  it('writes no NaN into the thumb when min and max meet', async () => {
    const wrapper = await mount({ min: 0, max: 0, modelValue: 0 })
    expect(thumbStyle(wrapper.html())).not.toContain('NaN')
  })

  it('writes no NaN when max is below min', async () => {
    const wrapper = await mount({ min: 10, max: 5, modelValue: 10 })
    expect(thumbStyle(wrapper.html())).not.toContain('NaN')
  })

  it('cannot be dragged when there is only one reachable value', async () => {
    const wrapper = await mount({ min: 0, max: 0, modelValue: 0 })
    expect(wrapper.find('[data-disabled]').exists()).toBe(true)
  })

  it('stays interactive when there is a range', async () => {
    const wrapper = await mount({ min: 0, max: 10, modelValue: 5 })
    expect(wrapper.find('[data-disabled]').exists()).toBe(false)
  })
})
