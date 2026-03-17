<script setup lang="ts">
const route = useRoute('challenges-challengeId')
const router = useRouter()

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

// Self-enroll via ?enroll=true query param (e.g. from QR code)
const shouldEnroll = ref(route.query.enroll === 'true')
if (shouldEnroll.value) {
  router.replace({ query: { ...route.query, enroll: undefined } })
}

const { executeMutation: enrollInChallenge } = useEnrollInChallengeMutation()

watch(isAuthReady, async (ready) => {
  if (!ready || !shouldEnroll.value) return
  shouldEnroll.value = false

  try {
    await enrollInChallenge({ challengeId: route.params.challengeId })
    refresh({ requestPolicy: 'network-only' })
  } catch {
    // Silently handle — page still loads normally
  }
}, { immediate: true })

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
        :user-score="data.myCurrentProject?.myPoints ?? 0"
      />
      <PluginChallenge
        v-if="data.challenge.__typename === 'PluginChallenge'"
        :challenge="data.challenge"
      />
    </template>
  </div>
</template>
