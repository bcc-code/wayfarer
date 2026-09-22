<script setup lang="ts">
import type { RouteLocationRaw } from 'vue-router'
import type {
  QuizFormData,
  QuizQuestionFormData,
} from '../../../../../../components/admin/quiz/AdminQuizForm.vue'
import {
  planQuizQuestionSave,
  orderedQuestionIds,
  bettingClearFlags,
} from '../../../../../../utils/quizSave'

definePageMeta({
  permission: 'projects:view',
  layout: 'admin',
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
          translationStatus {
            ...TranslationStatus
          }
        }
      }
    }
  }
`)

gql(`
  mutation ReorderQuizQuestions($quizId: ID!, $questionIds: [ID!]!) {
    reorderQuizQuestions(quizId: $quizId, questionIds: $questionIds) {
      id
      questionOrder
    }
  }
`)

const route = useRoute('admin-projects-projectId-challenges-challengeId-quiz')
const toast = useToast()

const { isAuthReady } = useAuthReady()
const {
  data,
  fetching,
  error,
  executeQuery: refetchQuiz,
} = useAdminChallengeQuizPageQuery({
  variables: computed(() => ({
    challengeId: route.params.challengeId,
  })),
  pause: computed(() => !isAuthReady.value),
})

// Trailing breadcrumb crumbs; the path above them is derived from the route.
useAdminPage(() => [
  {
    label: data.value?.challenge.name ?? 'Utfordring',
    to: {
      name: 'admin-projects-projectId-challenges-challengeId',
      params: {
        projectId: route.params.projectId,
        challengeId: route.params.challengeId,
      },
    } as RouteLocationRaw,
  },
  'Quiz',
])

const { executeMutation: createQuiz } = useCreateQuizMutation()
const { executeMutation: updateQuiz } = useUpdateQuizMutation()
const { executeMutation: addQuizQuestion } = useAddQuizQuestionMutation()
const { executeMutation: updateQuizQuestion } = useUpdateQuizQuestionMutation()
const { executeMutation: deleteQuizQuestion } = useDeleteQuizQuestionMutation()
const { executeMutation: reorderQuizQuestions } =
  useReorderQuizQuestionsMutation()

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
              translationStatus: a.translationStatus,
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
      bettingEnabled: q.bettingEnabled ?? undefined,
      bettingMinPercentage: q.bettingMinPercentage ?? undefined,
      bettingMaxPercentage: q.bettingMaxPercentage ?? undefined,
      bettingMinAbsolute: q.bettingMinAbsolute ?? undefined,
      bettingMaxAbsolute: q.bettingMaxAbsolute ?? undefined,
      translationStatus: q.translationStatus,
    })),
  }
})

const isNewQuiz = computed(() => !quizData.value?.id)

const saving = ref(false)
const formDirty = ref(false)

/** Shared by add and update; the two inputs take the same shape. */
function questionInput(question: QuizQuestionFormData) {
  return {
    questionText: question.questionText,
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
    bettingEnabled: question.bettingEnabled,
    bettingMinPercentage: question.bettingMinPercentage,
    bettingMaxPercentage: question.bettingMaxPercentage,
    bettingMinAbsolute: question.bettingMinAbsolute,
    bettingMaxAbsolute: question.bettingMaxAbsolute,
  }
}

async function saveQuizSettings(form: QuizFormData) {
  const input = {
    name: form.name,
    description: form.description,
    image: form.image,
    timeoutSeconds: form.timeoutSeconds,
    randomizeQuestions: form.randomizeQuestions,
    revealCorrectAnswers: form.revealCorrectAnswers,
    allowRetakes: form.allowRetakes,
    completionPoints: form.completionPoints,
  }

  if (form.id) {
    const result = await updateQuiz({ id: form.id, input })
    return { id: form.id, error: result.error?.message }
  }

  const result = await createQuiz({
    input: {
      projectId: route.params.projectId,
      challengeId: route.params.challengeId,
      ...input,
    },
  })
  return {
    id: result.data?.createQuiz.id,
    error: result.error?.message ?? (result.data ? undefined : 'Ukjent feil'),
  }
}

/**
 * Questions are saved in parallel and every failure is collected. The previous
 * version awaited one mutation per question in a loop and ignored each result,
 * so a quiz could half-save and still report success.
 *
 * Updates deliberately omit `questionOrder`: positions are unique per quiz, so
 * two questions swapping would collide. Order is applied once at the end by
 * `reorderQuizQuestions`, which parks every row on a negative order first.
 */
async function saveQuestions(quizId: string, form: QuizFormData) {
  const existing = quizData.value?.questions ?? []
  const plan = planQuizQuestionSave(existing, form.questions)
  const errors: string[] = []

  const deletions = plan.deletes.map(async (id) => {
    const result = await deleteQuizQuestion({ id })
    if (result.error)
      errors.push(`Kunne ikke slette spørsmål: ${result.error.message}`)
  })

  const updates = plan.updates.map(async ({ id, question }) => {
    const original = existing.find((q) => q.id === id)
    const result = await updateQuizQuestion({
      id,
      input: {
        ...questionInput(question),
        ...bettingClearFlags(original, question),
      },
    })
    if (result.error) {
      errors.push(`«${question.questionText}»: ${result.error.message}`)
    }
  })

  const addedIds = new Map<number, string>()
  const additions = plan.adds.map(async ({ question, order, position }) => {
    const result = await addQuizQuestion({
      quizId,
      input: {
        questionType: question.questionType,
        questionOrder: order,
        ...questionInput(question),
      },
    })
    const id = result.data?.addQuizQuestion.id
    if (id) addedIds.set(position, id)
    else {
      errors.push(
        `«${question.questionText}»: ${result.error?.message ?? 'kunne ikke legges til'}`,
      )
    }
  })

  await Promise.all([...deletions, ...updates, ...additions])

  // Skipped when something failed: ordering a list that is missing a question
  // would silently shuffle the rest.
  if (!errors.length) {
    const ids = orderedQuestionIds(form.questions, addedIds)
    if (ids.length) {
      const result = await reorderQuizQuestions({ quizId, questionIds: ids })
      if (result.error) {
        errors.push(`Kunne ikke lagre rekkefølgen: ${result.error.message}`)
      }
    }
  }

  return errors
}

async function saveQuiz(quizFormData: QuizFormData) {
  saving.value = true
  try {
    const quiz = await saveQuizSettings(quizFormData)
    if (!quiz.id || quiz.error) {
      toast.add({
        title: 'Kunne ikke lagre quizen',
        description: quiz.error,
        color: 'error',
      })
      return
    }

    const errors = await saveQuestions(quiz.id, quizFormData)
    await refetchQuiz({ requestPolicy: 'network-only' })

    // Stay on the page when anything failed: the form still holds what was
    // meant to be saved, and leaving would lose it.
    if (errors.length) {
      toast.add({
        title: `${errors.length} spørsmål ble ikke lagret`,
        description: errors.join('\n'),
        color: 'error',
      })
      return
    }

    toast.add({
      title: isNewQuiz.value ? 'Quiz opprettet' : 'Quiz oppdatert',
      color: 'success',
    })
    formDirty.value = false

    navigateTo({
      name: 'admin-projects-projectId-challenges-challengeId',
      params: {
        projectId: route.params.projectId,
        challengeId: route.params.challengeId,
      },
    })
  } finally {
    saving.value = false
  }
}

// Check if challenge is a quiz challenge
const isQuizChallenge = computed(() => {
  return data.value?.challenge.__typename === 'QuizChallenge'
})
</script>

<template>
  <div>
    <div>
      <AdminQueryState :fetching :error>
        <template v-if="data">
          <div v-if="!isQuizChallenge" class="text-center py-12">
            <p class="text-muted">
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
              v-model:dirty="formDirty"
              :saving
              :quiz-data="quizData"
              :translation-status="
                data?.challenge.__typename === 'QuizChallenge'
                  ? (data.challenge.quiz?.translationStatus ?? [])
                  : []
              "
              :project-id="route.params.projectId"
              :challenge-id="route.params.challengeId"
              @save="saveQuiz"
            />
          </template>
        </template>
      </AdminQueryState>
    </div>
  </div>
</template>
