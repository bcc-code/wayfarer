<script setup lang="ts">
import type { PredefinedQuestionData } from './types'

defineProps<{
  question: PredefinedQuestionData
  totalQuestions: number
}>()

const selectedAnswer = ref<string>()
const isAnswerConfirmed = ref(false)
</script>

<template>
  <div class="text-center p-default flex flex-col gap-default grow">
    <div class="grow flex flex-col items-center justify-center py-15 gap-1">
      <p class="text-caption text-text-muted">
        {{
          $t('quiz.questionNumber', {
            current: question.questionOrder,
            total: totalQuestions,
          })
        }}
      </p>
      <h2 class="text-heading text-text-default text-balance">
        {{ question.questionText }}
      </h2>
    </div>

    <template
      v-for="alternative in question.predefinedAnswers"
      :key="alternative.id"
    >
      <QuizAlternative
        :text="alternative.answerText"
        :highlighted="selectedAnswer === alternative.id"
        :confirmed="isAnswerConfirmed"
        :correct="alternative.isCorrect === true"
        :wrong="alternative.isCorrect === false"
        @click="selectedAnswer = alternative.id"
      />
    </template>

    <DesignButton
      size="large"
      class="grow-0"
      :disabled="selectedAnswer === undefined"
    >
      {{ $t('quiz.lockAnswer') }}
    </DesignButton>
  </div>
</template>
