// @vitest-environment nuxt
import { describe, it, expect } from 'vitest'
import { mountSuspended } from '@nuxt/test-utils/runtime'
import { CalendarDate } from '@internationalized/date'
import DateRangeField from '../../layers/admin/app/components/admin/DateRangeField.vue'

const mount = (start?: string, end?: string) =>
  mountSuspended(DateRangeField, { props: { start, end } })

const field = (wrapper: Awaited<ReturnType<typeof mount>>) =>
  wrapper.findComponent({ name: 'UInputDate' })

describe('DateRangeField', () => {
  it('parses the stored ISO strings into the range field', async () => {
    const wrapper = await mount(
      '2025-12-09T00:00:00+01:00',
      '2026-04-18T00:00:00+02:00',
    )

    const value = field(wrapper).props('modelValue') as {
      start?: CalendarDate
      end?: CalendarDate
    }
    expect(value.start?.toString()).toBe('2025-12-09')
    expect(value.end?.toString()).toBe('2026-04-18')
  })

  // Editing one end is the point of the field, so the other must survive it.
  it('writes back only the end that changed', async () => {
    const wrapper = await mount(
      '2025-12-09T00:00:00+01:00',
      '2026-04-18T00:00:00+02:00',
    )

    field(wrapper).vm.$emit('update:modelValue', {
      start: new CalendarDate(2025, 12, 9),
      end: new CalendarDate(2026, 5, 1),
    })
    await wrapper.vm.$nextTick()

    expect(wrapper.props('start')).toBe('2025-12-09T00:00:00+01:00')
    expect(wrapper.emitted('update:end')?.at(-1)?.[0]).toContain('2026-05-01')
    expect(wrapper.emitted('update:start')?.at(-1)?.[0]).toContain('2025-12-09')
  })

  // A segment being typed reports undefined; clearing the stored date then
  // would lose it.
  it('keeps a stored date when the field reports it as undefined', async () => {
    const wrapper = await mount(
      '2025-12-09T00:00:00+01:00',
      '2026-04-18T00:00:00+02:00',
    )

    field(wrapper).vm.$emit('update:modelValue', {
      start: undefined,
      end: new CalendarDate(2026, 4, 18),
    })
    await wrapper.vm.$nextTick()

    expect(wrapper.emitted('update:start')).toBeUndefined()
  })

  it('renders an empty range without throwing', async () => {
    const wrapper = await mount()

    const value = field(wrapper).props('modelValue') as {
      start?: CalendarDate
      end?: CalendarDate
    }
    expect(value.start).toBeUndefined()
    expect(value.end).toBeUndefined()
  })
})
