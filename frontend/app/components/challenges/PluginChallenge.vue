<script setup lang="ts">
import type { ChallengePageQuery } from '~/api/generated'

type PluginChallengeData = Extract<
  ChallengePageQuery['challenge'],
  { __typename: 'PluginChallenge' }
>

const props = defineProps<{
  challenge: PluginChallengeData
}>()

const { track } = useAnalytics()

onMounted(() => {
  track(AnalyticsEvent.ChallengeOpened, {
    challenge_id: props.challenge.id,
    challenge_name: props.challenge.name,
    challenge_type: 'plugin',
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
