<script setup lang="ts">
const props = defineProps<{
  count: number
  /** Participants in the project; omitted when it is not known yet. */
  total?: number
  bar?: boolean
}>()

const share = computed(() => {
  if (!props.total) return undefined
  return Math.round((props.count / props.total) * 100)
})
</script>

<template>
  <div>
    <span class="font-medium tabular-nums">{{ formatNumber(count) }}</span>
    <span v-if="share !== undefined" class="text-muted ml-1 text-xs">
      av {{ formatNumber(total ?? 0) }} · {{ share }} %
    </span>
    <div
      v-if="bar && share !== undefined"
      class="bg-elevated mt-1 h-1.5 overflow-hidden rounded-full"
    >
      <div
        class="bg-primary h-full rounded-full"
        :style="{ width: `${Math.min(share, 100)}%` }"
      />
    </div>
  </div>
</template>
