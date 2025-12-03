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
    <UModal
      v-model:open="open"
      :ui="{ content: 'bg-background-default' }"
      :transition="false"
      modal
      fullscreen
    >
      <button
        class="grid aspect-square place-items-center overflow-hidden rounded-full"
      >
        <NuxtImg
          v-if="achievement.image && achievement.achievedAt != null"
          :src="achievement.image"
          class="size-full object-cover"
        />
        <NuxtImg
          v-else
          src="/images/achievement-placeholder.png"
          class="size-full object-cover"
        />
      </button>
      <template #content="{ close }">
        <PageLayout
          :title="achievement.achievedAt ? 'Unlocked!' : 'Achievement'"
        >
          <template #action>
            <DesignIconButton icon="lucide:x" @click="close" />
          </template>

          <div class="flex h-full flex-col items-center justify-center gap-6">
            <div
              :class="[
                'grid aspect-square size-55 place-items-center overflow-hidden rounded-full',
                { 'shadow-large': achievement.achievedAt },
              ]"
            >
              <NuxtImg
                v-if="achievement.image && achievement.achievedAt != null"
                :src="achievement.image"
                class="size-full object-cover"
              />
              <NuxtImg
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
                {{ achievement.description }}
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
        </PageLayout>
      </template>
    </UModal>
  </div>
</template>
