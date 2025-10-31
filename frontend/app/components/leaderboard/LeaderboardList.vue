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
      class="px-4 py-3 border border-default items-center flex gap-4 rounded-xl"
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
        <p v-if="entry.description" class="text-sm text-dimmed">
          {{ entry.description }}
        </p>
      </div>
      <p class="tabular-nums font-bold">{{ entry.score }}</p>
    </li>
  </ul>
</template>
