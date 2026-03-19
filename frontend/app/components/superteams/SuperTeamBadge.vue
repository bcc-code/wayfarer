<script setup lang="ts">
defineProps<{
  superteam: Partial<SuperTeam>
}>()
</script>

<template>
  <DesignImage
    :image="superteam.imageObject"
    :alt="superteam.name"
    :style="{ 'background-color': superteam.color }"
    class="rounded-full"
  />
</template>
