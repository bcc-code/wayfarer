<script setup lang="ts">
import { cva } from 'cva'

withDefaults(
  defineProps<{
    icon: string
    variant?: 'primary' | 'secondary' | 'tertiary'
    size?: 'small' | 'medium' | 'large'
  }>(),
  {
    variant: 'secondary',
    size: 'medium',
  },
)

const classes = cva('aspect-square relative grid place-items-center', {
  variants: {
    variant: {
      primary: 'bg-accent text-on-accent',
      secondary: 'bg-border-default gradient-border',
      tertiary: 'text-default',
    },
    size: {
      small: 'size-9 rounded-button-small',
      medium: 'size-11 rounded-button-medium',
      large: 'size-13 rounded-button-large',
    },
  },
})
</script>

<template>
  <button :class="classes({ variant, size })">
    <Icon :name="icon" />
  </button>
</template>
