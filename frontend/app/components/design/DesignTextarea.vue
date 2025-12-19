<script setup lang="ts">
const props = defineProps<{
  label?: string
  placeholder?: string
  maxlength?: number
}>()

const modelValue = defineModel<string>()

const characterCount = computed(() => modelValue.value?.length ?? 0)
const isOverLimit = computed(
  () => props.maxlength && characterCount.value > props.maxlength,
)
</script>

<template>
  <UFormField :ui="{ label: 'text-label' }" :label>
    <UTextarea
      v-model="modelValue"
      :ui="{
        base: 'bg-background-indent rounded-list! focus-visible:ring-accent text-label p-medium! placeholder:text-text-hint',
      }"
      :rows="8"
      :placeholder
      autoresize
      class="w-full"
    />
    <div v-if="maxlength" class="flex justify-end px-1 pt-1 tabular-nums">
      <span
        class="text-caption"
        :class="{
          'text-accent-negative': isOverLimit,
        }"
      >
        {{ characterCount }} / {{ maxlength }}
      </span>
    </div>
  </UFormField>
</template>
