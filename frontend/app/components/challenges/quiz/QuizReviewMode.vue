<script setup lang="ts">
import type {
  PredefinedQuestionData,
  PredefinedResponseData,
  OrderingQuestionData,
  OrderingResponseData,
} from './types'

const props = defineProps<{
  questions: (PredefinedQuestionData | OrderingQuestionData)[]
  responses: (PredefinedResponseData | OrderingResponseData)[]
  revealCorrectAnswers: boolean
}>()

const emit = defineEmits<{
  finish: []
}>()

const currentQuestionIndex = ref(0)

// Build a map from questionId to response for O(1) lookup
const responseMap = computed(() => {
  const map = new Map<string, PredefinedResponseData | OrderingResponseData>()
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
  const response = currentResponse.value
  if (!response || response.__typename !== 'PredefinedResponse') return []
  return response.selectedAnswers.map((a) => a.id)
})

const preSubmittedOrder = computed(() => {
  const response = currentResponse.value
  if (!response || response.__typename !== 'OrderingResponse') return []
  return response.submittedOrder
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
    v-if="currentQuestion && currentQuestion.__typename === 'PredefinedQuestion'"
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
  <QuizOrderingQuestion
    v-else-if="currentQuestion && currentQuestion.__typename === 'OrderingQuestion'"
    :key="currentQuestion.id"
    :question="currentQuestion"
    :total-questions="questions.length"
    :current-index="currentQuestionIndex"
    submission-id=""
    :readonly="true"
    :pre-submitted-order="preSubmittedOrder"
    :show-correct-answers="revealCorrectAnswers"
    :show-previous-button="!isFirstQuestion"
    :is-last-question="isLastQuestion"
    @previous="handlePrevious"
    @next="handleNext"
  />
</template>
