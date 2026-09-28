<script setup lang="ts">
import type { ChallengePageQuery } from '~/api/generated'

type SimpleChallengeData = Extract<
  ChallengePageQuery['challenge'],
  { __typename: 'SimpleChallenge' }
>

const props = defineProps<{
  challenge: SimpleChallengeData
}>()

const { track } = useAnalytics()

onMounted(() => {
  track(AnalyticsEvent.ChallengeOpened, {
    challenge_id: props.challenge.id,
    challenge_name: props.challenge.name,
    challenge_type: 'simple',
  })
})
</script>

<template>
  <PageLayout>
    <template #action>
      <NuxtLink :to="{ name: 'challenges' }">
        <DesignIconButton icon="IconClose" />
      </NuxtLink>
    </template>
  </PageLayout>
</template>
