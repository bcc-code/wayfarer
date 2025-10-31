<script setup lang="ts">
defineProps<{
  error: Error
}>()
</script>

<template>
  <div
    class="size-full grow flex flex-col items-center justify-center gap-2 text-center"
  >
    <h3 class="text-2xl font-bold">Uh oh!</h3>
    <p class="text-muted text-balance">{{ error.message }}</p>
  </div>
</template>
