<script setup lang="ts">
const props = defineProps<{
  label?: string
  placeholder?: string
  maxlength?: number
}>()

const modelValue = defineModel<string>()

const characterCount = computed(() => modelValue.value?.length ?? 0)

// We want to always show the count if the limit is low.
// If the limit is over the threshold, we show the count when
// user input is over half of the max length
const isShowThreshold = 500
const isShowCount = computed(() => {
  if (!props.maxlength) return false
  if (props.maxlength <= isShowThreshold) return true
  return characterCount.value >= props.maxlength / 2
})

const isOverLimit = computed(
  () => props.maxlength && characterCount.value > props.maxlength,
)
</script>

<template>
  <UFormField :ui="{ label: 'text-label' }" :label>
    <UTextarea
      v-model="modelValue"
      :ui="{
        base: 'bg-background-indent rounded-list! focus-visible:ring-accent ring-border-default text-label p-medium! placeholder:text-text-hint',
      }"
      :rows="8"
      :placeholder
      autoresize
      class="w-full"
    />
    <div
      v-if="isShowCount"
      class="flex justify-end px-medium pt-1 tabular-nums"
    >
      <span
        class="text-caption"
        :class="{
          'text-text-hint': !isOverLimit,
          'text-accent-negative': isOverLimit,
        }"
      >
        {{ characterCount }} / {{ maxlength }}
      </span>
    </div>
  </UFormField>
</template>
