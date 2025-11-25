<script setup lang="ts">
defineProps<{
  icon: string
}>()
</script>

<template>
  <button
    class="rounded-button-medium bg-border-default gradient-border grid size-11 place-items-center"
  >
    <Icon :name="icon" />
  </button>
</template>
