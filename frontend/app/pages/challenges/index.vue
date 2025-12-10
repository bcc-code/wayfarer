<script setup lang="ts">
const { isAuthReady } = useAuthReady()

const { data, fetching, error } = useChallengesPageQuery({
  pause: computed(() => !isAuthReady.value),
})
</script>

<template>
  <PageLayout :title="$t('pages.challenges')">
    <LoadingState v-if="fetching" />
    <ErrorState v-else-if="error" :error />
    <div
      v-else-if="data?.myCurrentProject.challenges.length"
      class="space-y-list-section-gap"
    >
      <ChallengeCard
        v-for="challenge in data.myCurrentProject.challenges"
        :key="challenge.id"
        :challenge
      />
    </div>
    <EmptyState v-else />
  </PageLayout>
</template>
