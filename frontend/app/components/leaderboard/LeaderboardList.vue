<script setup lang="ts">
defineProps<{
  variant: 'compact' | 'expanded'
  leaderboard: LeaderboardEntry[]
  hideImages?: boolean
}>()
</script>

<template>
  <div v-if="variant == 'expanded'" class="space-y-list-section-gap">
    <DesignPanel v-for="(item, index) in leaderboard" :key="index">
      <LeaderboardItem :item :hide-image="hideImages" />
    </DesignPanel>
  </div>
  <DesignPanel
    v-else-if="variant == 'compact'"
    class="gap-list-section-inset flex flex-col"
  >
    <template v-for="(item, index) in leaderboard" :key="item.id">
      <hr v-if="index > 0" class="border-border-default mx-3 h-px" />
      <LeaderboardItem :item :hide-image="hideImages" />
    </template>
  </DesignPanel>
</template>
