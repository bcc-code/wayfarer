<script setup lang="ts">
import {
  formatPercent,
  percentageOfResponses,
} from '../../../utils/quizResultsExport'

interface Option {
  answer: { id: string; answerText: string }
  count: number
  percentage: number
  isCorrect: boolean
}

defineProps<{
  options: Option[]
  responseCount: number
  correctCount: number
}>()
</script>

<template>
  <div class="space-y-3">
    <p v-if="responseCount > 0" class="text-muted text-sm">
      {{ correctCount }} av {{ responseCount }} svarte helt riktig
      <span class="text-dimmed">·</span>
      {{ formatPercent(percentageOfResponses(correctCount, responseCount)) }}
    </p>
    <AdminQuizResultBar
      v-for="option in options"
      :key="option.answer.id"
      :label="option.answer.answerText"
      :count="option.count"
      :percentage="option.percentage"
      :is-correct="option.isCorrect"
    />
  </div>
</template>
