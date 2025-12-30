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
