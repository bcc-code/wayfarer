<script setup lang="ts">
defineProps<{
  error: Error
}>()
</script>

<template>
  <div
    class="size-full bg-error/5 text-error rounded-xl grow flex flex-col items-center justify-center gap-2 text-center"
  >
    <h3 class="text-2xl font-bold">Oops! Something went wrong</h3>
    <p class="opacity-50 text-balance">{{ error.message }}</p>
  </div>
</template>
