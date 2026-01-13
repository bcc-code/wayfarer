<script setup lang="ts">
import { parseDate, type DateValue } from '@internationalized/date'
import { extractDateOnly, formatDateWithTimezone } from '~/utils/dates'

const props = defineProps<{
  placeholder?: string
}>()

const start = defineModel<string>('start')
const end = defineModel<string>('end')

const isOpen = ref(false)

function toCalendarDate(dateStr: string | undefined): DateValue | undefined {
  const dateOnly = extractDateOnly(dateStr)
  if (!dateOnly) return undefined
  try {
    return parseDate(dateOnly)
  } catch {
    return undefined
  }
}

function toISOWithTimezone(dateValue: DateValue): string {
  return formatDateWithTimezone(dateValue.year, dateValue.month, dateValue.day)
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
