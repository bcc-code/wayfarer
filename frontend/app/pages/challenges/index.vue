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

const activeChallenges = computed(() =>
  data.value?.myCurrentProject.challenges.filter((challenge) => {
    const challengePastEndTime =
      challenge.endTime && new Date(challenge.endTime).getTime() < Date.now()

    if (challenge.__typename === 'QuizChallenge') {
      const quizPastEndTime =
        challenge.quiz.endTime &&
        new Date(challenge.quiz.endTime).getTime() < Date.now()
      const quizCompleted = challenge.quiz.userSubmissions?.some(
        (s) => s.completedAt && !s.autoSubmitted,
      )

      return !quizCompleted && !challengePastEndTime && !quizPastEndTime
    }

    return !challenge.userCompletedAt && !challengePastEndTime
  }),
)

const completedChallenges = computed(() =>
  data.value?.myCurrentProject.challenges.filter((challenge) => {
    const challengePastEndTime =
      challenge.endTime && new Date(challenge.endTime).getTime() < Date.now()

    if (challenge.__typename === 'QuizChallenge') {
      const quizPastEndTime =
        challenge.quiz.endTime &&
        new Date(challenge.quiz.endTime).getTime() < Date.now()
      const quizCompleted = challenge.quiz.userSubmissions?.some(
        (s) => s.completedAt && !s.autoSubmitted,
      )

      return quizCompleted || challengePastEndTime || quizPastEndTime
    }

    return Boolean(challenge.userCompletedAt) || challengePastEndTime
  }),
)

const joinCode = computed(() =>
  data.value?.myCurrentProject.myTeam?.joinCode.split(''),
)

const tab = ref<'active' | 'completed'>('active')
const tabChallenges = computed(() =>
  tab.value === 'active' ? activeChallenges.value : completedChallenges.value,
)
</script>

<template>
  <PageLayout :title="$t('pages.challenges')">
    <div class="px-list-outside">
      <DesignTabs
        v-model="tab"
        :tabs="[
          { key: 'active', value: 'active', label: $t('challenges.active') },
          {
            key: 'completed',
            value: 'completed',
            label: $t('challenges.completed'),
          },
        ]"
      />
    </div>
    <LoadingState v-if="isInitialLoading" />
    <ErrorState v-else-if="error" :error />
    <div v-else class="space-y-list-section-gap p-list-outside mt-3 grow">
      <template v-for="challenge in tabChallenges" :key="challenge.id">
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
        <ChallengeCard v-else :challenge />
      </template>
      <EmptyState
        v-if="!tabChallenges?.length"
        :title="$t('emptyStates.challenges')"
      />
    </div>
  </PageLayout>
</template>
