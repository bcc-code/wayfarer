<script setup lang="ts">
gql(`
  query ChallengePage($challengeId: ID!) {
    challenge(id: $challengeId) {
      id
      name
      description
      image
      url
      buttonText
      publishedAt
      endTime
      userCompletedAt
    }
  }
`)

const route = useRoute('challenges-challengeId')

const { isAuthReady } = useAuthReady()
const { data, fetching, error } = useChallengePageQuery({
  variables: {
    challengeId: route.params.challengeId,
  },
  pause: computed(() => !isAuthReady.value),
})
</script>

<template>
  <PageLayout :title="$t('pages.challenge')">
    <template #action>
      <NuxtLink :to="{ name: 'challenges' }">
        <DesignIconButton icon="lucide:x" />
      </NuxtLink>
    </template>

    <LoadingState v-if="fetching" />
    <ErrorState v-else-if="error" :error />
    <pre v-else-if="data">{{ data.challenge }}</pre>
  </PageLayout>
</template>
