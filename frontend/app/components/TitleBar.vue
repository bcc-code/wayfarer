<script setup lang="ts">
import { cva } from 'cva'

withDefaults(
  defineProps<{
    title?: string
    shadow?: boolean
    blurred?: boolean
    size?: 'large' | 'small'
  }>(),
  {
    size: 'large',
  },
)

const { y } = useWindowScroll()
const hasScrolled = computed(() => y.value > 25)

const classes = cva(
  'relative flex items-center justify-between gap-4 px-6 pb-3',
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
</script>

<template>
  <ProgressiveBlur
    direction="up"
    :class="[shadow && 'from-shadow-blank/0 to-shadow-default bg-linear-to-t']"
    :enabled="blurred"
  >
    <header :class="classes({ hasScrolled, size })">
      <h1
        v-if="title"
        :class="[
          'text-text-default absolute transition-all duration-300 ease-out',
          {
            'text-heading bottom-3 left-6 translate-x-0 translate-y-0':
              !hasScrolled,
            'text-label top-1/2 left-1/2 -translate-x-1/2 -translate-y-1/2':
              hasScrolled,
          },
        ]"
      >
        {{ title }}
      </h1>
      <div
        :class="[
          'absolute right-6 bottom-3 size-11 transition-all duration-300 ease-out',
          {
            'bottom-3': !hasScrolled,
            'top-1/2 -translate-y-1/2': hasScrolled,
          },
        ]"
      >
        <slot name="action" />
      </div>
    </header>
  </ProgressiveBlur>
</template>
