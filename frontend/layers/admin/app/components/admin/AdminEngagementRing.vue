<script setup lang="ts">
const props = withDefaults(
  defineProps<{
    count: number
    /** Participants; without it the ring stays empty and only the count shows. */
    total?: number
    image?: string | null
    size?: number
  }>(),
  { size: 44 },
)

const STROKE = 3
// Clear of the ring: at STROKE + 1 the image edge read as touching the track.
const INSET = STROKE + 4

const share = computed(() =>
  props.total ? Math.min(props.count / props.total, 1) : 0,
)

const radius = computed(() => (props.size - STROKE) / 2)
const circumference = computed(() => 2 * Math.PI * radius.value)
</script>

<template>
  <div class="flex flex-col items-center gap-1">
    <div
      class="relative shrink-0"
      :style="{ width: `${size}px`, height: `${size}px` }"
    >
      <!-- -rotate-90 starts the arc at twelve o'clock. -->
      <svg
        :width="size"
        :height="size"
        class="-rotate-90"
        aria-hidden="true"
        focusable="false"
      >
        <circle
          :cx="size / 2"
          :cy="size / 2"
          :r="radius"
          fill="none"
          :stroke-width="STROKE"
          class="text-dimmed stroke-current opacity-40"
        />
        <circle
          v-if="share > 0"
          :cx="size / 2"
          :cy="size / 2"
          :r="radius"
          fill="none"
          :stroke-width="STROKE"
          stroke-linecap="round"
          :stroke-dasharray="circumference"
          :stroke-dashoffset="circumference * (1 - share)"
          class="text-primary stroke-current"
        />
      </svg>

      <!-- Own clipping box: an image of any aspect ratio stays a disc inside
           the ring. -->
      <div
        class="bg-elevated absolute flex items-center justify-center overflow-hidden rounded-full"
        :style="{ inset: `${INSET}px` }"
      >
        <img v-if="image" :src="image" class="size-full object-cover" alt="" />
        <UIcon v-else name="lucide:award" class="text-dimmed size-1/2" />
      </div>
    </div>
    <span class="text-xs font-medium tabular-nums">
      {{ formatNumber(count) }}
    </span>
  </div>
</template>
