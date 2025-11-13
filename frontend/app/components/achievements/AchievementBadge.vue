<script setup lang="ts">
defineProps<{
  achievement: Partial<Achievement>
}>()
</script>

<template>
  <div
    class="bg-background-indent grid aspect-square place-items-center overflow-hidden rounded-full"
    v-bind="achievement"
  >
    <NuxtImg
      v-if="!achievement.hidden"
      :src="achievement.image"
      class="size-full object-cover"
    />
    <span v-else>?</span>
  </div>
</template>
