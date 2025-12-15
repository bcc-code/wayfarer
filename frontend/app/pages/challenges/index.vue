<script setup lang="ts">
const { isAuthReady } = useAuthReady()

const { data, fetching, error } = useChallengesPageQuery({
  pause: computed(() => !isAuthReady.value),
})

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
</script>

<template>
  <PageLayout :title="$t('pages.challenges')">
    <LoadingState v-if="fetching" />
    <ErrorState v-else-if="error" :error />
    <div
      v-else-if="visibleChallenges.length"
      class="space-y-list-section-gap p-list-outside"
    >
      <ChallengeCard
        v-for="challenge in visibleChallenges"
        :key="challenge.id"
        :challenge
      />
    </div>
    <EmptyState v-else />
  </PageLayout>
</template>
