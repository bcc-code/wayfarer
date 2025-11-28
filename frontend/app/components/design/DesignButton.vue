<script setup lang="ts">
import { cva } from 'cva'

withDefaults(
  defineProps<{
    variant?: 'primary' | 'secondary' | 'tertiary'
    size?: 'small' | 'medium' | 'large'
  }>(),
  {
    variant: 'primary',
    size: 'medium',
  },
)

const classes = cva(
  'relative flex items-center justify-center gap-2 text-label grow',
  {
    variants: {
      variant: {
        primary: 'bg-accent text-on-accent gradient-border',
        secondary: 'bg-border-default text-text-default gradient-border',
        tertiary: 'text-default',
      },
      size: {
        small:
          'px-button-small-horizontal py-button-small-vertical rounded-button-small',
        medium:
          'px-button-medium-horizontal py-button-medium-vertical rounded-button-medium',
        large:
          'px-button-large-horizontal py-button-large-vertical rounded-button-large',
      },
    },
  },
)
</script>

<template>
  <button :class="classes({ size, variant })">
    <slot />
  </button>
</template>
