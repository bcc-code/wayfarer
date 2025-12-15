<script setup lang="ts">
import { cva } from 'cva'

const modelValue = defineModel<boolean>({ required: true })

const thumbClasses = cva(
  'rounded-3xl bg-white w-9 h-6 transition duration-100 ease-out shadow-medium',
  {
    variants: {
      checked: {
        true: 'translate-x-6.5',
        false: 'translate-x-0',
      },
    },
  },
)

const trackClasses = cva(
  'rounded-button-large p-1 border border-border-default w-[72px] transition duration-100 ease-out',
  {
    variants: {
      checked: {
        true: 'bg-accent-positive',
        false: 'bg-background-indent',
      },
    },
  },
)
</script>

<template>
  <label :class="trackClasses({ checked: modelValue })">
    <input type="checkbox" class="hidden" v-model="modelValue" />
    <div :class="thumbClasses({ checked: modelValue })" />
  </label>
</template>
