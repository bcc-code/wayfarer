<script setup lang="ts">
const route = useRoute('challenges-challengeId')

const { isAuthReady } = useAuthReady()
const {
  data,
  fetching,
  error,
  executeQuery: refresh,
} = useChallengePageQuery({
  variables: {
    challengeId: route.params.challengeId,
  },
  pause: computed(() => !isAuthReady.value),
  requestPolicy: 'network-only', // prevents some race conditions
})

// Listen for Firestore realtime updates
useFirestoreRefresh(['ChallengePageDocument'], () => {
  refresh({ requestPolicy: 'network-only' })
})

const isInitialLoading = computed(() => fetching.value && !data.value)
</script>

<template>
  <div class="h-full">
    <LoadingState v-if="isInitialLoading" />
    <ErrorState v-else-if="error" :error />
    <template v-else-if="data">
      <SimpleChallenge
        v-if="data.challenge.__typename === 'SimpleChallenge'"
        :challenge="data.challenge"
      />
      <QuizChallenge
        v-if="data.challenge.__typename === 'QuizChallenge'"
        :challenge="data.challenge"
        :user-score="data.myCurrentProject?.leaderboard.me?.score ?? 0"
      />
      <PluginChallenge
        v-if="data.challenge.__typename === 'PluginChallenge'"
        :challenge="data.challenge"
      />
    </template>
  </div>
</template>
