<script setup lang="ts">
import { cva } from 'cva'

defineProps<{
  disabled?: boolean
}>()

const modelValue = defineModel<boolean>({ required: true })

const thumbClasses = cva(
  'rounded-3xl bg-white w-9 h-6 transition duration-100 ease-out shadow-medium',
  {
    variants: {
      checked: {
        true: 'translate-x-6.5',
        false: 'translate-x-0',
      },
      disabled: {
        true: 'opacity-50',
        false: '',
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
      disabled: {
        true: 'opacity-50',
        false: '',
      },
    },
  },
)
</script>

<template>
  <label
    :class="trackClasses({ checked: modelValue, disabled })"
    :data-checked="modelValue"
  >
    <input v-model="modelValue" :disabled type="checkbox" class="hidden" />
    <div :class="thumbClasses({ checked: modelValue, disabled })" />
  </label>
</template>
