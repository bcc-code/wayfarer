<script setup lang="ts">
defineProps<{
  title: string
  value: string
  subtitle?: string
  icon?: string
}>()
</script>

<template>
  <UCard variant="subtle">
    <div class="space-y-2">
      <p class="text-sm text-muted">
        {{ title }}
      </p>
      <p class="text-4xl font-bold tabular-nums">
        {{ value }}
      </p>
      <p v-if="subtitle" class="text-sm text-dimmed">
        {{ subtitle }}
      </p>
    </div>
  </UCard>
</template>
