<script setup lang="ts">
import type { ChallengeFormData } from '../../../../../../components/admin/challenge/AdminChallengeForm.vue'

definePageMeta({
  permission: 'projects:view',
  layout: 'admin',
})

gql(`
  query AdminProjectChallengePage($challengeId: ID!, $projectId: ID!) {
    # The denominator for the completion count.
    participants: users(first: 0, filter: { projectId: $projectId }) {
      totalCount
    }
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
      completionCount
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
      ... on QuizChallenge {
        quiz {
          id
          completionPoints
          allowRetakes
          randomizeQuestions
          revealCorrectAnswers
          timeoutSeconds
          questions {
            id
          }
        }
      }
      translationStatus {
        ...TranslationStatus
      }
    }
  }
`)

const route = useRoute('admin-projects-projectId-challenges-challengeId')
const toast = useToast()
const { confirm } = useConfirm()

const { isAuthReady } = useAuthReady()
const { data, fetching, error } = useAdminProjectChallengePageQuery({
  variables: computed(() => ({
    challengeId: route.params.challengeId,
    projectId: route.params.projectId,
  })),
  pause: computed(() => !isAuthReady.value),
})

// Supplies the trailing breadcrumb crumb and the navbar title; everything
// above it is derived from the route.
useAdminPage(() => data.value?.challenge.name)

const { executeMutation } = useUpdateChallengeMutation()
const { executeMutation: executeDelete } = useDeleteChallengeMutation()

const challengeTypeLabels: Record<string, string> = {
  SimpleChallenge: 'Enkel',
  ExternalChallenge: 'Ekstern',
  QuizChallenge: 'Quiz',
  PluginChallenge: 'Plugin',
}

const isPublished = computed(() => {
  const publishedAt = data.value?.challenge.publishedAt
  return !!publishedAt && new Date(publishedAt) <= new Date()
})

const quiz = computed(() =>
  data.value?.challenge.__typename === 'QuizChallenge'
    ? data.value.challenge.quiz
    : null,
)

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
  const confirmed = await confirm({
    title: `Slette "${data.value?.challenge.name}"?`,
    description: 'Denne handlingen kan ikke angres.',
    icon: 'lucide:triangle-alert',
  })

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
  <AdminQueryState :fetching :error>
    <div v-if="initialData && data" class="space-y-8">
      <div class="flex flex-wrap items-start justify-between gap-4">
        <div class="min-w-0">
          <div class="flex flex-wrap items-center gap-3">
            <h1 class="text-3xl font-bold">{{ data.challenge.name }}</h1>
            <UBadge variant="subtle">
              {{
                challengeTypeLabels[data.challenge.__typename ?? ''] ??
                'Utfordring'
              }}
            </UBadge>
            <UBadge v-if="!isPublished" variant="soft" color="warning">
              Ikke publisert
            </UBadge>
          </div>
          <!-- Completion belongs to the header, not a section: it is status,
               and this page is for editing. -->
          <p class="text-muted mt-1 text-sm">
            {{ formatNumber(data.challenge.completionCount) }} av
            {{ formatNumber(data.participants.totalCount) }} deltakere har
            fullført
          </p>
        </div>

        <div class="flex shrink-0 flex-wrap gap-2">
          <UButton
            v-if="data.challenge.__typename === 'QuizChallenge'"
            variant="soft"
            icon="lucide:list-checks"
            :to="{
              name: 'admin-projects-projectId-challenges-challengeId-sessions',
              params: {
                projectId: route.params.projectId,
                challengeId: route.params.challengeId,
              },
            }"
          >
            Sesjoner
          </UButton>
          <AdminChallengeQrModal
            :challenge-id="route.params.challengeId"
            :challenge-name="data.challenge.name"
          />
          <UButton variant="soft" color="error" @click="handleDelete">
            Slett
          </UButton>
        </div>
      </div>

      <AdminChallengeForm
        :initial-data="initialData"
        :project-id="route.params.projectId"
        :challenge-id="route.params.challengeId"
        :colors="data.challenge.project.branding.colors"
        :translation-status="data.challenge.translationStatus ?? []"
        submit-label="Lagre endringer"
        is-edit-mode
        @submit="handleSubmit"
      />

      <AdminChallengeQuizSection
        v-if="data.challenge.__typename === 'QuizChallenge'"
        :project-id="route.params.projectId"
        :challenge-id="route.params.challengeId"
        :quiz="quiz"
      />
    </div>
  </AdminQueryState>
</template>
