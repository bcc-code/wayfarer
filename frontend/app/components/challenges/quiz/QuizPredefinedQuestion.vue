<script setup lang="ts">
import type { PredefinedQuestionData, QuestionResult } from './types'

const props = defineProps<{
  question: PredefinedQuestionData
  totalQuestions: number
  currentIndex: number
  submissionId: string
}>()

const emit = defineEmits<{
  answerSubmitted: [result: QuestionResult]
}>()

const { track } = useAnalytics()
const { executeMutation: submitAnswer } = useSubmitQuizAnswerMutation()

const selectedAnswer = ref<string>()
const isAnswerConfirmed = ref(false)
const isSubmitting = ref(false)
const submittedResult = ref<{ isCorrect: boolean | null } | null>(null)

async function handleLockAnswer() {
  if (!selectedAnswer.value || isSubmitting.value) return

  isSubmitting.value = true

  const result = await submitAnswer({
    submissionId: props.submissionId,
    input: {
      questionId: props.question.id,
      selectedAnswerIds: [selectedAnswer.value],
    },
  })

  if (result.data?.submitQuizAnswer) {
    const response = result.data.submitQuizAnswer
    const isCorrect =
      response.__typename === 'PredefinedResponse'
        ? (response.isCorrect ?? null)
        : null

    submittedResult.value = { isCorrect }
    isAnswerConfirmed.value = true

    track(AnalyticsEvent.QuizAnswerSubmitted, {
      question_id: props.question.id,
      is_correct: isCorrect,
      current_question: props.currentIndex + 1,
      total_questions: props.totalQuestions,
    })
  }

  isSubmitting.value = false
}

function handleContinue() {
  emit('answerSubmitted', {
    questionId: props.question.id,
    isCorrect: submittedResult.value?.isCorrect ?? null,
  })
}

const { t } = useI18n()
const continueText = computed(() => {
  if (props.currentIndex === props.totalQuestions - 1) {
    return t('quiz.continue')
  }
  return t('quiz.nextQuestion')
})
</script>

<template>
  <div class="text-center p-default flex flex-col gap-default grow">
    <div class="grow flex flex-col items-center justify-center py-15 gap-1">
      <p class="text-caption text-text-muted">
        {{
          $t('quiz.questionNumber', {
            current: currentIndex + 1,
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
        :correct="isAnswerConfirmed && alternative.isCorrect === true"
        :wrong="
          isAnswerConfirmed &&
          selectedAnswer === alternative.id &&
          alternative.isCorrect === false
        "
        :selected="selectedAnswer === alternative.id"
        :disabled="isAnswerConfirmed"
        @click="!isAnswerConfirmed && (selectedAnswer = alternative.id)"
      />
    </template>

    <DesignButton
      v-if="!isAnswerConfirmed"
      size="large"
      class="grow-0"
      :disabled="selectedAnswer === undefined || isSubmitting"
      :loading="isSubmitting"
      @click="handleLockAnswer"
    >
      {{ $t('quiz.lockAnswer') }}
    </DesignButton>

    <DesignButton v-else size="large" class="grow-0" @click="handleContinue">
      {{ continueText }}
    </DesignButton>
  </div>
</template>
