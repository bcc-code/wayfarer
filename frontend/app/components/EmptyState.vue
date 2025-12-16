<script setup lang="ts">
defineProps<{
  title: string
}>()
</script>

<template>
  <div
    class="flex size-full grow items-center justify-center p-default text-center"
  >
    <p v-if="title" class="text-title text-balance text-text-default w-full">
      {{ title }}
    </p>
  </div>
</template>
