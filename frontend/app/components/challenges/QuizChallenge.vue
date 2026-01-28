<script setup lang="ts">
import type {
  QuizChallengeData,
  QuestionResult,
  PredefinedQuestionData,
  PredefinedResponseData,
  OrderingQuestionData,
  OrderingResponseData,
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
      <h1 v-else class="text-heading text-text-default">
        {{ challenge.quiz.name }}
      </h1>
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
      <QuizPredefinedQuestion
        v-if="currentQuestion.__typename === 'PredefinedQuestion'"
        :key="`predefined:${currentQuestion.id}`"
        :question="currentQuestion"
        :total-questions="questions.length"
        :current-index="currentQuestionIndex"
        :submission-id="activeSubmission?.id ?? ''"
        @answer-submitted="handleAnswerSubmitted"
      />
      <QuizOrderingQuestion
        v-else-if="currentQuestion.__typename === 'OrderingQuestion'"
        :key="`ordering:${currentQuestion.id}`"
        :question="currentQuestion"
        :total-questions="questions.length"
        :current-index="currentQuestionIndex"
        :submission-id="activeSubmission?.id ?? ''"
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
  </PageLayout>
</template>
