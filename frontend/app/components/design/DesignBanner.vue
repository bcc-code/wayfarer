<script setup lang="ts">
withDefaults(
  defineProps<{
    title: string
    dismissable?: boolean
  }>(),
  {
    dismissable: true,
  },
)

const emit = defineEmits<{
  close: []
}>()
</script>

<template>
  <div class="flex flex-col gap-3 rounded-card bg-background-raised p-default">
    <div class="grow space-y-small flex justify-between">
      <p class="text-title text-text-default text-balance">{{ title }}</p>
      <DesignIconButton
        v-if="dismissable"
        icon="lucide:x"
        size="small"
        @click="emit('close')"
      />
    </div>
    <slot name="action" />
  </div>
</template>
