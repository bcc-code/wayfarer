<script setup lang="ts">
defineProps<{
  title: string
}>()
</script>

<template>
  <div class="flex h-full flex-col">
    <header class="flex items-center justify-between gap-4 px-6 pt-12 pb-3">
      <h1 class="text-heading">{{ title }}</h1>
      <slot name="action" />
    </header>
    <div class="p-list-outside grow pb-24">
      <slot />
    </div>
  </div>
</template>
