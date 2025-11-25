<script setup lang="ts">
defineProps<{
  title?: string
  achievements?: Partial<Achievement>[]
}>()
</script>

<template>
  <div class="p-medium space-y-medium">
    <h2 v-if="title" class="text-title">{{ title }}</h2>
    <div class="gap-medium grid grid-cols-4">
      <AchievementBadge
        v-for="achievement in achievements"
        :key="achievement.id"
        :achievement
      />
    </div>
  </div>
</template>
