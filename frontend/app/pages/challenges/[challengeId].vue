<script setup lang="ts">
const route = useRoute('challenges-challengeId')

const { isAuthReady } = useAuthReady()
const { data, fetching, error } = useChallengePageQuery({
  variables: {
    challengeId: route.params.challengeId,
  },
  pause: computed(() => !isAuthReady.value),
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
      />
    </template>
  </div>
</template>
