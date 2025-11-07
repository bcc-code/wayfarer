<script setup lang="ts">
interface Props {
  maxBlur?: number // Maximum blur amount in pixels
  layers?: number // Number of blur layers for smoothness
}

const props = withDefaults(defineProps<Props>(), {
  maxBlur: 8,
  layers: 4,
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
    <div
      v-for="(layer, index) in blurLayers"
      :key="index"
      class="blur-layer"
      :style="{
        backdropFilter: `blur(${layer.blur}px)`,
        WebkitBackdropFilter: `blur(${layer.blur}px)`,
        maskImage: `linear-gradient(to bottom, transparent 0%, black ${layer.maskMid}%, black 100%)`,
        WebkitMaskImage: `linear-gradient(to bottom, transparent 0%, black ${layer.maskMid}%, black 100%)`,
      }"
    />
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
