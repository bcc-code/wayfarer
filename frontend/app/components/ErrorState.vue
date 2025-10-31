<script setup lang="ts">
defineProps<{
  error: Error
}>()
</script>

<template>
  <div
    class="size-full grow flex flex-col items-center justify-center gap-2 text-2xl"
  >
    <h3 class="text-xl font-bold">Uh oh!</h3>
    <p>{{ error.message }}</p>
  </div>
</template>
