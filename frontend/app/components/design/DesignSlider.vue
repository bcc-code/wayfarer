<script setup lang="ts">
import { normalizeProps, useMachine } from '@zag-js/vue'
import * as slider from '@zag-js/slider'

const props = defineProps<{
  min?: number
  max?: number
  step?: number
}>()

const modelValue = defineModel<number>({ default: 0 })

const service = useMachine(slider.machine, {
  id: useId(),
  defaultValue: [modelValue.value],
  min: props.min,
  max: props.max,
  step: props.step,
  onValueChange({ value }) {
    const v = value[0]
    if (v) {
      modelValue.value = v
    }
  },
  thumbAlignment: 'contain',
})
const api = computed(() => slider.connect(service, normalizeProps))
</script>

<template>
  <div
    v-bind="api.getRootProps()"
    class="p-small bg-background-indent rounded-modal"
  >
    <div
      v-bind="api.getControlProps()"
      class="relative flex items-center h-[calc(var(--slider-thumb-height))]"
    >
      <div
        v-bind="api.getTrackProps()"
        class="w-full flex items-center mx-default"
      >
        <div v-bind="api.getRangeProps()" class="h-1.5 bg-accent" />
      </div>
      <div
        v-for="(_, index) in api.value"
        :key="index"
        v-bind="api.getThumbProps({ index })"
        class="bg-accent p-button-medium-vertical shadow-medium rounded-button-medium flex gap text-on-accent"
      >
        <input v-bind="api.getHiddenInputProps({ index })" />
        <IconChevronLeft class="size-5" />
        <IconChevronRight class="size-5" />
      </div>
    </div>
  </div>
</template>
