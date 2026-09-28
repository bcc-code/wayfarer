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
    One row of every results chart on this page, so the blocks read as one
    system rather than five.

    The bar is a single hue for every row. Shading each bar by its own value
    would double-encode length as colour and burn the only free channel on
    information the length already carries; answer options are nominal, so
    there is no order for a ramp to mean anything.

    Correctness is an icon plus a label, never colour alone — that is the one
    signal on this page a colourblind presenter must not miss.

    Label above, bar below, rather than label | bar | value on one line: answer
    texts here run from "Oslo" to a full sentence, and a three-column row either
    truncates them or collapses the bar to nothing at narrow widths.
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
    <!--
      role="img" with the row's numbers as its name: the bar is decoration on
      top of text that already states the value, so a screen reader gets the
      sentence once rather than an unlabelled graphic.
    -->
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
