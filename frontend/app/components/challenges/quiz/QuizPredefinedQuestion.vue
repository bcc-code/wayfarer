<script setup lang="ts">
import type { PredefinedQuestionData, QuestionResult } from './types'

const props = defineProps<{
  question: PredefinedQuestionData
  totalQuestions: number
  currentIndex: number
  submissionId: string
  // Review mode props
  readonly?: boolean
  preSelectedAnswerIds?: string[]
  showCorrectAnswers?: boolean
  showPreviousButton?: boolean
  isLastQuestion?: boolean
}>()

const emit = defineEmits<{
  answerSubmitted: [result: QuestionResult]
  previous: []
  next: []
}>()

const { track } = useAnalytics()
const { executeMutation: submitAnswer } = useSubmitQuizAnswerMutation()

// In readonly mode, pre-populate selected answer from props
const selectedAnswer = ref<string | undefined>(props.preSelectedAnswerIds?.[0])
const isAnswerConfirmed = ref(props.readonly ?? false)
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

// In readonly mode, we use isLastQuestion prop to determine the next button text
const nextButtonText = computed(() => {
  if (props.readonly) {
    return props.isLastQuestion
      ? t('quiz.finishReview')
      : t('quiz.nextQuestion')
  }
  return continueText.value
})

function handlePrevious() {
  emit('previous')
}

function handleNext() {
  emit('next')
}

// Determine if we should show correct/wrong highlighting
// In normal mode: always show after confirmation
// In readonly mode: only show if showCorrectAnswers is true
function shouldShowCorrect(alternative: { isCorrect: boolean | null }) {
  if (!isAnswerConfirmed.value) return false
  if (props.readonly && !props.showCorrectAnswers) return false
  return alternative.isCorrect === true
}

function shouldShowWrong(alternative: {
  id: string
  isCorrect: boolean | null
}) {
  if (!isAnswerConfirmed.value) return false
  if (props.readonly && !props.showCorrectAnswers) return false
  return (
    selectedAnswer.value === alternative.id && alternative.isCorrect === false
  )
}
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
        :correct="shouldShowCorrect(alternative)"
        :wrong="shouldShowWrong(alternative)"
        :selected="selectedAnswer === alternative.id"
        :disabled="isAnswerConfirmed"
        @click="!isAnswerConfirmed && (selectedAnswer = alternative.id)"
      />
    </template>

    <!-- Normal mode: Lock answer / Continue buttons -->
    <template v-if="!readonly">
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
    </template>

    <!-- Readonly/review mode: Previous / Next navigation -->
    <template v-else>
      <div class="flex gap-small grow-0">
        <DesignButton
          v-if="showPreviousButton"
          size="large"
          variant="secondary"
          class="flex-1"
          @click="handlePrevious"
        >
          {{ $t('quiz.previousQuestion') }}
        </DesignButton>
        <DesignButton
          size="large"
          :class="showPreviousButton ? 'flex-1' : 'w-full'"
          @click="handleNext"
        >
          {{ nextButtonText }}
        </DesignButton>
      </div>
    </template>
  </div>
</template>
