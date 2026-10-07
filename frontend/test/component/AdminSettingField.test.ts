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

const mount = (props: Record<string, unknown>) =>
  mountSuspended(AdminSettingField, { props })

describe('AdminSettingField', () => {
  it('shows the key, description and the variable it overrides', async () => {
    const wrapper = await mount({ setting: setting(), modelValue: '0.1' })

    expect(wrapper.text()).toContain('otel_sampling_ratio')
    expect(wrapper.text()).toContain('OpenTelemetry sampling rate')
    expect(wrapper.text()).toContain('OTEL_SAMPLING_RATIO')
  })

  it('warns when a change only applies after a restart', async () => {
    const wrapper = await mount({ setting: setting(), modelValue: '0.1' })

    expect(wrapper.text()).toContain('Krever omstart')
  })

  it('stays quiet for a setting that applies immediately', async () => {
    const wrapper = await mount({
      setting: setting({
        key: 'log_level',
        value: 'info',
        valueType: SettingValueType.Text,
        requiresRestart: false,
      }),
      modelValue: 'info',
    })

    expect(wrapper.text()).not.toContain('Krever omstart')
  })

  // The page owns the draft and the save button; the field only reports that
  // this row is part of the pending batch.
  it('marks itself edited once the draft diverges from the stored value', async () => {
    const wrapper = await mount({ setting: setting(), modelValue: '0.1' })
    expect(wrapper.text()).not.toContain('Endret')

    await wrapper.setProps({ modelValue: '0.5' })

    expect(wrapper.text()).toContain('Endret')
  })

  it('emits edits upward rather than saving them', async () => {
    const wrapper = await mount({ setting: setting(), modelValue: '0.1' })

    await wrapper.find('input').setValue('0.5')

    expect(wrapper.emitted('update:modelValue')).toEqual([['0.5']])
    expect(wrapper.emitted('save')).toBeUndefined()
  })

  it('renders a key it cannot write as read-only', async () => {
    const wrapper = await mount({
      setting: setting({ editable: false }),
      modelValue: '0.1',
    })

    expect(wrapper.text()).toContain('Ikke redigerbar')
    expect(wrapper.find('input').exists()).toBe(false)
  })

  describe('a JSON setting', () => {
    const jsonSetting = () =>
      setting({
        key: 'frontend_config',
        value: '{"a":1}',
        valueType: SettingValueType.Json,
        envVar: null,
        requiresRestart: false,
      })

    it('is not marked edited merely because the draft is indented', async () => {
      const wrapper = await mount({
        setting: jsonSetting(),
        modelValue: '{\n  "a": 1\n}',
      })

      expect(wrapper.text()).not.toContain('Endret')
    })

    it('flags JSON that does not parse', async () => {
      const wrapper = await mount({
        setting: jsonSetting(),
        modelValue: '{not json',
      })

      expect(wrapper.text()).toContain('Ugyldig JSON')
    })

    it('accepts JSON that parses', async () => {
      const wrapper = await mount({
        setting: jsonSetting(),
        modelValue: '{"a": 2}',
      })

      expect(wrapper.text()).not.toContain('Ugyldig JSON')
      expect(wrapper.text()).toContain('Endret')
    })
  })
})
