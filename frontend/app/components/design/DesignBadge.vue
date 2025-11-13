<script setup lang="ts">
import { cva } from 'cva'

withDefaults(
  defineProps<{
    variant?: 'primary' | 'secondary'
    icon?: string
  }>(),
  {
    variant: 'primary',
  },
)

const classes = cva(
  'flex items-center justify-center gap-1.5 text-label grow text-tiny rounded-lg px-3 py-1.5 w-max',
  {
    variants: {
      variant: {
        primary: 'bg-accent text-on-accent',
        secondary: 'bg-background-indent text-default',
      },
    },
  },
)
</script>

<template>
  <span :class="classes({ variant })">
    <UIcon v-if="icon" :name="icon" />
    <slot />
  </span>
</template>
