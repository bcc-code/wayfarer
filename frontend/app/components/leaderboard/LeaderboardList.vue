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
  <ul class="space-y-2">
    <li
      v-for="entry in leaderboard"
      :key="entry.name"
      class="border-default flex items-center gap-4 rounded-xl border px-4 py-3"
    >
      <NuxtImg
        v-if="entry.image"
        :src="entry.image"
        height="32"
        width="32"
        class="rounded-full"
      />
      <div class="grow">
        <p>{{ entry.name }}</p>
        <p v-if="entry.description" class="text-dimmed text-sm">
          {{ entry.description }}
        </p>
      </div>
      <p class="font-bold tabular-nums">{{ entry.score }}</p>
    </li>
  </ul>
</template>
