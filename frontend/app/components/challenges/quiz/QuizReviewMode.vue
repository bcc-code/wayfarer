<script setup lang="ts">
import type {
  PredefinedQuestionData,
  PredefinedResponseData,
  OrderingQuestionData,
  OrderingResponseData,
  QuizActionState,
  QuizActionHandlers,
  QuizQuestionExposed,
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

// Template ref for the current question component
const currentQuestionRef = ref<QuizQuestionExposed | null>(null)

// Forward the action state from the current question component
const actionState = computed<QuizActionState | undefined>(() => {
  return currentQuestionRef.value?.actionState
})

// Forward the handlers from the current question component
// These must be methods that delegate to the current question's handlers
const handlers: QuizActionHandlers = {
  submit: async () => {
    await currentQuestionRef.value?.handlers.submit()
  },
  continue: () => {
    currentQuestionRef.value?.handlers.continue()
  },
  changeBet: () => {
    currentQuestionRef.value?.handlers.changeBet()
  },
  previous: () => {
    currentQuestionRef.value?.handlers.previous()
  },
  next: () => {
    currentQuestionRef.value?.handlers.next()
  },
}

defineExpose({ actionState, handlers })
</script>

<template>
  <div>
    <div
      class="flex flex-col items-center justify-center py-6 px-default gap-1 text-center"
    >
      <p v-if="questions.length > 1" class="text-caption text-text-muted">
        {{
          $t('quiz.questionNumber', {
            current: currentQuestionIndex + 1,
            total: questions.length,
          })
        }}
      </p>
      <h1 class="text-heading text-text-default text-balance text-center">
        {{ currentQuestion?.questionText }}
      </h1>
    </div>
    <QuizPredefinedQuestion
      v-if="
        currentQuestion && currentQuestion.__typename === 'PredefinedQuestion'
      "
      ref="currentQuestionRef"
      :key="`predefined:${currentQuestion.id}`"
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
      v-else-if="
        currentQuestion && currentQuestion.__typename === 'OrderingQuestion'
      "
      ref="currentQuestionRef"
      :key="`ordering:${currentQuestion.id}`"
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
  </div>
</template>
