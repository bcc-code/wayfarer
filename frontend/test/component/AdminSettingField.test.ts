// @vitest-environment nuxt
import { describe, it, expect } from 'vitest'
import { mountSuspended } from '@nuxt/test-utils/runtime'
import AdminSettingField from '../../layers/admin/app/components/admin/AdminSettingField.vue'
import { SettingValueType } from '../../app/api/generated'

const setting = (overrides: Record<string, unknown> = {}) => ({
  key: 'otel_sampling_ratio',
  value: '0.1',
  valueType: SettingValueType.Float,
  description: 'OpenTelemetry sampling rate',
  requiresRestart: true,
  envVar: 'OTEL_SAMPLING_RATIO',
  editable: true,
  ...overrides,
})

describe('AdminSettingField', () => {
  it('shows the key, description and the variable it overrides', async () => {
    const wrapper = await mountSuspended(AdminSettingField, {
      props: { setting: setting() },
    })

    expect(wrapper.text()).toContain('otel_sampling_ratio')
    expect(wrapper.text()).toContain('OpenTelemetry sampling rate')
    expect(wrapper.text()).toContain('OTEL_SAMPLING_RATIO')
  })

  it('warns when a change only applies after a restart', async () => {
    const wrapper = await mountSuspended(AdminSettingField, {
      props: { setting: setting() },
    })

    expect(wrapper.text()).toContain('Krever omstart')
  })

  it('stays quiet for a setting that applies immediately', async () => {
    const wrapper = await mountSuspended(AdminSettingField, {
      props: {
        setting: setting({
          key: 'log_level',
          value: 'info',
          valueType: SettingValueType.Text,
          requiresRestart: false,
        }),
      },
    })

    expect(wrapper.text()).not.toContain('Krever omstart')
  })

  it('offers no save button until the value is edited', async () => {
    const wrapper = await mountSuspended(AdminSettingField, {
      props: { setting: setting() },
    })

    expect(wrapper.text()).not.toContain('Lagre')

    await wrapper.find('input').setValue('0.5')

    expect(wrapper.text()).toContain('Lagre')
  })

  it('emits the canonical string form on save', async () => {
    const wrapper = await mountSuspended(AdminSettingField, {
      props: { setting: setting() },
    })

    await wrapper.find('input').setValue('0.5')
    await wrapper.find('button').trigger('click')

    expect(wrapper.emitted('save')).toEqual([['0.5']])
  })

  it('refuses to save a value that does not parse', async () => {
    const wrapper = await mountSuspended(AdminSettingField, {
      props: {
        setting: setting({
          key: 'frontend_config',
          value: '{}',
          valueType: SettingValueType.Json,
        }),
      },
    })

    await wrapper.find('textarea').setValue('{not json')
    await wrapper.find('button').trigger('click')

    expect(wrapper.emitted('save')).toBeUndefined()
  })

  // Writes are restricted to keys the application knows about, so a stray row
  // must not render a control that would fail on save.
  it('renders a key it cannot write as read-only', async () => {
    const wrapper = await mountSuspended(AdminSettingField, {
      props: { setting: setting({ editable: false }) },
    })

    expect(wrapper.text()).toContain('Ikke redigerbar')
    expect(wrapper.find('input').exists()).toBe(false)
  })

  // Another admin changing a setting must not leave a stale draft on screen.
  it('resyncs the draft when the stored value changes', async () => {
    const wrapper = await mountSuspended(AdminSettingField, {
      props: { setting: setting() },
    })

    await wrapper.find('input').setValue('0.5')
    await wrapper.setProps({ setting: setting({ value: '0.9' }) })

    expect((wrapper.find('input').element as HTMLInputElement).value).toBe(
      '0.9',
    )
    expect(wrapper.text()).not.toContain('Lagre')
  })
})
