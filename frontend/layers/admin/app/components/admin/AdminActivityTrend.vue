<script setup lang="ts">
import { dailyAverage } from '../../utils/sparkline'

interface TrendPoint {
  date: string
  points: number
  activeUsers: number
}

const props = defineProps<{
  trend: TrendPoint[]
  /** Days the trend covers, for the labels. */
  days?: number
}>()

// Two tiles, not one chart: points run to thousands and active users to dozens,
// so a shared axis flattens one into the baseline.
//
// Points are additive, so the window total is meaningful. Active users is a
// distinct count *per day* — summing it would count the same person once per
// day they appeared — so it is averaged. For the same reason there is no
// "unique users this week": the daily counts cannot produce it.
const tiles = computed(() => {
  if (props.trend.length === 0) return []
  const window = props.days ?? props.trend.length
  return [
    {
      key: 'points',
      label: `Poeng siste ${window} dager`,
      value: formatNumber(
        props.trend.reduce((sum, point) => sum + point.points, 0),
      ),
      points: props.trend.map((p) => ({ date: p.date, value: p.points })),
    },
    {
      key: 'activeUsers',
      label: 'Aktive deltakere per dag',
      value: formatNumber(dailyAverage(props.trend.map((p) => p.activeUsers))),
      points: props.trend.map((p) => ({ date: p.date, value: p.activeUsers })),
    },
  ]
})
</script>

<template>
  <div v-if="tiles.length" class="@container">
    <div class="grid gap-3 @2xl:grid-cols-2">
      <div
        v-for="tile in tiles"
        :key="tile.key"
        class="bg-elevated/50 rounded-lg p-3"
      >
        <p class="text-muted text-xs">{{ tile.label }}</p>
        <p class="mb-2 text-xl font-semibold">{{ tile.value }}</p>
        <AdminSparkline
          :points="tile.points"
          :label="tile.label"
          :height="36"
          empty-label="Ingen aktivitet i perioden"
        />
      </div>
    </div>
  </div>
</template>
