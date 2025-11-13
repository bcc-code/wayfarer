<script setup lang="ts">
defineProps<{
  icon?: string
  title?: string
  description?: string
}>()
</script>

<template>
  <div
    class="flex size-full grow flex-col items-center justify-center gap-4 rounded-xl p-8 text-center"
  >
    <div
      class="bg-background-indent flex size-16 items-center justify-center rounded-full"
    >
      <UIcon
        :name="icon || 'lucide:square-dashed-mouse-pointer'"
        class="text-accent-contrast size-8"
      />
    </div>
    <div class="flex flex-col gap-1">
      <h3 class="text-label">
        {{ title || $t('emptyState') }}
      </h3>
      <p v-if="description" class="text-caption text-muted max-w-sm">
        {{ description }}
      </p>
    </div>
    <div v-if="$slots.actions" class="flex justify-center gap-2">
      <slot name="actions" />
    </div>
  </div>
</template>
