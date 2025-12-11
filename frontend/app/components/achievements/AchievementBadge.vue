<script setup lang="ts">
const props = defineProps<{
  achievement: Partial<Achievement>
}>()

const { track } = useAnalytics()

const open = ref(false)

watch(open, (isOpen) => {
  if (isOpen) {
    track(AnalyticsEvent.AchievementClicked, {
      achievement_id: props.achievement.id,
      achievement_name: props.achievement.name,
      is_unlocked: !!props.achievement.achievedAt,
    })
  }
})
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
        class="grid aspect-square place-items-center overflow-hidden rounded-full"
      >
        <img
          v-if="achievement.imageCompleted && achievement.achievedAt != null"
          :src="achievement.imageCompleted"
          class="size-full object-cover"
        />
        <img
          v-else-if="achievement.imagePending"
          :src="achievement.imagePending"
          class="size-full object-cover"
        />
        <img
          v-else
          src="/images/achievement-placeholder.png"
          class="size-full object-cover"
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
            <img
              v-if="
                achievement.imageCompleted && achievement.achievedAt != null
              "
              :src="achievement.imageCompleted"
              class="size-full object-cover"
            />
            <img
              v-else-if="achievement.imagePending"
              :src="achievement.imagePending"
              class="size-full object-cover"
            />
            <img
              v-else
              src="/images/achievement-placeholder.png"
              class="size-full object-cover"
            />
          </div>
          <div
            class="flex flex-col items-center gap-1 text-center text-balance"
          >
            <h3 class="text-heading">
              {{ achievement.name }}
            </h3>
            <p class="text-label">
              {{
                achievement.achievedAt
                  ? achievement.descriptionCompleted
                  : achievement.descriptionPending
              }}
            </p>
          </div>
          <div
            v-if="achievement.achievedAt"
            class="rounded-full bg-background-indent py-2 px-3 text-label text-accent-contrast"
          >
            +{{ achievement.points }} {{ $t('points') }}
          </div>
          <div
            v-else-if="achievement.points"
            class="rounded-full bg-background-indent py-2 px-3 text-label text-text-muted"
          >
            {{ $t('givesYouXPoints', { points: achievement.points }) }}
          </div>
        </div>
      </template>
    </DesignDrawer>
  </div>
</template>
