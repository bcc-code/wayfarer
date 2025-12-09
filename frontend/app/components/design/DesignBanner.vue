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
    class="flex items-start gap-3 rounded-card bg-background-raised p-default"
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
