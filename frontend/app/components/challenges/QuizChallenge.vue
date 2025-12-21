<script setup lang="ts">
import type { QuizChallengeData, QuestionResult } from './quiz/types'
import type { FinalizeQuizMutation, StartQuizMutation } from '~/api/generated'

const props = defineProps<{
  challenge: QuizChallengeData
}>()

const emit = defineEmits<{
  start: []
  complete: []
}>()

const { track } = useAnalytics()
const { executeMutation: startQuiz } = useStartQuizMutation()
const { executeMutation: finalizeQuiz } = useFinalizeQuizMutation()

const currentQuestionIndex = ref(0)
const questionResults = ref<QuestionResult[]>([])
const quizCompleted = ref(false)
const finalResult = ref<FinalizeQuizMutation['finalizeQuiz'] | null>(null)
const startedSubmission = ref<StartQuizMutation['startQuiz'] | null>(null)

// Start with loading true if we need to start a quiz (no active submission and can start)
const needsToStartQuiz = computed(() => {
  return (
    !props.challenge.quiz.userActiveSubmission?.id &&
    props.challenge.quiz.userCanStart
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
      response.__typename === 'PredefinedResponse'
        ? (response.isCorrect ?? null)
        : null,
  }))
})

// Check if we can start a new quiz
const canStartQuiz = computed(() => {
  return props.challenge.quiz.userCanStart
})

onMounted(async () => {
  track(AnalyticsEvent.ChallengeOpened, {
    challenge_id: props.challenge.id,
    challenge_name: props.challenge.name,
    challenge_type: 'quiz',
  })

  // If there's no active submission and we can start, start the quiz
  if (needsToStartQuiz.value) {
    const result = await startQuiz({
      quizId: props.challenge.quiz.id,
    })
    if (result.data?.startQuiz) {
      startedSubmission.value = result.data.startQuiz
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
        v-if="activeSubmission && !quizCompleted"
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
      <QuizResult
        :score="completedSubmission.score ?? 0"
        :max-score="completedSubmission.maxScore ?? 0"
        :points-awarded="completedSubmission.pointsAwarded ?? 0"
        :results="completedSubmissionResults"
      />
    </template>

    <template v-else-if="currentQuestion">
      <QuizPredefinedQuestion
        v-if="currentQuestion.__typename === 'PredefinedQuestion'"
        :key="currentQuestion.id"
        :question="currentQuestion"
        :total-questions="questions.length"
        :current-index="currentQuestionIndex"
        :submission-id="activeSubmission?.id ?? ''"
        @answer-submitted="handleAnswerSubmitted"
      />
      <QuizNumberQuestion
        v-else-if="currentQuestion.__typename === 'NumberQuestion'"
        :question="currentQuestion"
      />
      <QuizJsonQuestion
        v-else-if="currentQuestion.__typename === 'JsonQuestion'"
        :question="currentQuestion"
      />
      <QuizFreeTextQuestion
        v-else-if="currentQuestion.__typename === 'FreeTextQuestion'"
        :question="currentQuestion"
      />
    </template>

    <!-- Fallback: show completed submission result even if retakes are allowed -->
    <template v-else-if="completedSubmission">
      <QuizResult
        :score="completedSubmission.score ?? 0"
        :max-score="completedSubmission.maxScore ?? 0"
        :points-awarded="completedSubmission.pointsAwarded ?? 0"
        :results="completedSubmissionResults"
      />
    </template>

    <!-- Final fallback: show loading for any unexpected state -->
    <template v-else>
      <div class="flex items-center justify-center grow">
        <LoadingState />
      </div>
    </template>
  </PageLayout>
</template>
