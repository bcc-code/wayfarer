<script setup lang="ts">
defineProps<{
  title: string
}>()
</script>

<template>
  <div>
    <header class="px-6 pt-12 pb-3">
      <h1 class="text-heading">{{ title }}</h1>
    </header>
    <div class="p-list-outside space-y-list-section-gap">
      <slot />
    </div>
  </div>
</template>
