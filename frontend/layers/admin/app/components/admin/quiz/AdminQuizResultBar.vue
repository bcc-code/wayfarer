<script setup lang="ts">
import { formatPercent } from '../../../utils/quizResultsExport'

withDefaults(
  defineProps<{
    label: string
    count: number
    /** Share of the question's responses, 0–100. Drives the bar width. */
    percentage: number
    /**
     * Marks the correct option. Left undefined where correctness does not
     * apply — number buckets and free-text groups have no right answer.
     */
    isCorrect?: boolean
  }>(),
  { isCorrect: undefined },
)
</script>

<template>
  <!--
    One hue for every bar: options are nominal, so shading by value would
    double-encode length as colour. Correctness is an icon plus a word, never
    colour alone. Label above the bar because answer texts run to full
    sentences, which a three-column row would truncate.
  -->
  <div>
    <div class="mb-1 flex items-baseline justify-between gap-3">
      <span class="flex min-w-0 items-baseline gap-1.5 text-sm">
        <span class="truncate" :title="label">{{ label }}</span>
        <UIcon
          v-if="isCorrect"
          name="lucide:check"
          class="text-success size-4 shrink-0 self-center"
        />
        <span v-if="isCorrect" class="text-success shrink-0 text-xs"
          >Riktig</span
        >
      </span>
      <span class="text-muted shrink-0 text-sm tabular-nums">
        {{ count }}
        <span class="text-dimmed">·</span>
        {{ formatPercent(percentage) }}
      </span>
    </div>
    <div
      class="bg-accented h-2.5 w-full overflow-hidden rounded-sm"
      role="img"
      :aria-label="`${label}: ${count}, ${formatPercent(percentage)}`"
    >
      <div
        class="bg-primary h-full rounded-r-[4px] transition-[width] duration-300"
        :style="{ width: `${Math.max(0, Math.min(100, percentage))}%` }"
      />
    </div>
  </div>
</template>
