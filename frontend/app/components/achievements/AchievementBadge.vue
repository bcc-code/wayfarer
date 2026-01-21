<script setup lang="ts">
type ProjectCardAchievement =
  ProfilePageQuery['myCurrentProject']['achievements'][number]

const props = defineProps<{
  achievement: ProjectCardAchievement
}>()

const { track } = useAnalytics()
const { openAchievementId, clearOpenAchievementId } = useAchievementSheet()

const open = ref(false)

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
  } else if (openAchievementId.value === props.achievement.id) {
    clearOpenAchievementId()
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

function descriptionFor(achievement: ProjectCardAchievement) {
  if (achievement.achievedAt) {
    return achievement.descriptionCompleted
  } else {
    return achievement.descriptionPending
  }
}
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
        <div class="flex h-full flex-col items-center justify-center gap-6">
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
          <div
            class="flex flex-col items-center gap-1 text-center text-balance"
          >
            <h3 class="text-heading" v-html="achievement.name" />
            <p class="text-label" v-html="descriptionFor(achievement)" />
          </div>
          <div
            v-if="achievement.achievedAt"
            class="rounded-full bg-background-indent py-2 px-3 text-label text-accent-contrast"
          >
            +{{ formatNumber(achievement.points ?? 0) }} {{ $t('points') }}
          </div>
          <div
            v-else-if="achievement.points"
            class="rounded-full bg-background-indent py-2 px-3 text-label text-text-muted"
          >
            {{
              $t('givesYouXPoints', {
                points: formatNumber(achievement.points),
              })
            }}
          </div>

          <!-- <template v-if="achievement.__typename === 'ContentAchievement'">
            <div
              v-if="achievement.nextItem"
              class="w-full mt-auto flex flex-col p-default text-center"
            >
              <p class="text-label text-text-default">
                {{ achievement.nextItem.externalContent.title }}
              </p>
              <NuxtLink
                v-if="achievement.nextItem.externalContent.url"
                :to="achievement.nextItem.externalContent.url"
                class="contents"
              >
                <DesignButton size="large" variant="primary">
                  {{ $t('achievement.nextStep') }}
                </DesignButton>
              </NuxtLink>
            </div>
          </template> -->
        </div>
      </template>
    </DesignDrawer>
  </div>
</template>
