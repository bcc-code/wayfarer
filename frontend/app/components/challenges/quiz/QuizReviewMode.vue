<script setup lang="ts">
import type { PredefinedQuestionData, PredefinedResponseData } from './types'

const props = defineProps<{
  questions: PredefinedQuestionData[]
  responses: PredefinedResponseData[]
  revealCorrectAnswers: boolean
}>()

const emit = defineEmits<{
  finish: []
}>()

const currentQuestionIndex = ref(0)

// Build a map from questionId to response for O(1) lookup
const responseMap = computed(() => {
  const map = new Map<string, PredefinedResponseData>()
  for (const response of props.responses) {
    map.set(response.question.id, response)
  }
  return map
})

const currentQuestion = computed(() => {
  return props.questions[currentQuestionIndex.value]
})

const currentResponse = computed(() => {
  if (!currentQuestion.value) return undefined
  return responseMap.value.get(currentQuestion.value.id)
})

const preSelectedAnswerIds = computed(() => {
  return currentResponse.value?.selectedAnswers.map((a) => a.id) ?? []
})

const isFirstQuestion = computed(() => currentQuestionIndex.value === 0)
const isLastQuestion = computed(
  () => currentQuestionIndex.value === props.questions.length - 1,
)

function handlePrevious() {
  if (!isFirstQuestion.value) {
    currentQuestionIndex.value--
  }
}

function handleNext() {
  if (isLastQuestion.value) {
    emit('finish')
  } else {
    currentQuestionIndex.value++
  }
}
</script>

<template>
  <QuizPredefinedQuestion
    v-if="currentQuestion"
    :key="currentQuestion.id"
    :question="currentQuestion"
    :total-questions="questions.length"
    :current-index="currentQuestionIndex"
    submission-id=""
    :readonly="true"
    :pre-selected-answer-ids="preSelectedAnswerIds"
    :show-correct-answers="revealCorrectAnswers"
    :show-previous-button="!isFirstQuestion"
    :is-last-question="isLastQuestion"
    @previous="handlePrevious"
    @next="handleNext"
  />
</template>
