<script setup lang="ts" generic="Entry extends Partial<LeaderboardEntry>">
defineProps<{
  item: Entry
  badge?: string
  hideMedal?: boolean
  isMe?: boolean
}>()

const colorClasses = [
  { light: 'text-[#F2BC28]!', dark: 'text-[#864802]!' },
  { light: 'text-[#D4D4D4]!', dark: 'text-[#525252]!' },
  { light: 'text-[#C47E49]!', dark: 'text-[#512012]!' },
]

const getColorClasses = (rank: number | undefined, mode: 'dark' | 'light') => {
  if (!rank) return
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
        'text-text-default grid aspect-square size-10 place-items-center rounded-full',
        {
          'border-border-default border':
            hideMedal || (item.rank && item.rank > 3),
          'text-accent-contrast!': isMe && hideMedal,
        },
        !hideMedal && getColorClasses(item.rank, 'dark'),
      ]"
    >
      <NuxtImg
        v-if="!hideMedal && item.rank && item.rank <= 3"
        :src="`/images/medals/${item.rank}.png`"
        class="col-span-full row-span-full object-cover"
      />
      <span class="col-span-full row-span-full">{{ item.rank }}</span>
    </div>
    <div class="grow">
      <div class="flex gap-2">
        <p
          :class="[
            'text-label',
            { 'text-accent-contrast': isMe, 'text-text-default': !isMe },
          ]"
        >
          {{ item.name }}
        </p>
        <span
          v-if="isMe"
          class="bg-accent text-on-accent text-caption flex gap-1 rounded-full px-1.5"
        >
          {{ $t('standings.you') }}
        </span>
      </div>

      <p
        v-if="badge"
        class="text-caption text-accent-contrast flex items-center gap-0.5"
      >
        <Icon name="lucide:badge-check" class="size-3.5" />
        <span>{{ badge }}</span>
      </p>
      <p v-else-if="item.description" class="text-caption text-muted">
        {{ item.description }}
      </p>
    </div>
    <p
      :class="[
        'text-label tabular-nums',
        { 'text-accent-contrast': hideMedal || (item.rank && item.rank > 3) },
        !hideMedal && getColorClasses(item.rank, 'light'),
      ]"
    >
      {{ item.score }}
    </p>
  </div>
</template>
