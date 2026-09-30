<script setup lang="ts">
import { vConfetti } from '@neoconfetti/vue'

/**
 * What an achievement looks like when you open it: the badge, its name and
 * description, progress, and what it is worth.
 *
 * Extracted from `AchievementBadge` so the admin panel previews the real
 * thing. The badge keeps the drawer and the celebration around it; this is
 * only the contents.
 */
const props = defineProps<{
  achievement: {
    name: string
    descriptionPending: string
    descriptionCompleted: string
    points?: number | null
    achievedAt?: string | null
    imagePendingObject?: { url: string } | null
    imageCompletedObject?: { url: string } | null
    completedItemCount?: number
    totalItems?: number
  }
  /** Set while celebrating a freshly earned achievement. */
  confetti?: boolean
}>()

const { t } = useI18n()

const currentImage = computed(() =>
  props.achievement.achievedAt && props.achievement.imageCompletedObject?.url
    ? props.achievement.imageCompletedObject
    : props.achievement.imagePendingObject,
)

const description = computed(() =>
  props.achievement.achievedAt
    ? props.achievement.descriptionCompleted
    : props.achievement.descriptionPending,
)

// Awards persist even if the required items change later, so progress towards
// the current requirements is only shown while an achievement is pending.
const progressLabel = computed(() => {
  const { achievedAt, completedItemCount, totalItems } = props.achievement
  if (achievedAt || totalItems === undefined || totalItems <= 0) return null

  return t('achievement.progress', {
    completed: formatNumber(completedItemCount ?? 0),
    total: formatNumber(totalItems),
  })
})
</script>

<template>
  <div
    class="relative flex h-full flex-col items-center justify-center gap-6 overflow-hidden"
  >
    <div v-if="confetti" v-confetti />
    <div
      :class="[
        'grid aspect-square size-55 place-items-center overflow-hidden rounded-full',
        { 'shadow-large': achievement.achievedAt },
      ]"
    >
      <DesignImage
        :image="currentImage"
        :alt="achievement.name"
        fallback="/images/achievement-placeholder.png"
        class="size-full"
      />
    </div>
    <div class="flex flex-col items-center gap-1 text-center text-balance">
      <h3 class="text-heading" v-html="achievement.name" />
      <p class="text-label" v-html="description" />
      <p v-if="progressLabel" class="text-label tabular-nums text-text-muted">
        {{ progressLabel }}
      </p>
    </div>
    <div
      v-if="achievement.achievedAt && achievement.points"
      class="rounded-full bg-background-indent py-2 px-3 text-label text-accent-contrast"
    >
      +{{ formatNumber(achievement.points) }} {{ $t('points') }}
    </div>
    <div
      v-else-if="achievement.points"
      class="rounded-full bg-background-indent py-2 px-3 text-label text-text-muted"
    >
      {{ $t('givesYouXPoints', { points: formatNumber(achievement.points) }) }}
    </div>
  </div>
</template>
