<script setup lang="ts">
defineProps<{
  error: Error
}>()
</script>

<template>
  <div
    class="bg-error/5 text-error flex size-full grow flex-col items-center justify-center gap-2 rounded-xl p-8 text-center"
  >
    <h3 class="text-2xl font-bold">{{ $t('error.title') }}</h3>
    <p class="text-balance opacity-50">{{ error.message }}</p>
  </div>
</template>
