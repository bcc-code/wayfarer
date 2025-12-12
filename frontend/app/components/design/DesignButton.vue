<script setup lang="ts">
import { cva } from 'cva'

withDefaults(
  defineProps<{
    variant?: 'primary' | 'secondary' | 'tertiary'
    size?: 'small' | 'medium' | 'large'
    disabled?: boolean
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
        secondary:
          'bg-border-default text-text-default gradient-border backdrop-blur-2xl',
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
      disabled: {
        true: 'opacity-50 pointer-events-none',
        false: '',
      },
    },
  },
)
</script>

<template>
  <button :class="classes({ size, variant, disabled })">
    <slot />
  </button>
</template>
