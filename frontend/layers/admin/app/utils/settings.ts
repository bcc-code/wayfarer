import { SettingValueType } from '~/api/generated'

/**
 * JSON is stored as one long line. Indenting it for editing means the draft no
 * longer equals the stored string, so dirty checks compare `normalizeSetting`
 * instead.
 */
export function formatSetting(valueType: SettingValueType, value: string) {
  if (valueType !== SettingValueType.Json) return value
  try {
    return JSON.stringify(JSON.parse(value), null, 2)
  } catch {
    return value
  }
}

export function normalizeSetting(valueType: SettingValueType, value: string) {
  if (valueType !== SettingValueType.Json) return value
  try {
    return JSON.stringify(JSON.parse(value))
  } catch {
    return value
  }
}

export function isSettingValueValid(
  valueType: SettingValueType,
  value: string,
) {
  switch (valueType) {
    case SettingValueType.Json:
      try {
        JSON.parse(value)
        return true
      } catch {
        return false
      }
    case SettingValueType.Int:
    case SettingValueType.Float:
      return value !== '' && Number.isFinite(Number(value))
    default:
      return value !== ''
  }
}
