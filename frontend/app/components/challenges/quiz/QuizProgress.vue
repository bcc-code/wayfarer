<script setup lang="ts">
import { cva } from 'cva'

type QuizChallengeData = Extract<
  ChallengePageQuery['challenge'],
  { __typename: 'QuizChallenge' }
>

type QuizSubmissionData = QuizChallengeData['quiz']['userSubmissions'][number]

defineProps<{
  submission?: QuizSubmissionData
}>()

const dotClass = cva('', {
  variants: {
    active: {
      true: 'bg-accent-contrast',
      false: 'bg-border-default',
    },
  },
})
</script>

<template>
  <div
    v-if="submission"
    class="bg-background-raised gradient-border px-5 py-2 rounded-button-medium flex items-center gap-medium"
  >
    <div
      v-for="question in submission.orderedQuestions"
      :key="question.id"
      :class="dotClass({ active: true })"
      class="w-2 h-2 rounded-full"
    />
  </div>
</template>
