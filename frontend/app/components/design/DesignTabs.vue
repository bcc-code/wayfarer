<script
  setup
  lang="ts"
  generic="
    T,
    Tab extends { key: string; label: string; value: T; enabled?: boolean }
  "
>
import { cva } from 'cva'
import { gsap } from 'gsap'

const props = withDefaults(
  defineProps<{
    tabs: Tab[]
    variant?: 'primary' | 'secondary'
  }>(),
  {
    variant: 'primary',
  },
)

const modelValue = defineModel<Tab['value']>({ required: true })

const containerClasses = cva(
  'rounded-navigation border-border-default flex border relative',
  {
    variants: {
      variant: {
        primary: 'p-small',
        secondary: 'p-navigation-inset',
      },
    },
  },
)

const buttonClasses = cva('text-accent-contrast relative grow z-10', {
  variants: {
    variant: {
      primary:
        'rounded-button-medium px-button-medium-horizontal py-button-medium-vertical text-label',
      secondary:
        'rounded-navigation-inset px-button-small-horizontal py-button-small-vertical text-tiny',
    },
    active: {
      true: '',
      false: '',
    },
  },
})

const indicatorClasses = cva(
  'absolute top-0 left-0 bg-background-raised gradient-border pointer-events-none',
  {
    variants: {
      variant: {
        primary: 'rounded-button-medium',
        secondary: 'rounded-navigation-inset',
      },
    },
  },
)

const enabledTabs = computed(() =>
  props.tabs.filter((tab) => tab.enabled !== false),
)

// Sliding indicator logic
const containerRef = ref<HTMLElement | null>(null)
const indicatorRef = ref<HTMLElement | null>(null)
const buttonRefs = ref<HTMLButtonElement[]>([])

function updateIndicator(animate = true) {
  if (!containerRef.value || !indicatorRef.value) return

  const activeIndex = enabledTabs.value.findIndex(
    (tab) => JSON.stringify(tab.value) === JSON.stringify(modelValue.value),
  )

  if (activeIndex === -1) return

  const activeButton = buttonRefs.value[activeIndex]
  if (!activeButton) return

  // Use offsetLeft/offsetTop for position relative to offsetParent (the container)
  const targetLeft = activeButton.offsetLeft
  const targetTop = activeButton.offsetTop
  const targetWidth = activeButton.offsetWidth
  const targetHeight = activeButton.offsetHeight

  const prefersReducedMotion = window.matchMedia(
    '(prefers-reduced-motion: reduce)',
  ).matches

  if (animate && !prefersReducedMotion) {
    gsap.to(indicatorRef.value, {
      left: targetLeft,
      top: targetTop,
      width: targetWidth,
      height: targetHeight,
      duration: 0.25,
      ease: 'power2.out',
    })
  } else {
    gsap.set(indicatorRef.value, {
      left: targetLeft,
      top: targetTop,
      width: targetWidth,
      height: targetHeight,
    })
  }
}

watch(modelValue, () => updateIndicator(true))

// Use ResizeObserver to detect when button content (like icons) finishes loading
let resizeObserver: ResizeObserver | null = null

onMounted(() => {
  // Initial update
  setTimeout(() => updateIndicator(false), 0)

  // Watch for size changes (e.g., when icons load)
  resizeObserver = new ResizeObserver(() => {
    updateIndicator(false)
  })

  // Observe all buttons for size changes
  buttonRefs.value.forEach((button) => {
    if (button) {
      resizeObserver?.observe(button)
    }
  })
})

onUnmounted(() => {
  resizeObserver?.disconnect()
})
</script>

<template>
  <div ref="containerRef" :class="containerClasses({ variant })">
    <!-- Sliding indicator -->
    <div ref="indicatorRef" :class="indicatorClasses({ variant })" />

    <button
      v-for="(tab, index) in enabledTabs"
      :key="tab.key"
      :ref="(el) => { if (el) buttonRefs[index] = el as HTMLButtonElement }"
      :class="
        buttonClasses({
          variant,
          active: JSON.stringify(tab.value) == JSON.stringify(modelValue),
        })
      "
      @click="modelValue = tab.value"
    >
      <slot name="tab" :tab>{{ tab.label }}</slot>
    </button>
  </div>
</template>
