<script setup lang="ts">
defineProps<{
  label?: string
  placeholder?: string
}>()

const modelValue = defineModel<string>()
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
  </UFormField>
</template>
