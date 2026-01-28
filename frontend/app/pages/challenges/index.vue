<script setup lang="ts">
const { isAuthReady } = useAuthReady()

const {
  data,
  fetching,
  error,
  executeQuery: refresh,
} = useChallengesPageQuery({
  pause: computed(() => !isAuthReady.value),
})

// Listen for Firestore realtime updates
useFirestoreRefresh(['ChallengesPageDocument'], () => {
  refresh({ requestPolicy: 'network-only' })
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

const joinCode = computed(() =>
  data.value?.myCurrentProject.myTeam?.joinCode.split(''),
)
</script>

<template>
  <PageLayout :title="$t('pages.challenges')">
    <LoadingState v-if="isInitialLoading" />
    <ErrorState v-else-if="error" :error />
    <div
      v-else-if="visibleChallenges.length"
      class="space-y-list-section-gap p-list-outside"
    >
      <template v-for="challenge in visibleChallenges" :key="challenge.id">
        <!-- This is very specific for the Ladder to Heaven project, and should be more generic later on -->
        <div
          v-if="challenge.__typename === 'PluginChallenge'"
          class="bg-accent text-on-accent rounded-card p-7 flex flex-col gap-default items-center"
        >
          <div
            class="text-center flex flex-col items-center gap-small py-medium"
          >
            <h3 class="text-heading">pc26.bcc.media</h3>
            <p class="text-label">
              {{ $t('gameNights.yourCodeHint') }}
            </p>
          </div>
          <p class="text-caption">
            {{ $t('gameNights.yourCode') }}
          </p>
          <div v-if="joinCode" class="grid grid-cols-6 gap-list-section-inset">
            <div
              v-for="(char, index) in joinCode"
              :key="index"
              class="text-heading p-medium aspect-[1/1.3] flex items-center justify-center border-3 border-on-accent rounded-list-inset text-center"
            >
              {{ char }}
            </div>
          </div>
        </div>
        <ChallengeCard v-else :challenge class="challenge-card" />
      </template>
    </div>
    <EmptyState v-else :title="$t('emptyStates.challenges')" />
  </PageLayout>
</template>
