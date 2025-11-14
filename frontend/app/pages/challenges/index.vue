<script setup lang="ts">
gql(`
  query ChallengesPage {
    myCurrentProject {
      challenges {
        id
        name
        description
        userCompletedAt
        image
        url
        buttonText
        publishedAt
        endTime
      }
    }
  }
`)

const { isAuthReady } = useAuthReady()

const { data, fetching, error } = useChallengesPageQuery({
  pause: computed(() => !isAuthReady.value),
})
const relevantChallenges = computed(() => {
  const notCompleted = data.value?.myCurrentProject.challenges.filter(
    (challenge) => !challenge.userCompletedAt,
  )
  const notEnded = notCompleted?.filter(
    (challenge) =>
      !challenge.endTime || new Date(challenge.endTime).getTime() > Date.now(),
  )
  return notEnded
})
</script>

<template>
  <PageLayout :title="$t('pages.challenges')">
    <LoadingState v-if="fetching" />
    <ErrorState v-else-if="error" :error />
    <div
      v-else-if="relevantChallenges?.length"
      class="space-y-list-section-gap"
    >
      <ChallengeCard
        v-for="challenge in relevantChallenges"
        :key="challenge.id"
        :challenge
      />
    </div>
    <EmptyState v-else />
  </PageLayout>
</template>
