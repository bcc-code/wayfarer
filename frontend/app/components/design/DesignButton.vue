<script setup lang="ts">
import { cva } from 'cva'

withDefaults(
  defineProps<{
    variant?: 'primary' | 'secondary' | 'tertiary'
    size?: 'small' | 'medium' | 'large'
    disabled?: boolean
    loading?: boolean
  }>(),
  {
    variant: 'primary',
    size: 'medium',
    loading: false,
  },
)

const buttonRef = ref<HTMLButtonElement | null>(null)
const { onPressStart, onPressEnd } = useButtonPress()

const classes = cva('relative text-label grow will-change-transform', {
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
})
</script>

<template>
  <button
    ref="buttonRef"
    :class="classes({ size, variant, disabled })"
    @pointerdown="onPressStart(buttonRef)"
    @pointerup="onPressEnd(buttonRef)"
    @pointerleave="onPressEnd(buttonRef)"
  >
    <Transition
      enter-active-class="transition-all duration-200 ease-out"
      enter-from-class="opacity-0 scale-90"
      enter-to-class="opacity-100 scale-100"
      leave-from-class="opacity-100 scale-100"
      leave-to-class="opacity-0 scale-90"
      leave-active-class="transition-all duration-200 ease-out"
    >
      <Icon
        v-if="loading"
        name="svg-spinners:bars-rotate-fade"
        class="absolute top-1/2 left-1/2 -translate-x-1/2 -translate-y-1/2"
      />
    </Transition>
    <span
      :class="[
        'flex items-center justify-center gap-2 transition-all duration-200 ease-out',
        { 'opacity-0 translate-y-1': loading },
      ]"
    >
      <slot />
    </span>
  </button>
</template>
