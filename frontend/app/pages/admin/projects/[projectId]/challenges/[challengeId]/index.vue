<script setup lang="ts">
import type { ChallengeFormData } from '~/components/admin/challenge/AdminChallengeForm.vue'

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
      notificationText
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

const initialData = computed(() => {
  if (!data.value) return undefined
  const c = data.value.challenge
  return {
    type: getChallengeType(c.__typename ?? 'SimpleChallenge'),
    name: c.name,
    description: c.description ?? undefined,
    image: c.image ?? undefined,
    url: c.__typename === 'ExternalChallenge' ? c.url : undefined,
    buttonText: c.buttonText ?? '',
    notificationText: c.notificationText ?? undefined,
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

async function handleSubmit(formData: ChallengeFormData) {
  const {
    type,
    allowSelfCompletion,
    url,
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
        :project-id="route.params.projectId"
        :challenge-id="route.params.challengeId"
        :colors="data?.challenge.project.branding.colors"
        submit-label="Lagre endringer"
        is-edit-mode
        :on-delete="handleDelete"
        @submit="handleSubmit"
      />
    </UContainer>
  </div>
</template>
