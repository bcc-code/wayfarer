<script setup lang="ts">
import { decode } from 'blurhash'

interface ImageData {
  url: string
  width?: number | null
  height?: number | null
  blurhash?: string | null
}

const props = withDefaults(
  defineProps<{
    image?: ImageData | null
    alt?: string
    fallback?: string
    // Size for blurhash canvas (higher = sharper placeholder but more CPU)
    blurhashSize?: number
  }>(),
  {
    alt: '',
    blurhashSize: 32,
  },
)

const isLoaded = ref(false)
const hasError = ref(false)
const isCached = ref(false)
const canvasRef = ref<HTMLCanvasElement | null>(null)
const imgRef = ref<HTMLImageElement | null>(null)

// Compute aspect ratio from image dimensions
const aspectRatio = computed(() => {
  if (props.image?.width && props.image?.height) {
    return props.image.width / props.image.height
  }
  return undefined
})

// Draw blurhash to canvas
function drawBlurhash() {
  if (!canvasRef.value || !props.image?.blurhash) return

  const canvas = canvasRef.value
  const ctx = canvas.getContext('2d')
  if (!ctx) return

  try {
    // Calculate canvas dimensions maintaining aspect ratio
    const width = props.blurhashSize
    const height = aspectRatio.value
      ? Math.round(props.blurhashSize / aspectRatio.value)
      : props.blurhashSize

    canvas.width = width
    canvas.height = height

    const pixels = decode(props.image.blurhash, width, height)
    const imageData = ctx.createImageData(width, height)
    imageData.data.set(pixels)
    ctx.putImageData(imageData, 0, 0)
  } catch (e) {
    // Invalid blurhash, silently fail
    console.warn('Failed to decode blurhash:', e)
  }
}

// Draw blurhash when canvas is mounted or image changes
watch(
  [canvasRef, () => props.image?.blurhash],
  () => {
    if (canvasRef.value && props.image?.blurhash) {
      drawBlurhash()
    }
  },
  { immediate: true },
)

// Reset loaded state when image URL changes and check if cached
watch(
  () => props.image?.url,
  (url) => {
    isLoaded.value = false
    hasError.value = false
    isCached.value = false

    if (url) {
      // Create a test image to check cache status synchronously
      const testImg = new Image()
      testImg.src = url
      // If complete is true immediately after setting src, image is cached
      if (testImg.complete) {
        isCached.value = true
        isLoaded.value = true
      }
    }
  },
  { immediate: true },
)

function onLoad() {
  isLoaded.value = true
}

function onError() {
  hasError.value = true
}

// Determine what to show
const hasBlurhash = computed(() => !!props.image?.blurhash)
const showFallback = computed(
  () => hasError.value || (!props.image?.url && props.fallback),
)
const showImage = computed(() => props.image?.url && !hasError.value)
</script>

<template>
  <div class="relative overflow-hidden">
    <!-- Blurhash placeholder - stays rendered for crossfade, hidden when cached -->
    <canvas
      v-if="hasBlurhash && !hasError && !isCached"
      ref="canvasRef"
      class="absolute inset-0 size-full object-cover transition-opacity duration-600"
      :class="isLoaded ? 'opacity-0' : 'opacity-100'"
      aria-hidden="true"
    />

    <!-- Actual image -->
    <img
      v-if="showImage"
      ref="imgRef"
      :src="image!.url"
      :alt="alt"
      :style="aspectRatio ? { aspectRatio } : undefined"
      class="size-full object-cover transition-opacity duration-600"
      :class="isLoaded || !hasBlurhash ? 'opacity-100' : 'opacity-0'"
      loading="lazy"
      @load="onLoad"
      @error="onError"
    >

    <!-- Fallback image -->
    <img
      v-else-if="showFallback && fallback"
      :src="fallback"
      :alt="alt"
      class="size-full object-cover"
    >

    <!-- Loading skeleton when no blurhash available -->
    <div
      v-else-if="!image?.url && !fallback"
      class="size-full animate-pulse bg-background-indent"
    />
  </div>
</template>
