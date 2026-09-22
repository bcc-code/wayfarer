<script setup lang="ts">
import { sparklineBars } from '../../utils/sparkline'

const props = withDefaults(
  defineProps<{
    /** One entry per day, oldest first. */
    points: Array<{ date: string; value: number }>
    /** Accessible name for the whole chart. */
    label: string
    /** Formats a value for the hover title and the table view. */
    format?: (value: number) => string
    height?: number
    /** Shown instead of the plot when nothing happened in the window. */
    emptyLabel?: string
  }>(),
  {
    format: (value: number) => String(value),
    height: 40,
    emptyLabel: 'Ingen aktivitet i perioden',
  },
)

// Fixed viewBox + preserveAspectRatio="none" stretches to the container
// without recomputing geometry on resize.
const VIEW_WIDTH = 240

const values = computed(() => props.points.map((point) => point.value))
const bars = computed(() =>
  sparklineBars(values.value, { width: VIEW_WIDTH, height: props.height }),
)

const lastIndex = computed(() => props.points.length - 1)

/** Day the pointer is over, or null. Drives both the tooltip and the lift. */
const hovered = ref<number | null>(null)
const hoveredBar = computed(() =>
  hovered.value === null ? null : (bars.value[hovered.value] ?? null),
)
const hoveredPoint = computed(() =>
  hovered.value === null ? null : (props.points[hovered.value] ?? null),
)

// A flat series fills the plot height with nothing in it, which reads as a
// broken chart. Quiet windows are normal here, so they get a sentence.
const hasActivity = computed(() => values.value.some((value) => value > 0))

const dayLabel = (date: string) =>
  new Date(date).toLocaleDateString('nb-NO', { day: 'numeric', month: 'short' })
</script>

<template>
  <div>
    <p v-if="!hasActivity" class="text-dimmed text-xs">
      {{ emptyLabel }}
    </p>

    <!-- The plot is aria-hidden and the table below is the accessible path, so
         the tooltip can be pointer-only without adding 14 tab stops. -->
    <div v-if="hasActivity" class="relative">
      <svg
        :viewBox="`0 0 ${VIEW_WIDTH} ${height}`"
        :height="height"
        class="w-full"
        preserveAspectRatio="none"
        aria-hidden="true"
        focusable="false"
        @pointerleave="hovered = null"
      >
        <!-- Baseline, so a near-empty window still reads as a chart. -->
        <rect
          :x="0"
          :y="height - 1"
          :width="VIEW_WIDTH"
          height="1"
          class="fill-current text-dimmed opacity-40"
        />

        <!-- One series: the states are one hue at different emphasis, so there
             is no categorical palette to validate. -->
        <rect
          v-for="bar in bars"
          :key="`bar-${bar.index}`"
          :x="bar.x"
          :y="bar.y"
          :width="bar.width"
          :height="bar.height"
          rx="1"
          class="fill-current transition-opacity"
          :class="[
            bar.index === lastIndex ? 'text-primary' : 'text-dimmed',
            hovered === bar.index
              ? 'opacity-100'
              : bar.index === lastIndex
                ? 'opacity-100'
                : 'opacity-60',
          ]"
        />

        <!-- Painted last so they sit above the bars. -->
        <rect
          v-for="bar in bars"
          :key="`hit-${bar.index}`"
          :x="bar.hitX"
          :y="0"
          :width="bar.hitWidth"
          :height="height"
          fill="transparent"
          @pointerenter="hovered = bar.index"
        />
      </svg>

      <!-- Positioned by percentage, so it tracks the column at any width. -->
      <div
        v-if="hoveredBar && hoveredPoint"
        data-slot="tooltip"
        class="bg-inverted text-inverted pointer-events-none absolute bottom-full z-10 mb-1 -translate-x-1/2 rounded px-2 py-1 text-xs whitespace-nowrap shadow"
        :style="{ left: `${hoveredBar.centerRatio * 100}%` }"
      >
        <span class="font-semibold">{{ format(hoveredPoint.value) }}</span>
        <span class="opacity-75"> · {{ dayLabel(hoveredPoint.date) }}</span>
      </div>
    </div>

    <div v-if="hasActivity" class="sr-only">
      <table>
        <caption>
          {{
            label
          }}
        </caption>
        <thead>
          <tr>
            <th scope="col">Dato</th>
            <th scope="col">Verdi</th>
          </tr>
        </thead>
        <tbody>
          <tr v-for="point in points" :key="point.date">
            <th scope="row">{{ dayLabel(point.date) }}</th>
            <td>{{ format(point.value) }}</td>
          </tr>
        </tbody>
      </table>
    </div>
  </div>
</template>
