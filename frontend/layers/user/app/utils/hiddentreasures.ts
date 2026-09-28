import type { Locale } from 'vue-i18n'

export function getHiddenTreasureLocale(locale: Locale): string {
  // Locales Hidden Treasures has no content for, so they fall back to 'no'.
  // Deliberately typed as string[] rather than Locale[]: it includes locales
  // that are currently commented out in nuxt.config (ml, pap, sl, ta) so the
  // mapping keeps working if they are re-enabled.
  const missingLanguages: string[] = ['nb', 'it', 'ml', 'pap', 'sl', 'ta', 'tr']

  if (locale.startsWith('zh')) {
    return 'zh'
  }

  if (missingLanguages.includes(locale)) {
    return 'no'
  }

  return locale
}
