<script setup lang="ts" generic="Entry extends Partial<LeaderboardEntry>">
defineProps<{
  leaderboard: Entry[]
  hideImages?: boolean
  badge?: (item: Entry, index: number) => string | undefined
  hideMedals?: boolean
}>()
</script>

<template>
  <div class="space-y-list-section-gap">
    <DesignPanel v-for="(item, index) in leaderboard" :key="index">
      <LeaderboardItem
        :item
        :hide-image="hideImages"
        :badge="badge ? badge(item, index) : undefined"
        :hide-medal="hideMedals"
        :is-me="item.tags?.includes(LeaderboardEntryTag.Me)"
      />
    </DesignPanel>
  </div>
</template>
