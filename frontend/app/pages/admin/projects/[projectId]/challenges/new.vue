<script setup lang="ts">
import type { ChallengeFormData } from '~/components/admin/challenge/AdminChallengeForm.vue'

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

const eventId = ref('')

const eventOptions = computed(() => {
  return (
    data.value?.events.edges.map((e) => ({
      value: e.node.id,
      label: e.node.name,
    })) ?? []
  )
})

async function handleSubmit(formData: ChallengeFormData) {
  const { type, allowSelfCompletion, url, publishedAt, pluginChallengeId, ...rest } =
    formData

  // Only include type-specific fields
  const input = {
    ...rest,
    type,
    publishedAt: publishedAt ? toISOString(publishedAt) : undefined,
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

  const challengeId = response.data?.createChallenge.id

  // For quiz challenges, create an empty quiz immediately (required by schema)
  if (type === ChallengeType.Quiz && challengeId) {
    const quizResult = await createQuiz({
      input: {
        projectId: route.params.projectId,
        challengeId,
        name: formData.name,
        description: '',
        randomizeQuestions: false,
        revealCorrectAnswers: true,
        allowRetakes: false,
        completionPoints: 0,
      },
    })

    if (quizResult.error) {
      toast.add({
        title: 'Feil',
        description: quizResult.error.message,
        color: 'error',
      })
      return
    }

    toast.add({
      title: 'Suksess',
      description: 'Utfordring opprettet',
      color: 'success',
    })

    navigateTo({
      name: 'admin-projects-projectId-challenges-challengeId-quiz',
      params: {
        projectId: route.params.projectId,
        challengeId,
      },
    })
  } else {
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
