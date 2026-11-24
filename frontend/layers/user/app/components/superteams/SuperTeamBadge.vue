<script setup lang="ts">
defineProps<{
  superteam: Partial<SuperTeam>
}>()
</script>

<template>
  <DesignImage
    :image="superteam.image"
    :alt="superteam.name"
    :style="{ 'background-color': superteam.color }"
    class="rounded-full"
  />
</template>
