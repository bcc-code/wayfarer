<script setup lang="ts">
import type { QuestionResult } from './types'

const props = defineProps<{
  score: number
  maxScore: number
  pointsAwarded: number
  results: QuestionResult[]
}>()

const { t } = useI18n()

const resultText = computed(() => {
  if (props.score === props.maxScore) {
    return t('quiz.result.perfect')
  }
  if (props.score > 0) {
    return t('quiz.result.wellDone')
  }
  return t('quiz.result.betterLuckNextTime')
})
</script>

<template>
  <div class="text-center p-default flex flex-col gap-large grow">
    <div class="grow flex flex-col items-center justify-center gap-default">
      <QuizProgress
        size="large"
        :total-questions="maxScore"
        :current-index="maxScore"
        :results
      />

      <h1 class="text-heading text-text-default">
        {{ resultText }}
        <br />
        {{
          score === 0
            ? $t('quiz.result.receivedNoPoints')
            : t('quiz.result.receivedPoints', { points: pointsAwarded })
        }}
      </h1>
    </div>

    <NuxtLink :to="{ name: 'challenges' }">
      <DesignButton size="large" class="w-full">
        {{ $t('quiz.done') }}
      </DesignButton>
    </NuxtLink>
  </div>
</template>
