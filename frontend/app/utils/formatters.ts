/**
 * Get initials from a full name
 * @param name - The full name
 * @returns The initials (e.g., "John Doe" => "JD")
 */
export function getInitials(name: string): string {
  const splitNames = name.split(' ')
  return splitNames
    .filter(Boolean)
    .map((name) => name[0])
    .join('')
}

/**
 * Capitalize the first letter of a string and lowercase the rest
 * @param str - The string to capitalize
 * @returns Capitalized string (e.g., "HELLO" => "Hello")
 */
export function capitalizeFirst(str: string): string {
  return str.charAt(0).toUpperCase() + str.slice(1).toLowerCase()
}

/**
 * Get the current locale from i18n
 */
function getLocale(): string {
  try {
    const nuxtApp = useNuxtApp()
    return nuxtApp.$i18n?.locale?.value || 'en'
  } catch {
    // Fallback for non-Nuxt environments (e.g., unit tests)
    return 'en'
  }
}

export function formatDate(dateString: string) {
  const date = new Date(dateString)
  return date.toLocaleDateString(getLocale(), {
    month: 'long',
    day: 'numeric',
    year: 'numeric',
  })
}

export function formatDateTime(dateString: string) {
  const date = new Date(dateString)
  return date.toLocaleString(getLocale(), {
    month: 'short',
    day: 'numeric',
    year: 'numeric',
    hour: 'numeric',
    minute: '2-digit',
  })
}

export function formatDateRange(startDate: string, endDate: string) {
  const locale = getLocale()
  const start = new Date(startDate)
  const end = new Date(endDate)

  const formatter = new Intl.DateTimeFormat(locale, {
    month: 'short',
    day: 'numeric',
    year: 'numeric',
  })
  return formatter.formatRange(start, end)
}

export function formatNumber(value: number): string {
  return value.toLocaleString(getLocale())
}

/**
 * Parse a user agent string into a readable summary
 * @param ua - The user agent string
 * @returns A human-readable summary (e.g., "Chrome 120 / macOS")
 */
export function parseUserAgent(ua: string): string {
  // Detect browser
  let browser = ''
  if (ua.includes('Firefox/')) {
    const match = ua.match(/Firefox\/(\d+)/)
    browser = `Firefox ${match?.[1] ?? ''}`
  } else if (ua.includes('Edg/')) {
    const match = ua.match(/Edg\/(\d+)/)
    browser = `Edge ${match?.[1] ?? ''}`
  } else if (ua.includes('Chrome/')) {
    const match = ua.match(/Chrome\/(\d+)/)
    browser = `Chrome ${match?.[1] ?? ''}`
  } else if (ua.includes('Safari/') && !ua.includes('Chrome')) {
    const match = ua.match(/Version\/(\d+)/)
    browser = `Safari ${match?.[1] ?? ''}`
  }

  // Detect OS
  let os = ''
  if (ua.includes('iPhone') || ua.includes('iPad')) {
    os = 'iOS'
  } else if (ua.includes('Android')) {
    os = 'Android'
  } else if (ua.includes('Mac OS X')) {
    os = 'macOS'
  } else if (ua.includes('Windows')) {
    os = 'Windows'
  } else if (ua.includes('Linux')) {
    os = 'Linux'
  }

  if (browser && os) {
    return `${browser} / ${os}`
  }
  return browser || os || ua.slice(0, 30) + '...'
}
