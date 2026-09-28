<script setup lang="ts">
import {
  parseDateTime,
  CalendarDateTime,
  type DateValue,
} from '@internationalized/date'

/** A `datetime-local` string, `YYYY-MM-DDTHH:mm`, as the admin forms store it. */
const model = defineModel<string>()

const isOpen = ref(false)

function toDateValue(value: string | undefined): CalendarDateTime | undefined {
  if (!value) return undefined
  try {
    return parseDateTime(value)
  } catch {
    return undefined
  }
}

const pad = (value: number) => String(value).padStart(2, '0')

const value = computed<DateValue | undefined>({
  get: () => toDateValue(model.value),
  set: (next) => {
    // Clearing is meaningful here: every field using this is optional, unlike
    // a project's required date range.
    if (!next) {
      model.value = undefined
      return
    }

    // The calendar emits a date with no time, so the time already stored has
    // to survive picking a different day.
    const current = toDateValue(model.value)
    const hour = next instanceof CalendarDateTime ? next.hour : current?.hour
    const minute =
      next instanceof CalendarDateTime ? next.minute : current?.minute

    model.value = `${next.year}-${pad(next.month)}-${pad(next.day)}T${pad(hour ?? 0)}:${pad(minute ?? 0)}`
  },
})
</script>

<template>
  <UInputDate v-model="value" granularity="minute" class="w-full">
    <template #trailing>
      <UPopover
        v-model:open="isOpen"
        :content="{ align: 'end' }"
        :ui="{ content: 'p-1' }"
      >
        <UButton
          variant="ghost"
          color="neutral"
          size="xs"
          icon="lucide:calendar"
          aria-label="Åpne kalender"
        />
        <template #content>
          <UCalendar
            :model-value="value"
            @update:model-value="
              (picked) => {
                value = picked as DateValue | undefined
                isOpen = false
              }
            "
          />
        </template>
      </UPopover>
    </template>
  </UInputDate>
</template>
