<script setup lang="ts">
withDefaults(defineProps<{ dismissible?: boolean }>(), { dismissible: false })

defineEmits<{ dismiss: [] }>()
</script>

<template>
  <div
    class="border-border-default bg-background-default shadow-small rounded-modal border"
  >
    <div class="p-default gap-medium flex">
      <slot name="icon">
        <IconInfo class="my-1.5 size-6 shrink-0" />
      </slot>
      <div class="text-label text-text-default my-2 grow">
        <slot />
      </div>
      <DesignIconButton
        v-if="dismissible"
        size="small"
        icon="IconClose"
        @click="$emit('dismiss')"
      />
    </div>
  </div>
</template>
