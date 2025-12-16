<script setup lang="ts">
import { cva } from 'cva'

withDefaults(
  defineProps<{
    title?: string
    shadow?: boolean
    blurred?: boolean
    size?: 'large' | 'small'
    animate?: boolean
  }>(),
  {
    size: 'large',
    animate: true,
  },
)

const { y } = useWindowScroll()
const hasScrolled = computed(() => y.value > 25)

const headerClasses = cva(
  'relative flex items-start justify-between gap-4 px-6 pb-3',
  {
    variants: {
      size: {
        large: '',
        small: '',
      },
      hasScrolled: {
        true: '',
        false: '',
      },
      animate: {
        true: '',
        false: '',
      },
    },
    compoundVariants: [
      { size: 'small', hasScrolled: true, class: 'pt-6 min-h-20' },
      { size: 'large', hasScrolled: true, class: 'pt-12 min-h-20' },
      { size: 'small', hasScrolled: false, class: 'pt-6 min-h-20' },
      { size: 'large', hasScrolled: false, class: 'pt-12 min-h-24' },
    ],
    defaultVariants: {
      size: 'large',
      hasScrolled: false,
    },
  },
)

const headingClasses = cva(
  'text-text-default transition-all duration-300 ease-out',
  {
    variants: {
      hasScrolled: {
        false: '',
        true: '',
      },
      animate: {
        true: 'absolute',
        false: 'text-heading',
      },
    },
    compoundVariants: [
      {
        hasScrolled: true,
        animate: true,
        class: 'text-label top-1/2 left-1/2 -translate-x-1/2 -translate-y-1/2',
      },
      {
        hasScrolled: false,
        animate: true,
        class: 'text-heading bottom-3 left-6 translate-x-0 translate-y-0',
      },
    ],
  },
)

const actionsClasses = cva(
  'right-6 bottom-3 size-11 transition-all duration-300 ease-out',
  {
    variants: {
      hasScrolled: {
        true: '',
        false: '',
      },
      animate: {
        true: 'absolute',
        false: '',
      },
    },
    compoundVariants: [
      {
        hasScrolled: false,
        animate: true,
        class: 'bottom-3',
      },
      {
        hasScrolled: true,
        animate: true,
        class: 'top-1/2 -translate-y-1/2',
      },
    ],
  },
)
</script>

<template>
  <ProgressiveBlur
    direction="up"
    :class="[shadow && 'from-shadow-blank/0 to-shadow-default bg-linear-to-t']"
    :enabled="blurred"
  >
    <header :class="headerClasses({ hasScrolled, size, animate })">
      <slot name="title">
        <h1 v-if="title" :class="headingClasses({ hasScrolled, animate })">
          {{ title }}
        </h1>
      </slot>
      <div :class="actionsClasses({ hasScrolled, animate })">
        <slot name="action" />
      </div>
    </header>
  </ProgressiveBlur>
</template>
