// @vitest-environment nuxt
import { describe, it, expect } from 'vitest'
import { mountSuspended } from '@nuxt/test-utils/runtime'
import { CalendarDate, CalendarDateTime } from '@internationalized/date'
import AdminDateTimeField from '../../layers/admin/app/components/admin/AdminDateTimeField.vue'

const mount = (modelValue?: string) =>
  mountSuspended(AdminDateTimeField, { props: { modelValue } })

const input = (wrapper: Awaited<ReturnType<typeof mount>>) =>
  wrapper.findComponent({ name: 'UInputDate' })

describe('AdminDateTimeField', () => {
  it('parses the datetime-local string the forms store', async () => {
    const wrapper = await mount('2026-02-09T14:29')

    expect(
      (input(wrapper).props('modelValue') as CalendarDateTime).toString(),
    ).toBe('2026-02-09T14:29:00')
  })

  it('writes back a zero-padded datetime-local string', async () => {
    const wrapper = await mount('2026-02-09T14:29')

    input(wrapper).vm.$emit(
      'update:modelValue',
      new CalendarDateTime(2026, 3, 4, 9, 5),
    )
    await wrapper.vm.$nextTick()

    expect(wrapper.emitted('update:modelValue')?.at(-1)?.[0]).toBe(
      '2026-03-04T09:05',
    )
  })

  // The calendar emits a date with no time, so the stored time has to survive
  // picking a different day.
  it('keeps the time when only a date arrives', async () => {
    const wrapper = await mount('2026-02-09T14:29')

    input(wrapper).vm.$emit('update:modelValue', new CalendarDate(2026, 2, 20))
    await wrapper.vm.$nextTick()

    expect(wrapper.emitted('update:modelValue')?.at(-1)?.[0]).toBe(
      '2026-02-20T14:29',
    )
  })

  // Every field using this is optional, unlike a project's date range.
  it('clears to undefined', async () => {
    const wrapper = await mount('2026-02-09T14:29')

    input(wrapper).vm.$emit('update:modelValue', undefined)
    await wrapper.vm.$nextTick()

    expect(wrapper.emitted('update:modelValue')?.at(-1)?.[0]).toBeUndefined()
  })

  it('starts empty with no value, and survives an unparseable one', async () => {
    expect(input(await mount()).props('modelValue')).toBeUndefined()
    expect(input(await mount('not a date')).props('modelValue')).toBeUndefined()
  })
})
