<script setup lang="ts">
import { cva } from 'cva'

withDefaults(
  defineProps<{
    icon: string
    variant?: 'secondary' | 'tertiary'
  }>(),
  {
    variant: 'secondary',
  },
)

const classes = cva(
  'rounded-button-medium aspect-square relative grid size-11 place-items-center',
  {
    variants: {
      variant: {
        secondary: 'bg-border-default gradient-border',
        tertiary: 'text-default',
      },
    },
  },
)
</script>

<template>
  <button :class="classes({ variant })">
    <Icon :name="icon" />
  </button>
</template>
