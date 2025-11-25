<script setup lang="ts" generic="Entry extends LeaderboardEntry">
defineProps<{
  leaderboard: Entry[]
  hideImages?: boolean
  badge?: (item: Entry) => string
}>()
</script>

<template>
  <div class="space-y-list-section-gap">
    <DesignPanel v-for="(item, index) in leaderboard" :key="index">
      <LeaderboardItem
        :item
        :hide-image="hideImages"
        :badge="badge ? badge(item) : undefined"
      />
    </DesignPanel>
  </div>
</template>
