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
      `aria-hidden` on the plot with a real table beside it, rather than a
      `role="img"` summary: the tooltip must not be the only way to reach a
      value, and native `<title>` tooltips do not appear on keyboard focus. The
      table is the accessible path and costs nothing visually.
    -->
    <svg
      v-if="hasActivity"
      :viewBox="`0 0 ${VIEW_WIDTH} ${height}`"
      :height="height"
      class="w-full"
      preserveAspectRatio="none"
      aria-hidden="true"
      focusable="false"
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

      <rect
        v-for="bar in bars"
        :key="bar.index"
        :x="bar.x"
        :y="bar.y"
        :width="bar.width"
        :height="bar.height"
        rx="1"
        class="fill-current"
        :class="
          bar.index === lastIndex ? 'text-primary' : 'text-dimmed opacity-60'
        "
      >
        <!--
          The most recent day carries the accent; the rest are the de-emphasis
          hue. One series, so there is no categorical palette to validate for
          colour-vision separation — the two states are the same hue at
          different emphasis, and both are design-system tokens rather than
          project branding, which is arbitrary per project and cannot be
          contrast-checked up front.
        -->
        <title>
          {{ dayLabel(points[bar.index]!.date) }}:
          {{ format(points[bar.index]!.value) }}
        </title>
      </rect>
    </svg>

    <!--
      The `sr-only` class goes on a wrapping div, not on the <table>. Applied to
      the table itself its <caption> escapes the clipping rect and renders as
      visible text under the chart — Tailwind's `sr-only` sets no `display`, so
      the element keeps `display: table` and the caption is laid out outside the
      1px clip box.
    -->
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
