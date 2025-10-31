<script setup lang="ts">
import { useChallengesPageQuery } from '~/api/generated'

gql(`
query ChallengesPage {
  user {
    currentProject {
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
}
`)

const { data, fetching, error } = useChallengesPageQuery()
const relevantChallenges = computed(() => {
  const notCompleted = data.value?.user.currentProject.challenges.filter(
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
  <div class="h-full">
    <h1 class="text-xl font-bold mb-4">Challenges</h1>

    <LoadingState v-if="fetching" />
    <ErrorState v-else-if="error" :error />
    <div v-else-if="data" class="space-y-4">
      <ChallengeCard
        v-for="challenge in relevantChallenges"
        :key="challenge.id"
        :challenge
      />
    </div>
  </div>
</template>
