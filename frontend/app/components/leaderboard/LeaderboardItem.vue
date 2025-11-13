<script setup lang="ts">
defineProps<{
  item: LeaderboardEntry
  hideImage?: boolean
}>()
</script>

<template>
  <div
    :class="[
      'rounded-list-inset hover:bg-background-indent active:bg-background-indent flex items-center gap-2.5 px-3 py-2',
      item.isMe ?? 'border-accent border',
    ]"
  >
    <span v-if="item.rank" class="text-dimmed text-end">{{ item.rank }}</span>
    <NuxtImg
      v-if="item.image && !hideImage"
      :src="item.image"
      height="40"
      width="40"
      class="bg-accent shrink-0 rounded-full"
    />
    <div class="grow">
      <p class="text-label">{{ item.name }}</p>
      <!-- <p v-if="item.description" class="text-caption text-muted">
        {{ item.description }}
      </p> -->
    </div>
    <p class="text-label text-accent-contrast tabular-nums">{{ item.score }}</p>
  </div>
</template>
