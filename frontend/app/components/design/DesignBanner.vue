<script setup lang="ts">
defineProps<{
  title: string
}>()

const emit = defineEmits<{
  close: []
}>()
</script>

<template>
  <div
    class="flex items-start gap-medium rounded-list bg-background-raised border border-border-default p-medium shadow-small"
  >
    <div class="grow space-y-small">
      <p class="text-label text-text-default">{{ title }}</p>
      <slot name="action" />
    </div>
    <DesignIconButton
      icon="lucide:x"
      variant="tertiary"
      class="-m-small"
      @click="emit('close')"
    />
  </div>
</template>
