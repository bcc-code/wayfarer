import dayjs from 'dayjs'

export function calculateDuration(startDate: string, endDate: string) {
  const start = new Date(startDate)
  const end = new Date(endDate)
  const diffTime = Math.abs(end.getTime() - start.getTime())
  const diffDays = Math.ceil(diffTime / (1000 * 60 * 60 * 24))

  if (diffDays === 1) return '1 day'
  if (diffDays < 7) return `${diffDays} days`
  if (diffDays < 30) {
    const weeks = Math.floor(diffDays / 7)
    return weeks === 1 ? '1 week' : `${weeks} weeks`
  }
  const months = Math.floor(diffDays / 30)
  return months === 1 ? '1 month' : `${months} months`
}

export function getProjectStatus(startDate: string, endDate: string) {
  const now = new Date()
  const start = new Date(startDate)
  const end = new Date(endDate)

  if (start > now) return 'Upcoming'
  if (end < now) return 'Completed'
  return 'In Progress'
}

export function isWithinRange(
  date: string | number | Date,
  startDate: string | number | Date,
  endDate: string | number | Date,
) {
  const start = new Date(startDate)
  const end = new Date(endDate)
  const input = new Date(date)
  return input >= start && input <= end
}

/**
 * Extract only the date portion (YYYY-MM-DD) from an ISO timestamp string.
 */
export function extractDateOnly(
  dateStr: string | undefined,
): string | undefined {
  if (!dateStr || dateStr.trim() === '') return undefined
  return dateStr.split('T')[0]
}

/**
 * Format a date as ISO string with timezone offset.
 * Creates a date at 01:00:00 local time to avoid timezone edge cases.
 */
export function formatDateWithTimezone(
  year: number,
  month: number,
  day: number,
): string {
  // Create a date at 01:00:00 local time
  const date = new Date(year, month - 1, day, 1, 0, 0)

  // Format components
  const y = date.getFullYear()
  const m = String(date.getMonth() + 1).padStart(2, '0')
  const d = String(date.getDate()).padStart(2, '0')
  const hours = String(date.getHours()).padStart(2, '0')
  const minutes = String(date.getMinutes()).padStart(2, '0')
  const seconds = String(date.getSeconds()).padStart(2, '0')

  // Get timezone offset
  const offset = -date.getTimezoneOffset()
  const offsetHours = Math.floor(Math.abs(offset) / 60)
  const offsetMinutes = Math.abs(offset) % 60
  const offsetSign = offset >= 0 ? '+' : '-'
  const timezoneStr = `${offsetSign}${String(offsetHours).padStart(2, '0')}:${String(offsetMinutes).padStart(2, '0')}`

  return `${y}-${m}-${d}T${hours}:${minutes}:${seconds}${timezoneStr}`
}

/**
 * Convert an ISO datetime string to local datetime-local format for HTML inputs.
 * Used in admin forms to display datetime values in the user's local timezone.
 */
export function toLocalDatetimeLocal(
  isoString: string | null | undefined,
): string | undefined {
  if (!isoString) return undefined
  return dayjs(isoString).format('YYYY-MM-DDTHH:mm')
}

/**
 * Convert a datetime-local value (local time) to ISO string for sending to API.
 * Parses the input as local time and returns UTC ISO format.
 */
export function toISOString(datetimeLocal: string | undefined): string | null {
  if (!datetimeLocal) return null
  return dayjs(datetimeLocal).toISOString()
}
