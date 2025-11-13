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
      { 'bg-accent! text-on-accent!': item.isMe },
    ]"
  >
    <span
      v-if="item.rank"
      :class="['text-dimmed text-end', { 'text-on-accent': item.isMe }]"
    >
      {{ item.rank }}
    </span>
    <NuxtImg
      v-if="item.image && !hideImage"
      :src="item.image"
      height="40"
      width="40"
      class="bg-accent shrink-0 rounded-full"
    />
    <div class="grow">
      <p class="text-label">{{ item.name }}</p>
    </div>
    <p
      :class="[
        'text-label text-accent-contrast tabular-nums',
        { 'text-on-accent': item.isMe },
      ]"
    >
      {{ item.score }}
    </p>
  </div>
</template>
