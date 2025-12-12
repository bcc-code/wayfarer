<script setup lang="ts">
import type { QuizChallengeData } from './quiz/types'

const props = defineProps<{
  challenge: QuizChallengeData
}>()

const emit = defineEmits<{
  start: []
}>()

const { executeMutation: startQuiz } = useStartQuizMutation()

onMounted(() => {
  if (!props.challenge.quiz.userActiveSubmission?.id) {
    startQuiz({
      quizId: props.challenge.quiz.id,
    }).then(() => {
      emit('start')
    })
  }
})

const activeSubmission = computed(() => {
  return props.challenge.quiz.userSubmissions.find(
    (submission) =>
      submission.id === props.challenge.quiz.userActiveSubmission?.id,
  )
})

const questions = computed(() => {
  if (props.challenge.quiz.randomizeQuestions) {
    return activeSubmission.value?.orderedQuestions.toSorted(
      () => Math.random() - 0.5,
    )
  }
  return activeSubmission.value?.orderedQuestions
})
</script>

<template>
  <PageLayout :bottom-padding="false">
    <template #action>
      <NuxtLink :to="{ name: 'challenges' }">
        <DesignIconButton icon="lucide:x" />
      </NuxtLink>
    </template>
    <template #title>
      <QuizProgress v-if="activeSubmission" :submission="activeSubmission" />
    </template>

    <template v-if="questions?.length">
      <template v-for="question in questions" :key="question.id">
        <QuizPredefinedQuestion
          v-if="question.__typename === 'PredefinedQuestion'"
          :question="question"
          :total-questions="questions.length"
        />
        <QuizNumberQuestion
          v-else-if="question.__typename === 'NumberQuestion'"
          :question="question"
        />
        <QuizJsonQuestion
          v-else-if="question.__typename === 'JsonQuestion'"
          :question="question"
        />
        <QuizFreeTextQuestion
          v-else-if="question.__typename === 'FreeTextQuestion'"
          :question="question"
        />
      </template>
    </template>
  </PageLayout>
</template>
