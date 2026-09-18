<script setup lang="ts">
import { normalizeProps, useMachine } from '@zag-js/vue'
import * as slider from '@zag-js/slider'
import { useElementSize } from '@vueuse/core'
import { cva } from 'cva'

const props = defineProps<{
  min?: number
  max?: number
  step?: number
  disabled?: boolean
}>()

const modelValue = defineModel<number>({ default: 0 })

const service = useMachine(slider.machine, {
  id: useId(),
  disabled: props.disabled,
  defaultValue: [modelValue.value],
  min: props.min,
  max: props.max,
  step: props.step,
  onValueChange({ value }) {
    const v = value[0]
    if (v !== undefined) {
      modelValue.value = v
    }
  },
  thumbAlignment: 'contain',
})
const api = computed(() => slider.connect(service, normalizeProps))

const trackRef = ref<HTMLElement | null>(null)
const { width: trackWidth } = useElementSize(trackRef)

const DOT_SIZE = 6 // w-1.5 = 6px
const DOT_GAP = 10 // gap-2.5 = 10px

const dotCount = computed(() => {
  if (trackWidth.value === 0) return 0
  // Formula: width = n * DOT_SIZE + (n - 1) * DOT_GAP
  // Solve for n: n = (width + DOT_GAP) / (DOT_SIZE + DOT_GAP)
  return Math.floor((trackWidth.value + DOT_GAP) / (DOT_SIZE + DOT_GAP))
})

const activeDotIndex = computed(() => {
  const min = props.min ?? 0
  const max = props.max ?? 100
  const value = api.value.value[0] ?? min
  const percent = (value - min) / (max - min)
  if (dotCount.value === 0) return -1
  return Math.round(percent * (dotCount.value - 1))
})

const dotClasses = cva('w-1.5 h-1.5 rounded-full', {
  variants: {
    active: {
      true: 'bg-accent',
      false: 'bg-border-default',
    },
  },
})
</script>

<template>
  <div
    v-bind="api.getRootProps()"
    class="p-small bg-background-indent rounded-modal data-disabled:cursor-not-allowed"
  >
    <div
      v-bind="api.getControlProps()"
      class="relative flex items-center h-[calc(var(--slider-thumb-height))]"
    >
      <div
        ref="trackRef"
        v-bind="api.getTrackProps()"
        class="w-full flex items-center justify-between gap-2.5 mx-default"
      >
        <span
          v-for="i in dotCount"
          :key="i"
          :class="dotClasses({ active: i - 1 <= activeDotIndex })"
        />
      </div>
      <div
        v-for="(_, index) in api.value"
        :key="index"
        v-bind="api.getThumbProps({ index })"
        class="bg-accent p-button-medium-vertical shadow-medium rounded-button-medium flex gap text-on-accent"
      >
        <input v-bind="api.getHiddenInputProps({ index })" >
        <IconChevronLeft class="size-5" />
        <IconChevronRight class="size-5" />
      </div>
    </div>
  </div>
</template>
