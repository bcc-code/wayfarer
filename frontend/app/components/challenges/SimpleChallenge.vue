<script setup lang="ts">
import type { ChallengePageQuery } from '~/api/generated'

type SimpleChallengeData = Extract<
  ChallengePageQuery['challenge'],
  { __typename: 'SimpleChallenge' }
>

defineProps<{
  challenge: SimpleChallengeData
}>()
</script>

<template>
  <PageLayout>
    <template #action>
      <NuxtLink :to="{ name: 'challenges' }">
        <DesignIconButton icon="lucide:x" />
      </NuxtLink>
    </template>
  </PageLayout>
</template>
