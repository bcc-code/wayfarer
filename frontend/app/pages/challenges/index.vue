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

const { data, fetching, error } = useChallengesPageQuery()
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
  <PageLayout title="Challenges">
    <LoadingState v-if="fetching" />
    <ErrorState v-else-if="error" :error />
    <div v-else-if="data" class="space-y-4">
      <ChallengeCard
        v-for="challenge in relevantChallenges"
        :key="challenge.id"
        :challenge
      />
    </div>
  </PageLayout>
</template>
