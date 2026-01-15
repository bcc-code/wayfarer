<script setup lang="ts">
const { isAuthReady } = useAuthReady()

const { data, fetching, error } = useChallengesPageQuery({
  pause: computed(() => !isAuthReady.value),
})

const isInitialLoading = computed(() => fetching.value && !data.value)

// Filter out quiz challenges without active sessions
const visibleChallenges = computed(() => {
  if (!data.value?.myCurrentProject.challenges) return []

  return data.value.myCurrentProject.challenges.filter((challenge) => {
    if (challenge.__typename === 'QuizChallenge') {
      // Hide quiz challenges without an active session
      if (!challenge.quiz.userActiveSession?.id) {
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
