<script setup lang="ts">
import type { D } from '@vite-pwa/assets-generator/dist/shared/assets-generator.DnoqiTld.mjs'
import { cva } from 'cva'

defineProps<{
  disabled?: boolean
  loading?: boolean
}>()

const modelValue = defineModel<boolean>({ required: true })

const thumbClasses = cva(
  'rounded-3xl bg-white w-9 h-6 transition duration-100 ease-out shadow-medium relative',
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
    <input v-model="modelValue" :disabled type="checkbox" class="hidden" >
    <div :class="thumbClasses({ checked: modelValue, disabled })">
      <Icon
        v-if="loading"
        name="svg-spinners:bars-rotate-fade"
        class="absolute text-black top-1/2 left-1/2 -translate-x-1/2 -translate-y-1/2"
      />
    </div>
  </label>
</template>
