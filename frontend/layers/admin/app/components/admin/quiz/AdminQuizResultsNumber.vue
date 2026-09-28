<script setup lang="ts">
import { bucketLabel, formatDecimal } from '../../../utils/quizResultsExport'

interface Bucket {
  from: number
  to: number
  count: number
  percentage: number
}

const props = defineProps<{
  buckets: Bucket[]
  average?: number | null
  median?: number | null
  min?: number | null
  max?: number | null
}>()

// Dropped entirely when nobody answered, rather than claiming an average of 0.
const stats = computed(() => {
  if (props.average == null) return []
  return [
    { label: 'Snitt', value: formatDecimal(props.average) },
    { label: 'Median', value: formatDecimal(props.median ?? 0) },
    { label: 'Laveste', value: formatDecimal(props.min ?? 0) },
    { label: 'Høyeste', value: formatDecimal(props.max ?? 0) },
  ]
})
</script>

<template>
  <div class="space-y-4">
    <dl v-if="stats.length" class="flex flex-wrap gap-x-8 gap-y-2">
      <div v-for="stat in stats" :key="stat.label">
        <dt class="text-muted text-xs">{{ stat.label }}</dt>
        <dd class="text-lg tabular-nums">{{ stat.value }}</dd>
      </div>
    </dl>
    <div class="space-y-3">
      <AdminQuizResultBar
        v-for="bucket in buckets"
        :key="`${bucket.from}-${bucket.to}`"
        :label="bucketLabel(bucket.from, bucket.to)"
        :count="bucket.count"
        :percentage="bucket.percentage"
      />
    </div>
  </div>
</template>
