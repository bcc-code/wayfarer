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

// A viewBox with a fixed width plus `preserveAspectRatio="none"` lets the SVG
// stretch to its container without recomputing geometry on resize, which a
// ResizeObserver would otherwise be needed for.
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

/**
 * A window where nothing happened gets a sentence, not a chart.
 *
 * A flat series still occupies the plot's full height, so the tile rendered as
 * a tall empty box with a stray baseline rule adrift in it — which reads as a
 * broken chart, the very thing the baseline was added to prevent. Quiet
 * stretches are normal in this domain, so this is a common state and deserves
 * to look deliberate.
 */
const hasActivity = computed(() => values.value.some((value) => value > 0))

const dayLabel = (date: string) =>
  new Date(date).toLocaleDateString('nb-NO', { day: 'numeric', month: 'short' })
</script>

<template>
  <div>
    <p v-if="!hasActivity" class="text-dimmed text-xs">
      {{ emptyLabel }}
    </p>

    <!--
      A styled tooltip, not a native SVG `<title>`: the browser's own tooltip
      takes about a second to appear, cannot be styled, and never shows on
      keyboard focus — it reads as nothing happening.

      `aria-hidden` on the plot with a real table beside it, rather than a
      `role="img"` summary. The tooltip is a pointer affordance; the table is
      the path for keyboard and screen-reader users, which is why the tooltip
      does not need its own focus handling and the chart does not add 14 tab
      stops per tile.
    -->
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
        <!--
          Baseline. Without it an all-zero window — a normal state for a project
          between bursts of activity — paints nothing at all, and the tile reads
          as a chart that failed to load rather than as zero.
        -->
        <rect
          :x="0"
          :y="height - 1"
          :width="VIEW_WIDTH"
          height="1"
          class="fill-current text-dimmed opacity-40"
        />

        <!--
          The most recent day carries the accent; the rest are the de-emphasis
          hue, and the hovered one lifts to full opacity so the chart is seen to
          respond. One series, so there is no categorical palette to validate
          for colour-vision separation — the states are the same hue at
          different emphasis, and all are design-system tokens rather than
          project branding, which is arbitrary per project and cannot be
          contrast-checked up front.
        -->
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

        <!--
          Hit columns, painted last so they sit above the bars. Full plot
          height and the full slot width including the gap: the bar itself is a
          2px sliver on a quiet day, which no one can hover, so the reader only
          has to be over the right column.
        -->
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

      <!--
        Positioned by the slot centre as a percentage, so it tracks the column
        however wide the container is — the svg stretches via
        `preserveAspectRatio="none"` rather than being re-measured.
      -->
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
