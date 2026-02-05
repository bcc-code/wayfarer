// Map between DB language codes and frontend locale codes
// DB uses 'no' for Norwegian, frontend uses 'nb'

const DB_TO_LOCALE: Record<string, string> = {
  no: 'nb',
}

const LOCALE_TO_DB: Record<string, string> = {
  nb: 'no',
}

export function dbLanguageToLocale(dbLang: string): string {
  return DB_TO_LOCALE[dbLang] || dbLang
}

export function localeToDbLanguage(locale: string): string {
  return LOCALE_TO_DB[locale] || locale
}

const LOCALE_TO_COUNTRY: Record<string, string> = {
  nb: 'no', // Norwegian Bokmål → Norway
  en: 'gb', // English → Great Britain
  et: 'ee', // Estonian → Estonia
  'zh-CN': 'cn', // Chinese Simplified → China
}

export function localeToFlagEmoji(locale: string): string {
  const countryCode = (LOCALE_TO_COUNTRY[locale] || locale).toUpperCase()
  return countryCode
    .split('')
    .map(char => String.fromCodePoint(0x1F1E6 + char.charCodeAt(0) - 65))
    .join('')
}
