<script setup lang="ts">
import type { ChallengeFormData } from '~/components/admin/challenge/AdminChallengeForm.vue'
import type { QuizFormData } from '~/components/admin/quiz/AdminQuizForm.vue'

definePageMeta({
  layout: 'admin',
  middleware: 'admin',
})

gql(`
  query AdminProjectChallengePage($challengeId: ID!) {
    challenge(id: $challengeId) {
      __typename
      id
      name
      description
      image
      buttonText
      publishedAt
      visibleAt
      startedAt
      endTime
      project {
        id
        name
        branding {
          colors {
            ...BrandingColorsFields
          }
        }
      }
      ... on SimpleChallenge {
        allowSelfCompletion
      }
      ... on ExternalChallenge {
        url
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
      ... on PluginChallenge {
        pluginChallengeId
      }
    }
  }
`)

const route = useRoute('admin-projects-projectId-challenges-challengeId')
const toast = useToast()

const { isAuthReady } = useAuthReady()
const { data, fetching, error } = useAdminProjectChallengePageQuery({
  variables: {
    challengeId: route.params.challengeId,
  },
  pause: computed(() => !isAuthReady.value),
})

const { executeMutation } = useUpdateChallengeMutation()
const { executeMutation: executeDelete } = useDeleteChallengeMutation()
const { executeMutation: createQuiz } = useCreateQuizMutation()
const { executeMutation: updateQuiz } = useUpdateQuizMutation()
const { executeMutation: addQuizQuestion } = useAddQuizQuestionMutation()
const { executeMutation: updateQuizQuestion } = useUpdateQuizQuestionMutation()
const { executeMutation: deleteQuizQuestion } = useDeleteQuizQuestionMutation()

function getChallengeType(typename: string): ChallengeType {
  switch (typename) {
    case 'ExternalChallenge':
      return ChallengeType.External
    case 'QuizChallenge':
      return ChallengeType.Quiz
    case 'PluginChallenge':
      return ChallengeType.Plugin
    default:
      return ChallengeType.Simple
  }
}

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

const initialData = computed(() => {
  if (!data.value) return undefined
  const c = data.value.challenge
  return {
    type: getChallengeType(c.__typename ?? 'SimpleChallenge'),
    name: c.name,
    description: c.description ?? undefined,
    image: c.image ?? undefined,
    url: c.__typename === 'ExternalChallenge' ? c.url : undefined,
    buttonText: c.buttonText,
    publishedAt: toLocalDatetimeLocal(c.publishedAt),
    endTime: toLocalDatetimeLocal(c.endTime),
    visibleAt: toLocalDatetimeLocal(c.visibleAt),
    startedAt: toLocalDatetimeLocal(c.startedAt),
    allowSelfCompletion:
      c.__typename === 'SimpleChallenge' ? c.allowSelfCompletion : undefined,
    pluginChallengeId:
      c.__typename === 'PluginChallenge' ? c.pluginChallengeId : undefined,
  }
})

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

async function saveQuiz(quizFormData: QuizFormData, challengeId: string) {
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
      throw new Error(updateResult.error.message)
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
        challengeId,
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
      throw new Error(createResult.error.message)
    }

    const quizId = createResult.data?.createQuiz.id
    if (!quizId) {
      throw new Error('Failed to create quiz')
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
}

async function handleSubmit(formData: ChallengeFormData) {
  const {
    type,
    allowSelfCompletion,
    url,
    quiz,
    publishedAt,
    endTime,
    visibleAt,
    startedAt,
    pluginChallengeId,
    ...rest
  } = formData

  // Only include type-specific fields
  const input = {
    ...rest,
    publishedAt: toISOString(publishedAt),
    endTime: toISOString(endTime),
    visibleAt: toISOString(visibleAt),
    startedAt: toISOString(startedAt),
    ...(type === ChallengeType.Simple && { allowSelfCompletion }),
    ...(type === ChallengeType.External && { url }),
    ...(type === ChallengeType.Plugin && { pluginChallengeId }),
  }

  const response = await executeMutation({
    id: route.params.challengeId,
    input,
  })

  if (response.error) {
    toast.add({
      title: response.error.name,
      description: response.error.message,
      color: 'error',
    })
    return
  }

  // Handle quiz if this is a quiz challenge
  if (type === ChallengeType.Quiz && quiz) {
    try {
      await saveQuiz(quiz, route.params.challengeId)
    } catch (err) {
      toast.add({
        title: 'Feil',
        description:
          err instanceof Error ? err.message : 'Kunne ikke lagre quiz',
        color: 'error',
      })
      return
    }
  }

  toast.add({
    title: 'Suksess',
    description: 'Utfordring oppdatert',
    color: 'success',
  })
  navigateTo({
    name: 'admin-projects-projectId',
    params: { projectId: route.params.projectId },
  })
}

async function handleDelete() {
  const confirmed = confirm(
    `Er du sikker på at du vil slette "${data.value?.challenge.name}"? Denne handlingen kan ikke angres.`,
  )

  if (!confirmed) return

  const response = await executeDelete({ id: route.params.challengeId })
  if (response.error) {
    toast.add({
      title: response.error.name,
      description: response.error.message,
      color: 'error',
    })
    return
  }

  toast.add({
    title: 'Suksess',
    description: 'Utfordring slettet',
    color: 'success',
  })
  navigateTo({
    name: 'admin-projects-projectId',
    params: { projectId: route.params.projectId },
  })
}
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
              label: 'Utfordringer',
            },
            {
              label: data?.challenge.name ?? route.params.challengeId,
              to: {
                name: 'admin-projects-projectId-challenges-challengeId',
                params: {
                  projectId: route.params.projectId,
                  challengeId: route.params.challengeId,
                },
              },
            },
          ]"
        />
      </UContainer>
    </div>
    <UContainer class="py-12">
      <LoadingState v-if="fetching" />
      <ErrorState v-else-if="error" :error />
      <AdminChallengeForm
        v-else-if="initialData"
        :initial-data="initialData"
        :quiz-data="quizData"
        :project-id="route.params.projectId"
        :colors="data?.challenge.project.branding.colors"
        submit-label="Lagre endringer"
        is-edit-mode
        :on-delete="handleDelete"
        @submit="handleSubmit"
      />
    </UContainer>
  </div>
</template>
