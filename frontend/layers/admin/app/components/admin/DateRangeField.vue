<script setup lang="ts">
import { parseDate, type DateValue } from '@internationalized/date'
import { extractDateOnly, formatDateWithTimezone } from '../../utils/dates'

const start = defineModel<string>('start')
const end = defineModel<string>('end')

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

/**
 * Partial by design: `UInputDate` needs to render one end while the other is
 * half-typed. Writes are one-way for the same reason — a segment mid-edit
 * reports `undefined`, and clearing the stored date then would lose it.
 */
const range = computed<{
  start: DateValue | undefined
  end: DateValue | undefined
}>({
  get: () => ({
    start: toCalendarDate(start.value),
    end: toCalendarDate(end.value),
  }),
  set: (value) => {
    if (value?.start) start.value = toISOWithTimezone(value.start)
    if (value?.end) end.value = toISOWithTimezone(value.end)
  },
})
</script>

<template>
  <!-- Typable segments: changing only the end date means editing the end,
       which the calendar cannot do — a range calendar restarts on every click.
       The calendar is for picking a fresh range, two months at a time so one
       spanning a month boundary is visible at once. -->
  <UInputDate v-model="range" range class="w-full">
    <template #trailing>
      <UPopover :content="{ align: 'end' }" :ui="{ content: 'p-1' }">
        <UButton
          variant="ghost"
          color="neutral"
          size="xs"
          icon="lucide:calendar"
          aria-label="Åpne kalender"
        />

        <template #content>
          <UCalendar v-model="range" range :number-of-months="2" />
        </template>
      </UPopover>
    </template>
  </UInputDate>
</template>
