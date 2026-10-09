<script setup lang="ts">
interface Props {
  maxBlur?: number // Maximum blur amount in pixels
  /**
   * Each layer is a separate masked `backdrop-filter`, and a masked backdrop
   * filter is among the most expensive things WebKit composites. One is the
   * default because both call sites sit over content that moves: the sticky
   * `TitleBar` and the fixed bottom navigation both re-blur every scroll
   * frame, so the cost is paid continuously rather than once. Raise it only
   * for a surface whose backdrop is genuinely static.
   */
  layers?: number
  direction?: 'up' | 'down'
  enabled?: boolean
}

const props = withDefaults(defineProps<Props>(), {
  maxBlur: 8,
  layers: 1,
  direction: 'down',
  enabled: true,
})

// Determine gradient direction based on prop
const gradientDirection = computed(() => {
  return props.direction === 'up' ? 'to top' : 'to bottom'
})

// Generate blur layers with increasing amounts and overlapping masks
const blurLayers = computed(() => {
  return Array.from({ length: props.layers }, (_, i) => {
    const progress = (i + 1) / props.layers
    // Use quadratic easing for smoother, more gradual blur increase
    const eased = progress * progress
    const blurAmount = eased * props.maxBlur
    // Position where this layer reaches full opacity
    const maskPosition = progress * 100
    return {
      blur: blurAmount,
      maskMid: maskPosition,
    }
  })
})
</script>

<template>
  <div class="progressive-blur">
    <template v-if="enabled">
      <div
        v-for="(layer, index) in blurLayers"
        :key="index"
        class="blur-layer"
        :style="{
          backdropFilter: `blur(${layer.blur}px)`,
          WebkitBackdropFilter: `blur(${layer.blur}px)`,
          maskImage: `linear-gradient(${gradientDirection}, transparent 0%, black ${layer.maskMid}%, black 100%)`,
          WebkitMaskImage: `linear-gradient(${gradientDirection}, transparent 0%, black ${layer.maskMid}%, black 100%)`,
        }"
      />
    </template>
    <slot />
  </div>
</template>

<style scoped>
.progressive-blur {
  position: relative;
  isolation: isolate;
}

.blur-layer {
  position: absolute;
  inset: 0;
  pointer-events: none;
}

.content {
  position: relative;
  z-index: 1;
}
</style>
