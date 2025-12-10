<script setup lang="ts">
import type { ChallengePageQuery } from '~/api/generated'

type QuizChallengeData = Extract<
  ChallengePageQuery['challenge'],
  { __typename: 'QuizChallenge' }
>

defineProps<{
  challenge: QuizChallengeData
}>()
</script>

<template>
  <PageLayout>
    <template #action>
      <NuxtLink :to="{ name: 'challenges' }">
        <DesignIconButton icon="lucide:x" />
      </NuxtLink>
    </template>
    <template #title>
      <QuizProgress />
    </template>

    <div class="gap-medium flex flex-col">
      <QuizAlternative text="This is an alternative" />
      <QuizAlternative text="This is a highlighted alternative" highlighted />
      <QuizAlternative
        text="This is a confirmed and wrong alternative"
        confirmed
        wrong
      />
      <QuizAlternative
        text="This is a confirmed and correct alternative"
        confirmed
        correct
      />
    </div>
  </PageLayout>
</template>
