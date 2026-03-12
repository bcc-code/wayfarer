<script setup lang="ts">
import { QuizSessionState } from '~/api/generated'

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
        (s) => s.completedAt,
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
        (s) => s.completedAt,
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
        <ChallengeCard
          v-if="challenge.__typename === 'PluginChallenge'"
          :challenge
        >
          <template #content>
            <div class="px-6 pb-6 pt-4 space-y-4">
              <div class="py-4 space-y-small">
                <p class="flex gap-small items-center text-label">
                  <span
                    class="rounded-full flex items-center justify-center size-6 bg-accent text-on-accent text-center"
                  >
                    1
                  </span>
                  <span class="text-text-default">
                    {{ $t('gameNights.yourCodeHint') }}
                  </span>
                </p>
                <p
                  class="text-title rounded-full bg-background-indent px-default py-1"
                >
                  <span class="text-text-hint">https://</span>
                  <span class="text-text-default">pc26.bcc.media</span>
                </p>
              </div>
              <div class="space-y-3">
                <p class="flex gap-small items-center text-label">
                  <span
                    class="rounded-full flex items-center justify-center size-6 bg-accent text-on-accent text-center"
                  >
                    2
                  </span>
                  <span class="text-text-default">
                    {{ $t('gameNights.yourCode') }}
                  </span>
                </p>
                <div
                  v-if="joinCode"
                  class="grid grid-cols-6 gap-list-section-inset"
                >
                  <div
                    v-for="(char, index) in joinCode"
                    :key="index"
                    class="text-heading h-16.75 px-medium flex items-center justify-center border-3 rounded-list-inset text-center"
                  >
                    {{ char }}
                  </div>
                </div>
              </div>
            </div>
          </template>
        </ChallengeCard>
        <ChallengeCard v-else :challenge />
      </template>
      <EmptyState
        v-if="!tabChallenges?.length"
        :title="$t('emptyStates.challenges')"
      />
    </div>
  </PageLayout>
</template>
