<script setup lang="ts">
defineProps<{
  leaderboard: {
    name: string
    score: number
    description?: string | null
    image?: string | null
  }[]
}>()
</script>

<template>
  <ul class="space-y-list-section-gap p-list-outside">
    <li
      v-for="(entry, index) in leaderboard"
      :key="index"
      class="bg-background-raised rounded-list p-list-section-inset shadow-small"
    >
      <div class="flex items-center gap-2.5 px-3 py-2">
        <NuxtImg
          v-if="entry.image"
          :src="entry.image"
          height="40"
          width="40"
          class="bg-accent shrink-0 rounded-full"
        />
        <div class="grow">
          <p class="text-label">{{ entry.name }}</p>
          <p v-if="entry.description" class="text-caption text-muted">
            {{ entry.description }}
          </p>
        </div>
        <p class="text-label text-accent tabular-nums">{{ entry.score }}</p>
      </div>
    </li>
  </ul>
</template>
