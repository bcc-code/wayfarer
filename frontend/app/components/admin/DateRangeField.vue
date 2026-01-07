<script setup lang="ts">
import { parseDate, type DateValue } from '@internationalized/date'

const props = defineProps<{
  placeholder?: string
}>()

const start = defineModel<string>('start')
const end = defineModel<string>('end')

const isOpen = ref(false)

function toDateString(dateStr: string | undefined): string | undefined {
  if (!dateStr || dateStr.trim() === '') return undefined
  // Handle both ISO timestamps and date-only strings
  return dateStr.split('T')[0]
}

function toCalendarDate(dateStr: string | undefined): DateValue | undefined {
  if (!dateStr || dateStr.trim() === '') return undefined
  const dateOnly = toDateString(dateStr)
  if (!dateOnly) return undefined
  try {
    return parseDate(dateOnly)
  } catch {
    return undefined
  }
}

function toISOWithTimezone(dateValue: DateValue): string {
  // Create a date at 01:00:00 local time
  const date = new Date(
    dateValue.year,
    dateValue.month - 1,
    dateValue.day,
    1,
    0,
    0,
  )

  // Format as ISO string with timezone
  const year = date.getFullYear()
  const month = String(date.getMonth() + 1).padStart(2, '0')
  const day = String(date.getDate()).padStart(2, '0')
  const hours = String(date.getHours()).padStart(2, '0')
  const minutes = String(date.getMinutes()).padStart(2, '0')
  const seconds = String(date.getSeconds()).padStart(2, '0')

  // Get timezone offset
  const offset = -date.getTimezoneOffset()
  const offsetHours = Math.floor(Math.abs(offset) / 60)
  const offsetMinutes = Math.abs(offset) % 60
  const offsetSign = offset >= 0 ? '+' : '-'
  const timezoneStr = `${offsetSign}${String(offsetHours).padStart(2, '0')}:${String(offsetMinutes).padStart(2, '0')}`

  return `${year}-${month}-${day}T${hours}:${minutes}:${seconds}${timezoneStr}`
}

const range = computed<{ start: DateValue; end: DateValue } | undefined>({
  get: () => {
    const startDate = toCalendarDate(start.value)
    const endDate = toCalendarDate(end.value)

    // Only return a range if both start and end are present
    if (startDate && endDate) {
      return { start: startDate, end: endDate }
    }

    // Return undefined if either is missing
    return undefined
  },
  set: (value) => {
    if (value?.start) {
      start.value = toISOWithTimezone(value.start)
    }
    if (value?.end) {
      end.value = toISOWithTimezone(value.end)
    }
  },
})

function _formatDate(dateStr: string | undefined) {
  if (!dateStr || dateStr.trim() === '') return undefined
  return formatDate(dateStr)
}

const displayValue = computed(() => {
  const startFormatted = _formatDate(start.value)
  const endFormatted = _formatDate(end.value)

  if (startFormatted && endFormatted) {
    return `${startFormatted} - ${endFormatted}`
  }
  if (startFormatted) {
    return startFormatted
  }
  if (endFormatted) {
    return endFormatted
  }
  return props.placeholder || 'Select a date range'
})
</script>

<template>
  <UPopover v-model:open="isOpen" :ui="{ content: 'p-1' }">
    <UInput :model-value="displayValue" readonly icon="lucide:calendar" />
    <template #content>
      <UCalendar v-model="range" range variant="soft" />
    </template>
  </UPopover>
</template>
