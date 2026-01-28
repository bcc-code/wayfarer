<script setup lang="ts">
import type {
  QuizChallengeData,
  QuestionResult,
  PredefinedQuestionData,
  PredefinedResponseData,
  OrderingQuestionData,
  OrderingResponseData,
  QuizQuestionExposed,
} from './quiz/types'
import type {
  FinalizeQuizMutation,
  StartQuizSessionMutation,
} from '~/api/generated'

const props = defineProps<{
  challenge: QuizChallengeData
}>()

const emit = defineEmits<{
  start: []
  complete: []
}>()

const { track } = useAnalytics()
const { executeMutation: startQuizSession } = useStartQuizSessionMutation()
const { executeMutation: finalizeQuiz } = useFinalizeQuizMutation()

const currentQuestionIndex = ref(0)
const questionResults = ref<QuestionResult[]>([])
const quizCompleted = ref(false)
const finalResult = ref<FinalizeQuizMutation['finalizeQuiz'] | null>(null)
const startedSubmission = ref<
  StartQuizSessionMutation['startQuizSession'] | null
>(null)
const isReviewMode = ref(false)

// Start with loading true if we need to start a quiz (no active submission and have session)
const needsToStartQuiz = computed(() => {
  return (
    !props.challenge.quiz.userActiveSubmission?.id &&
    props.challenge.quiz.userActiveSession?.id != null
  )
})
const isLoading = ref(needsToStartQuiz.value)

// Check if user has a completed submission (quiz already taken)
const completedSubmission = computed(() => {
  return props.challenge.quiz.userSubmissions.find((s) => s.completedAt)
})

// Build results from a completed submission's responses
const completedSubmissionResults = computed<QuestionResult[]>(() => {
  if (!completedSubmission.value) return []

  return completedSubmission.value.responses.map((response) => ({
    questionId: response.question.id,
    isCorrect:
      response.__typename === 'PredefinedResponse' ||
      response.__typename === 'OrderingResponse'
        ? (response.isCorrect ?? null)
        : null,
  }))
})

// Check if user can review their answers (has PredefinedResponse or OrderingResponse answers)
const canReview = computed(() => {
  if (!completedSubmission.value) return false
  return completedSubmission.value.responses.some(
    (r) =>
      r.__typename === 'PredefinedResponse' ||
      r.__typename === 'OrderingResponse',
  )
})

// Get questions for review mode (PredefinedQuestion and OrderingQuestion types)
const reviewQuestions = computed<
  (PredefinedQuestionData | OrderingQuestionData)[]
>(() => {
  if (!completedSubmission.value) return []
  return completedSubmission.value.orderedQuestions.filter(
    (q): q is PredefinedQuestionData | OrderingQuestionData =>
      q.__typename === 'PredefinedQuestion' ||
      q.__typename === 'OrderingQuestion',
  )
})

// Get responses for review mode (PredefinedResponse and OrderingResponse types)
const reviewResponses = computed<
  (PredefinedResponseData | OrderingResponseData)[]
>(() => {
  if (!completedSubmission.value) return []
  return completedSubmission.value.responses.filter(
    (r): r is PredefinedResponseData | OrderingResponseData =>
      r.__typename === 'PredefinedResponse' ||
      r.__typename === 'OrderingResponse',
  )
})

function handleStartReview() {
  isReviewMode.value = true
}

function handleFinishReview() {
  isReviewMode.value = false
}

// Check if we can start a new quiz
const canStartQuiz = computed(() => {
  return props.challenge.quiz.userCanStart
})

// Check if quiz is unavailable (no active session, or can't start and no submissions to show)
const isQuizUnavailable = computed(() => {
  // No active session means quiz is unavailable
  if (!props.challenge.quiz.userActiveSession?.id) return true
  if (props.challenge.quiz.userCanStart) return false
  if (completedSubmission.value) return false
  if (activeSubmission.value) return false
  return true
})

onMounted(async () => {
  track(AnalyticsEvent.ChallengeOpened, {
    challenge_id: props.challenge.id,
    challenge_name: props.challenge.name,
    challenge_type: 'quiz',
  })

  // If there's no active submission and we have a session, start the quiz
  if (needsToStartQuiz.value && props.challenge.quiz.userActiveSession?.id) {
    const result = await startQuizSession({
      sessionId: props.challenge.quiz.userActiveSession.id,
    })
    if (result.data?.startQuizSession) {
      startedSubmission.value = result.data.startQuizSession
    }
    isLoading.value = false
    emit('start')
    track(AnalyticsEvent.QuizStarted, {
      quiz_id: props.challenge.quiz.id,
      quiz_name: props.challenge.name,
      challenge_id: props.challenge.id,
    })
  }
})

const activeSubmission = computed(() => {
  // Use the started submission if we just started, otherwise find from existing submissions
  if (startedSubmission.value) {
    return startedSubmission.value
  }
  return props.challenge.quiz.userSubmissions.find(
    (submission) =>
      submission.id === props.challenge.quiz.userActiveSubmission?.id,
  )
})

const questions = computed(() => {
  return activeSubmission.value?.orderedQuestions ?? []
})

const currentQuestion = computed(() => {
  return questions.value[currentQuestionIndex.value]
})

const isLastQuestion = computed(() => {
  return currentQuestionIndex.value === questions.value.length - 1
})

// Get session state for ordering questions betting mode
const sessionState = computed(() => {
  return props.challenge.quiz.userActiveSession?.state
})

// Find existing ordering response for the current question
const currentResponse = computed(() => {
  const submission = activeSubmission.value
  if (!submission) return undefined

  if ('responses' in submission) {
    const response = submission.responses.find(
      (r) => r.question.id === currentQuestion.value?.id,
    )

    return response
  }

  return undefined
})

async function handleAnswerSubmitted(result: QuestionResult) {
  questionResults.value.push(result)

  if (isLastQuestion.value) {
    if (activeSubmission.value) {
      const response = await finalizeQuiz({
        submissionId: activeSubmission.value.id,
      })
      if (response.data?.finalizeQuiz) {
        finalResult.value = response.data.finalizeQuiz
        quizCompleted.value = true
        emit('complete')
        track(AnalyticsEvent.QuizCompleted, {
          quiz_id: props.challenge.quiz.id,
          quiz_name: props.challenge.name,
          submission_id: activeSubmission.value.id,
          score: response.data.finalizeQuiz.score ?? 0,
          max_score: response.data.finalizeQuiz.maxScore ?? 0,
          points_awarded: response.data.finalizeQuiz.pointsAwarded ?? 0,
        })
      }
    }
  } else {
    currentQuestionIndex.value++
  }
}

function handleQuizAbandoned() {
  if (activeSubmission.value && !quizCompleted.value) {
    track(AnalyticsEvent.QuizAbandoned, {
      quiz_id: props.challenge.quiz.id,
      quiz_name: props.challenge.name,
      questions_attempted: questionResults.value.length,
      total_questions: questions.value.length,
    })
  }
}

// Template refs for question components
const currentQuestionRef = ref<QuizQuestionExposed | null>(null)
const reviewModeRef = ref<QuizQuestionExposed | null>(null)

// Get the active ref based on current mode
const activeRef = computed(() =>
  isReviewMode.value ? reviewModeRef.value : currentQuestionRef.value,
)

// Get action state and handlers from the active question component
// For question components: actionState is ComputedRef<QuizActionState>, so we need .value
// For QuizReviewMode: actionState is ComputedRef<QuizActionState | undefined>, so we also need .value
const actionState = computed(() => activeRef.value?.actionState)
const handlers = computed(() => activeRef.value?.handlers)

// Determine button text for continue action
const { t } = useI18n()
const continueButtonText = computed(() => {
  if (isLastQuestion.value) {
    return t('quiz.continue')
  }
  return t('quiz.nextQuestion')
})

// Determine button text for next action in review mode
const nextButtonText = computed(() => {
  if (actionState.value?.isLastQuestion) {
    return t('quiz.finishReview')
  }
  return t('quiz.nextQuestion')
})
</script>

<template>
  <PageLayout :bottom-padding="false">
    <template #action>
      <NuxtLink :to="{ name: 'challenges' }" @click="handleQuizAbandoned">
        <DesignIconButton icon="IconClose" />
      </NuxtLink>
    </template>
    <template #title>
      <QuizProgress
        v-if="activeSubmission && !quizCompleted && questions.length > 1"
        :question="currentQuestion"
        :current-index="currentQuestionIndex"
        :total-questions="questions.length"
        :results="questionResults"
      />
    </template>

    <template v-if="isLoading">
      <div class="flex items-center justify-center grow">
        <LoadingState />
      </div>
    </template>

    <template v-else-if="quizCompleted && finalResult">
      <QuizResult
        :score="finalResult.score ?? 0"
        :max-score="finalResult.maxScore ?? 0"
        :points-awarded="finalResult.pointsAwarded ?? 0"
        :results="questionResults"
      />
    </template>

    <template v-else-if="completedSubmission && !canStartQuiz">
      <QuizReviewMode
        v-if="isReviewMode"
        ref="reviewModeRef"
        :questions="reviewQuestions"
        :responses="reviewResponses"
        :reveal-correct-answers="challenge.quiz.revealCorrectAnswers"
        @finish="handleFinishReview"
      />
      <QuizResult
        v-else
        :score="completedSubmission.score ?? 0"
        :max-score="completedSubmission.maxScore ?? 0"
        :points-awarded="completedSubmission.pointsAwarded ?? 0"
        :results="completedSubmissionResults"
        :can-review="canReview"
        @start-review="handleStartReview"
      />
    </template>

    <template v-else-if="currentQuestion">
      <div
        class="flex flex-col items-center justify-center py-6 gap-1 text-center"
      >
        <p v-if="questions.length > 1" class="text-caption text-text-muted">
          {{
            $t('quiz.questionNumber', {
              current: currentQuestionIndex + 1,
              total: questions.length,
            })
          }}
        </p>
        <h1 class="text-heading text-text-default text-balance">
          {{ currentQuestion.questionText }}
        </h1>
      </div>

      <QuizPredefinedQuestion
        v-if="currentQuestion.__typename === 'PredefinedQuestion'"
        ref="currentQuestionRef"
        :key="`predefined:${currentQuestion.id}`"
        :question="currentQuestion"
        :total-questions="questions.length"
        :current-index="currentQuestionIndex"
        :submission-id="activeSubmission?.id ?? ''"
        :is-last-question="isLastQuestion"
        @answer-submitted="handleAnswerSubmitted"
      />
      <QuizOrderingQuestion
        v-else-if="currentQuestion.__typename === 'OrderingQuestion'"
        ref="currentQuestionRef"
        :key="`ordering:${currentQuestion.id}`"
        :question="currentQuestion"
        :total-questions="questions.length"
        :current-index="currentQuestionIndex"
        :submission-id="activeSubmission?.id ?? ''"
        :session-state="sessionState"
        :existing-response="
          currentResponse?.__typename === 'OrderingResponse'
            ? currentResponse
            : undefined
        "
        :is-last-question="isLastQuestion"
        @answer-submitted="handleAnswerSubmitted"
      />
      <QuizNumberQuestion
        v-else-if="currentQuestion.__typename === 'NumberQuestion'"
        :key="`number:${currentQuestion.id}`"
        :question="currentQuestion"
      />
      <QuizJsonQuestion
        v-else-if="currentQuestion.__typename === 'JsonQuestion'"
        :key="`json:${currentQuestion.id}`"
        :question="currentQuestion"
      />
      <QuizFreeTextQuestion
        v-else-if="currentQuestion.__typename === 'FreeTextQuestion'"
        :key="`free-text:${currentQuestion.id}`"
        :question="currentQuestion"
      />
    </template>

    <!-- Fallback: show completed submission result even if retakes are allowed -->
    <template v-else-if="completedSubmission">
      <QuizReviewMode
        v-if="isReviewMode"
        ref="reviewModeRef"
        :questions="reviewQuestions"
        :responses="reviewResponses"
        :reveal-correct-answers="challenge.quiz.revealCorrectAnswers"
        @finish="handleFinishReview"
      />
      <QuizResult
        v-else
        :score="completedSubmission.score ?? 0"
        :max-score="completedSubmission.maxScore ?? 0"
        :points-awarded="completedSubmission.pointsAwarded ?? 0"
        :results="completedSubmissionResults"
        :can-review="canReview"
        @start-review="handleStartReview"
      />
    </template>

    <!-- Quiz unavailable: can't start and no submissions -->
    <template v-else-if="isQuizUnavailable">
      <div class="flex items-center justify-center grow p-default">
        <p class="text-body text-text-secondary text-center">
          {{ $t('quiz.unavailable') }}
        </p>
      </div>
    </template>

    <!-- Final fallback: show loading for any unexpected state -->
    <template v-else>
      <div class="flex items-center justify-center grow">
        <LoadingState />
      </div>
    </template>

    <template #footer>
      <div
        v-if="actionState && !quizCompleted"
        class="w-full p-default flex flex-col"
      >
        <!-- Normal mode -->
        <template v-if="actionState.mode === 'normal'">
          <DesignButton
            v-if="!actionState.isAnswerLocked"
            size="large"
            :disabled="!actionState.canSubmit || actionState.isSubmitting"
            :loading="actionState.isSubmitting"
            @click="handlers?.submit"
          >
            {{ $t('quiz.lockAnswer') }}
          </DesignButton>
          <DesignButton v-else size="large" @click="handlers?.continue">
            {{ continueButtonText }}
          </DesignButton>
        </template>

        <!-- Session betting mode -->
        <template v-else-if="actionState.mode === 'session-betting'">
          <DesignButton
            v-if="!actionState.isBetSaved || actionState.isEditing"
            size="large"
            :disabled="actionState.isSubmitting"
            :loading="actionState.isSubmitting"
            @click="handlers?.submit"
          >
            {{ $t('quiz.saveAnswer') }}
          </DesignButton>
          <DesignButton
            v-else
            size="large"
            variant="secondary"
            @click="handlers?.changeBet"
          >
            {{ $t('quiz.changeAnswer') }}
          </DesignButton>
        </template>

        <!-- Session locked -->
        <template v-else-if="actionState.mode === 'session-locked'">
          <div class="text-center text-text-hint text-body py-small">
            {{ $t('quiz.betting.sessionLocked') }}
          </div>
        </template>

        <!-- Review mode -->
        <template v-else-if="actionState.mode === 'review'">
          <div class="flex gap-small">
            <DesignButton
              v-if="actionState.showPreviousButton"
              size="large"
              variant="secondary"
              class="flex-1"
              @click="handlers?.previous"
            >
              {{ $t('quiz.previousQuestion') }}
            </DesignButton>
            <DesignButton
              size="large"
              :class="actionState.showPreviousButton ? 'flex-1' : 'w-full'"
              @click="handlers?.next"
            >
              {{ nextButtonText }}
            </DesignButton>
          </div>
        </template>
      </div>
    </template>
  </PageLayout>
</template>
