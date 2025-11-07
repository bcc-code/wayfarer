<script setup lang="ts">
defineProps<{
  variant: 'compact' | 'expanded'
  leaderboard: {
    name: string
    score: number
    description?: string | null
    image?: string | null
  }[]
}>()
</script>

<template>
  <ul v-if="variant == 'expanded'" class="space-y-list-section-gap">
    <div
      v-for="(item, index) in leaderboard"
      :key="index"
      class="bg-background-raised rounded-list p-list-section-inset shadow-small"
    >
      <LeaderboardItem :item />
    </div>
  </ul>
  <ul
    v-else-if="variant == 'compact'"
    class="bg-background-raised shadow-small p-list-section-inset gap-list-section-inset rounded-list flex flex-col"
  >
    <template v-for="(item, index) in leaderboard" :key="index">
      <hr v-if="index > 0" class="border-border-default mx-3 h-px" />
      <LeaderboardItem :item />
    </template>
  </ul>
</template>
