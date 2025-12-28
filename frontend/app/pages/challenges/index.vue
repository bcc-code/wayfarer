<script setup lang="ts">
const { isAuthReady } = useAuthReady()

const { data, fetching, error } = useChallengesPageQuery({
  pause: computed(() => !isAuthReady.value),
})

const isInitialLoading = computed(() => fetching.value && !data.value)

// Filter out completed quiz challenges that can't be retaken
const visibleChallenges = computed(() => {
  if (!data.value?.myCurrentProject.challenges) return []

  return data.value.myCurrentProject.challenges.filter((challenge) => {
    // Hide completed quiz challenges that don't allow retakes
    if (challenge.__typename === 'QuizChallenge') {
      const isCompleted = !!challenge.userCompletedAt
      const canStart = challenge.quiz.userCanStart
      // Hide if completed and can't start again
      if (isCompleted && !canStart) {
        return false
      }
    }
    return true
  })
})

// Staggered entrance animation for challenge cards
const cardsContainer = ref<HTMLElement | null>(null)
const { animate } = useStaggeredEntrance()
const hasAnimated = ref(false)

watch(
  () => visibleChallenges.value,
  (challenges) => {
    if (hasAnimated.value) return
    if (challenges.length > 0 && cardsContainer.value) {
      hasAnimated.value = true
      nextTick(() => {
        const cards = cardsContainer.value?.querySelectorAll('.challenge-card')
        if (cards) {
          animate(cards)
        }
      })
    }
  },
)
</script>

<template>
  <PageLayout :title="$t('pages.challenges')">
    <LoadingState v-if="isInitialLoading" />
    <ErrorState v-else-if="error" :error />
    <div
      v-else-if="visibleChallenges.length"
      ref="cardsContainer"
      class="space-y-list-section-gap p-list-outside"
    >
      <ChallengeCard
        v-for="challenge in visibleChallenges"
        :key="challenge.id"
        :challenge
        class="challenge-card"
      />
    </div>
    <EmptyState v-else :title="$t('emptyStates.challenges')" />
  </PageLayout>
</template>
