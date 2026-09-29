<script setup lang="ts">
type ProjectCardAchievement =
  ProfilePageQuery['myCurrentProject']['achievements'][number]

const props = defineProps<{
  achievement: ProjectCardAchievement
}>()

const { track } = useAnalytics()
const { openAchievementId, clearOpenAchievementId, celebrating } =
  useAchievementSheet()
const { executeMutation: markCelebrated } =
  useMarkAchievementCelebratedMutation()

const open = ref(false)
const showConfetti = ref(false)

// Determine which image to show based on achievement state
const currentImage = computed(() => {
  if (
    props.achievement.achievedAt &&
    props.achievement.imageCompletedObject?.url
  ) {
    return props.achievement.imageCompletedObject
  }
  return props.achievement.imagePendingObject
})

watch(open, (isOpen) => {
  if (isOpen) {
    track(AnalyticsEvent.AchievementClicked, {
      achievement_id: props.achievement.id,
      achievement_name: props.achievement.name,
      is_unlocked: !!props.achievement.achievedAt,
    })

    // Trigger confetti when opened for celebration
    if (celebrating.value) {
      setTimeout(() => {
        showConfetti.value = true
      }, 300)
    }
  } else {
    // When closing an uncelebrated achievement (e.g. opened via URL param), mark it celebrated
    if (props.achievement.achievedAt && !props.achievement.celebratedAt) {
      markCelebrated({ achievementId: props.achievement.id })
    }

    if (openAchievementId.value === props.achievement.id) {
      clearOpenAchievementId()
    }
  }
})

watch(
  openAchievementId,
  (id) => {
    if (id === props.achievement.id) {
      open.value = true
    }
  },
  { immediate: true },
)
</script>

<template>
  <div>
    <DesignDrawer
      v-model:open="open"
      :title="
        achievement.achievedAt
          ? $t('achievement.unlocked')
          : $t('achievement.title')
      "
    >
      <button
        class="grid aspect-square size-full place-items-center overflow-hidden rounded-full outline-none"
      >
        <DesignImage
          :image="currentImage"
          :alt="achievement.name"
          fallback="/images/achievement-placeholder.png"
          class="size-full"
        />
      </button>
      <template #content>
        <AchievementDetails :achievement :confetti="showConfetti" />
      </template>
    </DesignDrawer>
  </div>
</template>
