import type { Locale } from 'vue-i18n'

export function getHiddenTreasureLocale(locale: Locale): string {
  const missingLanguages: Locale[] = ['nb', 'it', 'ml', 'pap', 'sl', 'ta', 'tr']

  if (locale.startsWith('zh')) {
    return 'zh'
  }

  if (missingLanguages.includes(locale)) {
    return 'no'
  }

  return locale
}
