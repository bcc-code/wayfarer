import { describe, it, expect } from 'vitest'
import {
  formatSetting,
  normalizeSetting,
  isSettingValueValid,
} from '../../layers/admin/app/utils/settings'
import { SettingValueType } from '../../app/api/generated'

describe('settings value helpers', () => {
  describe('formatSetting', () => {
    it('indents JSON for editing', () => {
      expect(formatSetting(SettingValueType.Json, '{"a":1}')).toBe(
        '{\n  "a": 1\n}',
      )
    })

    it('leaves other types alone', () => {
      expect(formatSetting(SettingValueType.Text, 'info')).toBe('info')
      expect(formatSetting(SettingValueType.Float, '0.1')).toBe('0.1')
    })

    it('passes unparseable JSON through rather than losing the draft', () => {
      expect(formatSetting(SettingValueType.Json, '{not json')).toBe(
        '{not json',
      )
    })
  })

  // Indenting changes the string but not the value, so a staged draft must not
  // count as a change just for being formatted.
  describe('normalizeSetting', () => {
    it('makes an indented draft equal to the stored one-liner', () => {
      expect(normalizeSetting(SettingValueType.Json, '{\n  "a": 1\n}')).toBe(
        normalizeSetting(SettingValueType.Json, '{"a":1}'),
      )
    })

    it('keeps a real edit distinct', () => {
      expect(normalizeSetting(SettingValueType.Json, '{"a":2}')).not.toBe(
        normalizeSetting(SettingValueType.Json, '{"a":1}'),
      )
    })
  })

  describe('isSettingValueValid', () => {
    it.each([
      [SettingValueType.Json, '{"a":1}', true],
      [SettingValueType.Json, '{not json', false],
      [SettingValueType.Float, '0.25', true],
      [SettingValueType.Float, 'quite a lot', false],
      [SettingValueType.Float, '', false],
      [SettingValueType.Int, '42', true],
      [SettingValueType.Text, 'info', true],
      [SettingValueType.Text, '', false],
    ])('%s %s -> %s', (valueType, value, expected) => {
      expect(isSettingValueValid(valueType, value)).toBe(expected)
    })
  })
})
