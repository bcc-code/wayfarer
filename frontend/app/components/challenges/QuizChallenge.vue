<script setup lang="ts">
import type { ChallengePageQuery } from '~/api/generated'

type QuizChallengeData = Extract<
  ChallengePageQuery['challenge'],
  { __typename: 'QuizChallenge' }
>

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
  return activeSubmission.value?.orderedQuestions
})
</script>

<template>
  <PageLayout>
    <template #action>
      <NuxtLink :to="{ name: 'challenges' }">
        <DesignIconButton icon="lucide:x" />
      </NuxtLink>
    </template>
    <template #title>
      <QuizProgress :submission="activeSubmission" />
    </template>

    <div v-for="question in questions" :key="question.id">
      {{ question.questionText }}
    </div>
  </PageLayout>
</template>
