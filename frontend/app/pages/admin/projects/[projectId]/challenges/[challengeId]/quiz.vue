<script setup lang="ts">
import type { QuizFormData } from '~/components/admin/quiz/AdminQuizForm.vue'

definePageMeta({
  layout: 'admin',
  middleware: 'admin',
})

gql(`
  query AdminChallengeQuizPage($challengeId: ID!) {
    challenge(id: $challengeId) {
      __typename
      id
      name
      project {
        id
        name
      }
      ... on QuizChallenge {
        quiz {
          id
          name
          description
          image
          timeoutSeconds
          randomizeQuestions
          revealCorrectAnswers
          allowRetakes
          completionPoints
          questions {
            ...QuizQuestionFields
          }
        }
      }
    }
  }
`)

const route = useRoute('admin-projects-projectId-challenges-challengeId-quiz')
const toast = useToast()

const { isAuthReady } = useAuthReady()
const { data, fetching, error, executeQuery: refetchQuiz } = useAdminChallengeQuizPageQuery({
  variables: {
    challengeId: route.params.challengeId,
  },
  pause: computed(() => !isAuthReady.value),
})

const { executeMutation: createQuiz } = useCreateQuizMutation()
const { executeMutation: updateQuiz } = useUpdateQuizMutation()
const { executeMutation: addQuizQuestion } = useAddQuizQuestionMutation()
const { executeMutation: updateQuizQuestion } = useUpdateQuizQuestionMutation()
const { executeMutation: deleteQuizQuestion } = useDeleteQuizQuestionMutation()

function getQuestionType(typename: string): QuizQuestionType {
  switch (typename) {
    case 'NumberQuestion':
      return QuizQuestionType.Number
    case 'FreeTextQuestion':
      return QuizQuestionType.FreeText
    case 'JsonQuestion':
      return QuizQuestionType.Json
    case 'OrderingQuestion':
      return QuizQuestionType.Ordering
    default:
      return QuizQuestionType.Predefined
  }
}

const quizData = computed<QuizFormData | undefined>(() => {
  if (!data.value) return undefined
  const c = data.value.challenge
  if (c.__typename !== 'QuizChallenge' || !c.quiz) return undefined

  return {
    id: c.quiz.id,
    name: c.quiz.name,
    description: c.quiz.description,
    image: c.quiz.image ?? undefined,
    timeoutSeconds: c.quiz.timeoutSeconds ?? undefined,
    randomizeQuestions: c.quiz.randomizeQuestions,
    revealCorrectAnswers: c.quiz.revealCorrectAnswers,
    allowRetakes: c.quiz.allowRetakes,
    completionPoints: c.quiz.completionPoints,
    questions: c.quiz.questions.map((q) => ({
      id: q.id,
      questionType: getQuestionType(q.__typename ?? 'PredefinedQuestion'),
      questionText: q.questionText,
      questionOrder: q.questionOrder,
      timeoutSeconds: q.timeoutSeconds ?? undefined,
      points: q.points ?? undefined,
      allowMultipleSelection:
        q.__typename === 'PredefinedQuestion'
          ? q.allowMultipleSelection
          : undefined,
      predefinedAnswers:
        q.__typename === 'PredefinedQuestion'
          ? q.predefinedAnswers.map((a) => ({
              id: a.id,
              answerText: a.answerText,
              isCorrect: a.isCorrect ?? false,
              answerOrder: a.answerOrder,
            }))
          : undefined,
      orderingItems:
        q.__typename === 'OrderingQuestion'
          ? q.orderingItems.map((item, index) => ({
              id: item.id,
              itemText: item.itemText,
              correctOrder: index + 1,
            }))
          : undefined,
      minValue:
        q.__typename === 'NumberQuestion'
          ? (q.minValue ?? undefined)
          : undefined,
      maxValue:
        q.__typename === 'NumberQuestion'
          ? (q.maxValue ?? undefined)
          : undefined,
      stepValue:
        q.__typename === 'NumberQuestion'
          ? (q.stepValue ?? undefined)
          : undefined,
    })),
  }
})

const isNewQuiz = computed(() => !quizData.value?.id)

async function saveQuiz(quizFormData: QuizFormData) {
  if (quizFormData.id) {
    // Update existing quiz
    const updateResult = await updateQuiz({
      id: quizFormData.id,
      input: {
        name: quizFormData.name,
        description: quizFormData.description,
        image: quizFormData.image,
        timeoutSeconds: quizFormData.timeoutSeconds,
        randomizeQuestions: quizFormData.randomizeQuestions,
        revealCorrectAnswers: quizFormData.revealCorrectAnswers,
        allowRetakes: quizFormData.allowRetakes,
        completionPoints: quizFormData.completionPoints,
      },
    })

    if (updateResult.error) {
      toast.add({
        title: 'Feil',
        description: updateResult.error.message,
        color: 'error',
      })
      return
    }

    // Handle questions - delete removed, update existing, add new
    const existingQuestions = quizData.value?.questions ?? []
    const newQuestions = quizFormData.questions

    // Delete removed questions
    for (const existing of existingQuestions) {
      if (existing.id && !newQuestions.find((q) => q.id === existing.id)) {
        await deleteQuizQuestion({ id: existing.id })
      }
    }

    // Update existing and add new questions
    for (const question of newQuestions) {
      if (question.id) {
        // Update existing
        await updateQuizQuestion({
          id: question.id,
          input: {
            questionText: question.questionText,
            questionOrder: question.questionOrder,
            timeoutSeconds: question.timeoutSeconds,
            points: question.points,
            allowMultipleSelection: question.allowMultipleSelection,
            predefinedAnswers: question.predefinedAnswers?.map((a) => ({
              answerText: a.answerText,
              isCorrect: a.isCorrect,
              answerOrder: a.answerOrder,
            })),
            orderingItems: question.orderingItems?.map((item) => ({
              itemText: item.itemText,
              correctOrder: item.correctOrder,
            })),
            minValue: question.minValue,
            maxValue: question.maxValue,
            stepValue: question.stepValue,
          },
        })
      } else {
        // Add new
        await addQuizQuestion({
          quizId: quizFormData.id,
          input: {
            questionType: question.questionType,
            questionText: question.questionText,
            questionOrder: question.questionOrder,
            timeoutSeconds: question.timeoutSeconds,
            points: question.points,
            allowMultipleSelection: question.allowMultipleSelection,
            predefinedAnswers: question.predefinedAnswers?.map((a) => ({
              answerText: a.answerText,
              isCorrect: a.isCorrect,
              answerOrder: a.answerOrder,
            })),
            orderingItems: question.orderingItems?.map((item) => ({
              itemText: item.itemText,
              correctOrder: item.correctOrder,
            })),
            minValue: question.minValue,
            maxValue: question.maxValue,
            stepValue: question.stepValue,
          },
        })
      }
    }
  } else {
    // Create new quiz
    const createResult = await createQuiz({
      input: {
        projectId: route.params.projectId,
        challengeId: route.params.challengeId,
        name: quizFormData.name,
        description: quizFormData.description,
        image: quizFormData.image,
        timeoutSeconds: quizFormData.timeoutSeconds,
        randomizeQuestions: quizFormData.randomizeQuestions,
        revealCorrectAnswers: quizFormData.revealCorrectAnswers,
        allowRetakes: quizFormData.allowRetakes,
        completionPoints: quizFormData.completionPoints,
      },
    })

    if (createResult.error) {
      toast.add({
        title: 'Feil',
        description: createResult.error.message,
        color: 'error',
      })
      return
    }

    const quizId = createResult.data?.createQuiz.id
    if (!quizId) {
      toast.add({
        title: 'Feil',
        description: 'Kunne ikke opprette quiz',
        color: 'error',
      })
      return
    }

    // Add questions
    for (const question of quizFormData.questions) {
      await addQuizQuestion({
        quizId,
        input: {
          questionType: question.questionType,
          questionText: question.questionText,
          questionOrder: question.questionOrder,
          timeoutSeconds: question.timeoutSeconds,
          points: question.points,
          allowMultipleSelection: question.allowMultipleSelection,
          predefinedAnswers: question.predefinedAnswers?.map((a) => ({
            answerText: a.answerText,
            isCorrect: a.isCorrect,
            answerOrder: a.answerOrder,
          })),
          orderingItems: question.orderingItems?.map((item) => ({
            itemText: item.itemText,
            correctOrder: item.correctOrder,
          })),
          minValue: question.minValue,
          maxValue: question.maxValue,
          stepValue: question.stepValue,
        },
      })
    }
  }

  // Refetch to update cache before navigating
  await refetchQuiz({ requestPolicy: 'network-only' })

  toast.add({
    title: 'Suksess',
    description: isNewQuiz.value ? 'Quiz opprettet' : 'Quiz oppdatert',
    color: 'success',
  })

  navigateTo({
    name: 'admin-projects-projectId-challenges-challengeId',
    params: {
      projectId: route.params.projectId,
      challengeId: route.params.challengeId,
    },
  })
}

// Check if challenge is a quiz challenge
const isQuizChallenge = computed(() => {
  return data.value?.challenge.__typename === 'QuizChallenge'
})
</script>

<template>
  <div>
    <div class="border-default border-b py-2">
      <UContainer>
        <UBreadcrumb
          :items="[
            {
              label: 'Prosjekter',
              to: { name: 'admin-projects' },
            },
            {
              label: data?.challenge.project.name ?? route.params.projectId,
              to: {
                name: 'admin-projects-projectId',
                params: { projectId: route.params.projectId },
              },
            },
            {
              label: data?.challenge.name ?? 'Utfordring',
              to: {
                name: 'admin-projects-projectId-challenges-challengeId',
                params: {
                  projectId: route.params.projectId,
                  challengeId: route.params.challengeId,
                },
              },
            },
            {
              label: 'Quiz',
            },
          ]"
        />
      </UContainer>
    </div>
    <UContainer class="py-12">
      <LoadingState v-if="fetching" />
      <ErrorState v-else-if="error" :error />
      <template v-else-if="data">
        <div v-if="!isQuizChallenge" class="text-center py-12">
          <p class="text-text-muted">
            Denne utfordringen er ikke en quiz-utfordring.
          </p>
          <UButton
            class="mt-4"
            :to="{
              name: 'admin-projects-projectId-challenges-challengeId',
              params: {
                projectId: route.params.projectId,
                challengeId: route.params.challengeId,
              },
            }"
          >
            Tilbake til utfordring
          </UButton>
        </div>
        <template v-else>
          <h1 class="mb-6 text-2xl font-bold">
            {{ isNewQuiz ? 'Opprett quiz' : 'Rediger quiz' }}
          </h1>
          <AdminQuizForm
            :quiz-data="quizData"
            :project-id="route.params.projectId"
            :challenge-id="route.params.challengeId"
            @save="saveQuiz"
          />
        </template>
      </template>
    </UContainer>
  </div>
</template>
