<script setup lang="ts">
import AchievementDetails from '#layers/user/app/components/achievements/AchievementDetails.vue'

/**
 * The achievement as a participant opens it, rendered by the participant
 * app's own `AchievementDetails`.
 *
 * The adapter is the whole component: the app reads an achievement off a
 * query, the form holds a draft, and which of the two images and descriptions
 * applies depends on whether it has been earned — which the form's own
 * switcher decides, because a draft has never been earned by anyone.
 */
const props = defineProps<{
  achievement: {
    name?: string
    descriptionPending?: string
    descriptionCompleted?: string
    imagePending?: string
    imageCompleted?: string
    points?: number
  }
  /** Which of the two states to show. */
  achieved?: boolean
}>()

const preview = computed(() => ({
  name: props.achievement.name || 'Utmerkelse',
  descriptionPending: props.achievement.descriptionPending || '',
  descriptionCompleted: props.achievement.descriptionCompleted || '',
  points: props.achievement.points,
  // Any timestamp will do: the app only asks whether there is one.
  achievedAt: props.achieved ? new Date().toISOString() : null,
  imagePendingObject: props.achievement.imagePending
    ? { url: props.achievement.imagePending }
    : null,
  imageCompletedObject: props.achievement.imageCompleted
    ? { url: props.achievement.imageCompleted }
    : null,
}))
</script>

<template>
  <div
    class="bg-background-default aspect-[9/19.5] w-full overflow-y-auto rounded-xl p-list-outside"
  >
    <AchievementDetails :achievement="preview" />
  </div>
</template>
