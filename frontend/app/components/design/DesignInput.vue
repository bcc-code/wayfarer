<script setup lang="ts">
defineProps<{
  label?: string
}>()

const id = useId()

const modelValue = defineModel<string>()
</script>

<template>
  <div class="flex flex-col">
    <label
      v-if="label"
      :for="id"
      class="text-caption text-text-muted px-4 py-2"
    >
      {{ label }}
    </label>
    <DesignPanel class="focus-within:ring-2 focus-within:ring-accent-contrast">
      <input
        :id="id"
        v-model="modelValue"
        class="text-label w-full px-3 py-4 focus-visible:outline-none rounded-list size-full"
      >
    </DesignPanel>
  </div>
</template>
