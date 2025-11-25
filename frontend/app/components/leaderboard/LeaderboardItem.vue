<script setup lang="ts">
defineProps<{
  item: LeaderboardEntry
}>()

const colorClasses = [
  { light: 'text-[#F2BC28]', dark: 'text-[#864802]' },
  { light: 'text-[#D4D4D4]', dark: 'text-[#525252]' },
  { light: 'text-[#C47E49]', dark: 'text-[#512012]' },
]

const getColorClasses = (rank: number, mode: 'dark' | 'light') => {
  const color = colorClasses[rank - 1]
  if (color) {
    return color[mode]
  }
}
</script>

<template>
  <div
    :class="[
      'rounded-list-inset hover:bg-background-indent active:bg-background-indent flex items-center gap-2.5 px-3 py-2',
    ]"
  >
    <div
      :class="[
        'grid aspect-square size-10 place-items-center rounded-full',
        { 'border-border-default border': item.rank > 3 },
        getColorClasses(item.rank, 'dark'),
      ]"
    >
      <NuxtImg
        v-if="item.rank <= 3"
        :src="`/images/medals/${item.rank}.png`"
        class="col-span-full row-span-full object-cover"
      />
      <span class="col-span-full row-span-full">{{ item.rank }}</span>
    </div>
    <div class="grow">
      <p class="text-label">{{ item.name }}</p>
      <p class="text-caption text-muted">{{ item.name }}</p>
    </div>
    <p
      :class="['text-label tabular-nums', getColorClasses(item.rank, 'light')]"
    >
      {{ item.score }}
    </p>
  </div>
</template>
