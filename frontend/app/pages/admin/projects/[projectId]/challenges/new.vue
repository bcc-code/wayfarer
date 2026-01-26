<script setup lang="ts">
import type { ChallengeFormData } from '~/components/admin/challenge/AdminChallengeForm.vue'
import type { QuizFormData } from '~/components/admin/quiz/AdminQuizForm.vue'

definePageMeta({
  layout: 'admin',
  middleware: 'admin',
})

gql(`
  query AdminProjectChallengeNewPage($projectId: ID!) {
    project(id: $projectId) {
      id
      name
      branding {
        colors {
          ...BrandingColorsFields
        }
      }
    }
    events(first: 100, filter: { projectId: $projectId }) {
      edges {
        node {
          id
          name
        }
      }
    }
  }
`)

const route = useRoute('admin-projects-projectId-challenges-new')
const toast = useToast()

const { isAuthReady } = useAuthReady()
const { data } = useAdminProjectChallengeNewPageQuery({
  variables: {
    projectId: route.params.projectId,
  },
  pause: computed(() => !isAuthReady.value),
})

const { executeMutation } = useCreateChallengeMutation()
const { executeMutation: createQuiz } = useCreateQuizMutation()
const { executeMutation: addQuizQuestion } = useAddQuizQuestionMutation()

const eventId = ref('')

const eventOptions = computed(() => {
  return (
    data.value?.events.edges.map((e) => ({
      value: e.node.id,
      label: e.node.name,
    })) ?? []
  )
})

async function saveQuiz(quizFormData: QuizFormData, challengeId: string) {
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
        minValue: question.minValue,
        maxValue: question.maxValue,
        stepValue: question.stepValue,
      },
    })
  }
}

async function handleSubmit(formData: ChallengeFormData) {
  const { type, allowSelfCompletion, url, quiz, pluginChallengeId, ...rest } =
    formData

  // Only include type-specific fields
  const input = {
    ...rest,
    type,
    ...(type === ChallengeType.Simple && { allowSelfCompletion }),
    ...(type === ChallengeType.External && { url }),
    ...(type === ChallengeType.Plugin && { pluginChallengeId }),
  }

  const response = await executeMutation({
    projectId: route.params.projectId,
    eventId: eventId.value || undefined,
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
  const challengeId = response.data?.createChallenge.id
  if (type === ChallengeType.Quiz && quiz && challengeId) {
    try {
      await saveQuiz(quiz, challengeId)
    } catch (err) {
      toast.add({
        title: 'Feil',
        description: err instanceof Error ? err.message : 'Kunne ikke lagre quiz',
        color: 'error',
      })
      return
    }
  }

  toast.add({
    title: 'Suksess',
    description: 'Utfordring opprettet',
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
              label: data?.project.name ?? route.params.projectId,
              to: {
                name: 'admin-projects-projectId',
                params: { projectId: route.params.projectId },
              },
            },
            {
              label: 'Utfordringer',
            },
            {
              label: 'Ny',
            },
          ]"
        />
      </UContainer>
    </div>
    <UContainer class="py-12">
      <h1 class="mb-6 text-2xl font-bold">Opprett utfordring</h1>
      <AdminChallengeForm
        :project-id="route.params.projectId"
        :colors="data?.project.branding.colors"
        submit-label="Opprett utfordring"
        @submit="handleSubmit"
      >
        <template #before-type>
          <UFormField name="eventId" label="Arrangement (valgfritt)">
            <USelect
              v-model="eventId"
              :items="eventOptions"
              placeholder="Ingen (prosjekt-nivå)"
              class="w-full"
            />
          </UFormField>
        </template>
      </AdminChallengeForm>
    </UContainer>
  </div>
</template>
