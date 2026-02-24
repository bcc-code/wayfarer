<script setup lang="ts">
import type {
  QuizChallengeData,
  QuestionResult,
  PredefinedQuestionData,
  PredefinedResponseData,
  OrderingQuestionData,
  OrderingResponseData,
  NumberQuestionData,
  NumberResponseData,
  FreeTextQuestionData,
  FreeTextResponseData,
  QuizQuestionExposed,
} from './quiz/types'
import {
  QuizSessionState,
  type FinalizeQuizMutation,
  type StartQuizSessionMutation,
} from '~/api/generated'

const props = defineProps<{
  challenge: QuizChallengeData
  userScore?: number
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
const currentBetAmount = ref(0)
const finalResult = ref<FinalizeQuizMutation['finalizeQuiz'] | null>(null)
const startedSubmission = ref<
  StartQuizSessionMutation['startQuizSession'] | null
>(null)
const isReviewMode = ref(false)

// Start with loading true if we need to start a quiz (no active submission and have OPEN session)
const needsToStartQuiz = computed(() => {
  const session = props.challenge.quiz.userActiveSession
  return (
    !props.challenge.quiz.userActiveSubmission?.id &&
    session?.id != null &&
    session.state === QuizSessionState.Open
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

// Check if user can review their answers (has reviewable responses)
const canReview = computed(() => {
  if (!completedSubmission.value) return false
  return completedSubmission.value.responses.some(
    (r) =>
      r.__typename === 'PredefinedResponse' ||
      r.__typename === 'OrderingResponse' ||
      r.__typename === 'NumberResponse' ||
      r.__typename === 'FreeTextResponse',
  )
})

// Get questions for review mode
const reviewQuestions = computed<
  (
    | PredefinedQuestionData
    | OrderingQuestionData
    | NumberQuestionData
    | FreeTextQuestionData
  )[]
>(() => {
  if (!completedSubmission.value) return []
  return completedSubmission.value.orderedQuestions.filter(
    (
      q,
    ): q is
      | PredefinedQuestionData
      | OrderingQuestionData
      | NumberQuestionData
      | FreeTextQuestionData =>
      q.__typename === 'PredefinedQuestion' ||
      q.__typename === 'OrderingQuestion' ||
      q.__typename === 'NumberQuestion' ||
      q.__typename === 'FreeTextQuestion',
  )
})

// Get responses for review mode
const reviewResponses = computed<
  (
    | PredefinedResponseData
    | OrderingResponseData
    | NumberResponseData
    | FreeTextResponseData
  )[]
>(() => {
  if (!completedSubmission.value) return []
  return completedSubmission.value.responses.filter(
    (
      r,
    ): r is
      | PredefinedResponseData
      | OrderingResponseData
      | NumberResponseData
      | FreeTextResponseData =>
      r.__typename === 'PredefinedResponse' ||
      r.__typename === 'OrderingResponse' ||
      r.__typename === 'NumberResponse' ||
      r.__typename === 'FreeTextResponse',
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
  // Determine the target submission ID
  const submissionId =
    startedSubmission.value?.id ??
    props.challenge.quiz.userActiveSubmission?.id

  // Prefer props data when available (has full responses with pointsEarned from server)
  if (submissionId) {
    const fromProps = props.challenge.quiz.userSubmissions.find(
      (submission) => submission.id === submissionId,
    )
    if (fromProps) return fromProps
  }

  // Fallback to started submission (before page has refetched to include this submission)
  if (startedSubmission.value) {
    return startedSubmission.value
  }

  // If session is LOCKED or FINISHED but no active submission, use the completed submission
  const sessionState = props.challenge.quiz.userActiveSession?.state
  if (
    sessionState === QuizSessionState.Locked ||
    sessionState === QuizSessionState.Finished
  ) {
    return props.challenge.quiz.userSubmissions.find((s) => s.completedAt)
  }

  return undefined
})

const questions = computed(() => {
  return activeSubmission.value?.orderedQuestions ?? []
})

// Initialize from existing responses when resuming a quiz
function initializeFromExistingResponses() {
  const submission = activeSubmission.value
  if (!submission || !('responses' in submission)) return

  const responses = submission.responses
  if (responses.length === 0) return

  // Build question results from existing responses
  const existingResults: QuestionResult[] = []
  for (const question of questions.value) {
    const response = responses.find((r) => r.question.id === question.id)
    if (response) {
      const isCorrect =
        response.__typename === 'PredefinedResponse' ||
        response.__typename === 'OrderingResponse'
          ? (response.isCorrect ?? null)
          : null
      existingResults.push({ questionId: question.id, isCorrect })
    }
  }

  // Set question results for progress display
  questionResults.value = existingResults

  // Skip to first unanswered question
  const firstUnansweredIndex = questions.value.findIndex(
    (q) => !responses.some((r) => r.question.id === q.id),
  )
  if (firstUnansweredIndex !== -1) {
    currentQuestionIndex.value = firstUnansweredIndex
  } else if (responses.length === questions.value.length) {
    // All questions answered, stay on last question
    currentQuestionIndex.value = questions.value.length - 1
  }
}

// Initialize when component mounts if we have an existing submission with responses
watch(
  activeSubmission,
  (submission) => {
    if (
      submission &&
      'responses' in submission &&
      submission.responses.length > 0
    ) {
      initializeFromExistingResponses()
    }
  },
  { immediate: true },
)

const currentQuestion = computed(() => {
  return questions.value[currentQuestionIndex.value]
})

const isLastQuestion = computed(() => {
  return currentQuestionIndex.value === questions.value.length - 1
})

// Check if current question has betting enabled
const isBettingEnabled = computed(() => {
  return currentQuestion.value?.bettingEnabled ?? false
})

// Determine if the user won or lost the bet (for background color styling)
const isBettingWin = computed(() => {
  if (!isBettingEnabled.value) return null
  const pointsEarned = currentResponse.value?.pointsEarned
  if (pointsEarned === null || pointsEarned === undefined) return null
  return pointsEarned > 0
})

// Get session state for ordering questions betting mode
const sessionState = computed(() => {
  return props.challenge.quiz.userActiveSession?.state
})

// Check if session is in a state where we should show the question view (LOCKED or FINISHED)
const isSessionLockedOrFinished = computed(() => {
  return (
    sessionState.value === QuizSessionState.Locked ||
    sessionState.value === QuizSessionState.Finished
  )
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

// Populate bet amount from existing response, or reset to 0
watch(
  [currentQuestionIndex, currentResponse],
  () => {
    currentBetAmount.value = currentResponse.value?.betAmount ?? 0
  },
  { immediate: true },
)

async function handleAnswerSubmitted(result: QuestionResult) {
  // Update existing result or add new one (avoid duplicates when resuming)
  const existingIndex = questionResults.value.findIndex(
    (r) => r.questionId === result.questionId,
  )
  if (existingIndex !== -1) {
    questionResults.value[existingIndex] = result
  } else {
    questionResults.value.push(result)
  }

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

// Auto-advance when revealCorrectAnswers is false (skip showing the locked state)
// Don't auto-advance on the last question - let the user click "Finish"
watch(
  () => actionState.value?.isAnswerLocked,
  (isLocked) => {
    if (
      isLocked &&
      actionState.value?.mode === 'normal' &&
      !props.challenge.quiz.revealCorrectAnswers &&
      !isLastQuestion.value
    ) {
      handlers.value?.continue()
    }
  },
)

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

// QuizProgress visibility and props for both active quiz and review mode
const showQuizProgress = computed(() => {
  // Show during active quiz (multiple questions, not completed)
  if (
    activeSubmission.value &&
    !quizCompleted.value &&
    questions.value.length > 1
  ) {
    return true
  }
  // Show during review mode (multiple questions)
  if (isReviewMode.value && reviewQuestions.value.length > 1) {
    return true
  }
  return false
})

const progressCurrentIndex = computed(() => {
  if (isReviewMode.value) {
    const idx = reviewModeRef.value?.currentQuestionIndex
    return idx !== undefined ? unref(idx) : 0
  }
  return currentQuestionIndex.value
})

const progressTotalQuestions = computed(() => {
  if (isReviewMode.value) {
    return reviewQuestions.value.length
  }
  return questions.value.length
})

const progressResults = computed(() => {
  if (isReviewMode.value) {
    // Build results aligned with reviewQuestions order
    // Each result at index i corresponds to reviewQuestions[i]
    const responseMap = new Map(
      completedSubmissionResults.value.map((r) => [r.questionId, r]),
    )
    return reviewQuestions.value.map((q) => {
      const result = responseMap.get(q.id)
      return result ?? { questionId: q.id, isCorrect: null }
    })
  }
  return questionResults.value
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
        v-if="showQuizProgress"
        :current-index="progressCurrentIndex"
        :total-questions="progressTotalQuestions"
        :results="progressResults"
        :reveal-correct-answers="challenge.quiz.revealCorrectAnswers"
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
        :completed-at="finalResult.completedAt"
        :reveal-correct-answers="challenge.quiz.revealCorrectAnswers"
      />
    </template>

    <template
      v-else-if="
        completedSubmission && !canStartQuiz && !isSessionLockedOrFinished
      "
    >
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
        :completed-at="completedSubmission.completedAt"
        :reveal-correct-answers="challenge.quiz.revealCorrectAnswers"
        @start-review="handleStartReview"
      />
    </template>

    <template v-else-if="currentQuestion">
      <div
        v-show="isBettingEnabled && sessionState === QuizSessionState.Locked"
        class="grow flex flex-col py-6 px-default items-center justify-center"
      >
        <p class="text-heading text-balance text-center">
          {{ $t('quiz.betting.bettingClosed') }}
        </p>
        <p class="text-title text-text-muted">
          {{ $t('quiz.betting.yourBetIs', { bet: currentBetAmount }) }}
        </p>
      </div>
      <div
        v-show="!isBettingEnabled || sessionState !== QuizSessionState.Locked"
      >
        <div
          class="flex flex-col items-center justify-center py-3 px-medium gap-1 text-center"
        >
          <p v-if="questions.length > 1" class="text-caption text-text-muted">
            {{
              $t('quiz.questionNumber', {
                current: currentQuestionIndex + 1,
                total: questions.length,
              })
            }}
          </p>
          <h1
            class="text-heading limitedHeight:text-label text-text-default text-balance"
          >
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
          :existing-response="
            currentResponse?.__typename === 'PredefinedResponse'
              ? currentResponse
              : undefined
          "
          :is-last-question="isLastQuestion"
          :reveal-correct-answers="challenge.quiz.revealCorrectAnswers"
          :bet-amount="isBettingEnabled ? currentBetAmount : undefined"
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
          :reveal-correct-answers="challenge.quiz.revealCorrectAnswers"
          :bet-amount="isBettingEnabled ? currentBetAmount : undefined"
          @answer-submitted="handleAnswerSubmitted"
        />
        <QuizNumberQuestion
          v-else-if="currentQuestion.__typename === 'NumberQuestion'"
          ref="currentQuestionRef"
          :key="`number:${currentQuestion.id}`"
          :question="currentQuestion"
          :total-questions="questions.length"
          :current-index="currentQuestionIndex"
          :submission-id="activeSubmission?.id ?? ''"
          :existing-response="
            currentResponse?.__typename === 'NumberResponse'
              ? currentResponse
              : undefined
          "
          :is-last-question="isLastQuestion"
          :reveal-correct-answers="challenge.quiz.revealCorrectAnswers"
          :bet-amount="isBettingEnabled ? currentBetAmount : undefined"
          @answer-submitted="handleAnswerSubmitted"
        />
        <QuizJsonQuestion
          v-else-if="currentQuestion.__typename === 'JsonQuestion'"
          ref="currentQuestionRef"
          :key="`json:${currentQuestion.id}`"
          :question="currentQuestion"
          :total-questions="questions.length"
          :current-index="currentQuestionIndex"
          :submission-id="activeSubmission?.id ?? ''"
          :is-last-question="isLastQuestion"
          :reveal-correct-answers="challenge.quiz.revealCorrectAnswers"
          @answer-submitted="handleAnswerSubmitted"
        />
        <QuizFreeTextQuestion
          v-else-if="currentQuestion.__typename === 'FreeTextQuestion'"
          ref="currentQuestionRef"
          :key="`free-text:${currentQuestion.id}`"
          :question="currentQuestion"
          :total-questions="questions.length"
          :current-index="currentQuestionIndex"
          :submission-id="activeSubmission?.id ?? ''"
          :existing-response="
            currentResponse?.__typename === 'FreeTextResponse'
              ? currentResponse
              : undefined
          "
          :is-last-question="isLastQuestion"
          :reveal-correct-answers="challenge.quiz.revealCorrectAnswers"
          :bet-amount="isBettingEnabled ? currentBetAmount : undefined"
          @answer-submitted="handleAnswerSubmitted"
        />
      </div>
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
        :completed-at="completedSubmission.completedAt"
        :reveal-correct-answers="challenge.quiz.revealCorrectAnswers"
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
        :class="[
          'w-full p-default flex flex-col gap-4 shadow-large',
          {
            'bg-background-raised':
              isBettingEnabled && sessionState !== QuizSessionState.Finished,
            'bg-accent-positive':
              isBettingEnabled &&
              sessionState === QuizSessionState.Finished &&
              isBettingWin === true,
            'bg-accent-negative':
              isBettingEnabled &&
              sessionState === QuizSessionState.Finished &&
              isBettingWin === false,
          },
        ]"
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
            {{
              challenge.quiz.revealCorrectAnswers
                ? $t('quiz.lockAnswer')
                : continueButtonText
            }}
          </DesignButton>
          <DesignButton v-else size="large" @click="handlers?.continue">
            {{ continueButtonText }}
          </DesignButton>
        </template>

        <!-- Session betting mode -->
        <template v-else-if="actionState.mode === 'session-betting'">
          <QuizBettingModule
            v-if="isBettingEnabled && !actionState.isAnswerLocked"
            v-model="currentBetAmount"
            :available-points="userScore ?? 0"
            :min-percentage="currentQuestion?.bettingMinPercentage"
            :max-percentage="currentQuestion?.bettingMaxPercentage"
            :min-absolute="currentQuestion?.bettingMinAbsolute"
            :max-absolute="currentQuestion?.bettingMaxAbsolute"
            :disabled="actionState.isBetSaved && !actionState.isEditing"
            mode="betting"
          />
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
          <QuizBettingModule
            v-if="isBettingEnabled && !actionState.isAnswerLocked"
            v-model="currentBetAmount"
            :available-points="userScore ?? 0"
            :min-percentage="currentQuestion?.bettingMinPercentage"
            :max-percentage="currentQuestion?.bettingMaxPercentage"
            :min-absolute="currentQuestion?.bettingMinAbsolute"
            :max-absolute="currentQuestion?.bettingMaxAbsolute"
            :disabled="actionState.isBetSaved && !actionState.isEditing"
            mode="locked"
          />
          <NuxtLink :to="{ name: 'challenges' }" class="flex">
            <DesignButton size="large" variant="secondary">
              {{ $t('quiz.close') }}
            </DesignButton>
          </NuxtLink>
        </template>

        <!-- Session results -->
        <template v-else-if="actionState.mode === 'session-results'">
          <QuizBettingModule
            v-if="isBettingEnabled"
            mode="results"
            :points-earned="currentResponse?.pointsEarned"
            :bet-amount="currentResponse?.betAmount"
            :available-points="userScore ?? 0"
          />
          <NuxtLink :to="{ name: 'challenges' }" class="flex">
            <DesignButton
              size="large"
              variant="primary"
              :class="[
                'bg-black',
                {
                  'text-accent-positive!': isBettingEnabled && isBettingWin,
                  'text-accent-negative!': isBettingEnabled && !isBettingWin,
                },
              ]"
            >
              {{ $t('quiz.done') }}
            </DesignButton>
          </NuxtLink>
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
