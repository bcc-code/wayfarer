<script setup lang="ts">
import type { AchievementFormData } from '~/components/admin/achievement/AdminAchievementForm.vue'

definePageMeta({
  layout: 'admin',
  middleware: 'admin',
})

gql(`
  query AdminProjectAchievementPage($achievementId: ID!) {
    achievement(id: $achievementId) {
      __typename
      id
      name
      descriptionPending
      descriptionCompleted
      imagePendingObject {
        ...ImageFields
      }
      imageCompletedObject {
        ...ImageFields
      }
      notificationText
      achievedAt
      points
      hidden
      awardableFrom
      ... on ContentAchievement {
        items {
          id
          sortOrder
          externalContent {
            id
            planId
            taskId
            contentId
            contentType
            publishedAt
            source
            syncedAt
            createdAt
            updatedAt
            title
            translations {
              languageCode
              title
            }
          }
        }
      }
      ... on StreakAchievement {
        neededStreak
        streak {
          id
          name
          description
        }
      }
      ... on QuizAchievement {
        quiz {
          id
          name
        }
        minScorePercentage
        requireCompletion
      }
      project {
        id
        name
        branding {
          colors {
            ...BrandingColorsFields
          }
        }
      }
    }
  }
`)

const route = useRoute('admin-projects-projectId-achievements-achievementId')
const toast = useToast()

const { isAuthReady } = useAuthReady()
const { data, fetching, error } = useAdminProjectAchievementPageQuery({
  variables: {
    achievementId: route.params.achievementId,
  },
  pause: computed(() => !isAuthReady.value),
})

const { executeMutation: updateSimple } = useUpdateAchievementMutation()
const { executeMutation: updateContent } = useUpdateContentAchievementMutation()
const { executeMutation: updateStreak } = useUpdateStreakAchievementMutation()
const { executeMutation: updateQuiz } = useUpdateQuizAchievementMutation()
const { executeMutation: executeDelete } = useDeleteAchievementMutation()

type AchievementType = 'SIMPLE' | 'CONTENT' | 'STREAK' | 'QUIZ'

const achievementType = computed<AchievementType>(() => {
  const typename = data.value?.achievement.__typename
  switch (typename) {
    case 'ContentAchievement':
      return 'CONTENT'
    case 'StreakAchievement':
      return 'STREAK'
    case 'QuizAchievement':
      return 'QUIZ'
    default:
      return 'SIMPLE'
  }
})

const initialData = computed(() => {
  if (!data.value) return undefined
  const a = data.value.achievement

  const base = {
    name: a.name,
    descriptionPending: a.descriptionPending,
    descriptionCompleted: a.descriptionCompleted,
    notificationText: a.notificationText,
    imagePending: a.imagePendingObject?.url ?? '',
    imageCompleted: a.imageCompletedObject?.url ?? '',
    points: a.points,
    hidden: a.hidden,
    awardableFrom: toLocalDatetimeLocal(a.awardableFrom),
  }

  // Add type-specific fields
  if (a.__typename === 'ContentAchievement') {
    return {
      ...base,
      items: a.items.map((item) => ({
        id: item.id,
        externalContent: {
          id: item.externalContent.id,
          title: item.externalContent.title,
          contentType: item.externalContent.contentType,
          source: item.externalContent.source,
          publishedAt: item.externalContent.publishedAt,
        },
      })),
    }
  }

  if (a.__typename === 'StreakAchievement') {
    return {
      ...base,
      streakId: a.streak.id,
      neededStreak: a.neededStreak,
    }
  }

  if (a.__typename === 'QuizAchievement') {
    return {
      ...base,
      quizId: a.quiz?.id,
      minScorePercentage: a.minScorePercentage ?? undefined,
      requireCompletion: a.requireCompletion,
    }
  }

  return base
})

async function handleSubmit(formData: AchievementFormData) {
  let response

  const baseInput = {
    name: formData.name,
    descriptionPending: formData.descriptionPending,
    descriptionCompleted: formData.descriptionCompleted,
    notificationText: formData.notificationText,
    imagePending: formData.imagePending,
    imageCompleted: formData.imageCompleted,
    points: formData.points,
    hidden: formData.hidden,
    awardableFrom: toISOString(formData.awardableFrom),
  }

  switch (formData.achievementType) {
    case 'SIMPLE':
      response = await updateSimple({
        id: route.params.achievementId,
        input: baseInput,
      })
      break

    case 'CONTENT':
      console.log(baseInput)
      response = await updateContent({
        id: route.params.achievementId,
        input: {
          ...baseInput,
          items: formData.items?.map((item) => ({
            externalContentId: item.externalContent.id,
          })),
        },
      })
      break

    case 'STREAK':
      response = await updateStreak({
        id: route.params.achievementId,
        input: {
          ...baseInput,
          streakId: formData.streakId,
          neededStreak: formData.neededStreak,
        },
      })
      break

    case 'QUIZ':
      response = await updateQuiz({
        id: route.params.achievementId,
        input: {
          ...baseInput,
          quizId: formData.quizId,
          minScorePercentage: formData.minScorePercentage,
          requireCompletion: formData.requireCompletion,
        },
      })
      break
  }

  if (response?.error) {
    toast.add({
      title: response.error.name,
      description: response.error.message,
      color: 'error',
    })
    return
  }

  toast.add({
    title: 'Suksess',
    description: 'Utmerkelse oppdatert',
    color: 'success',
  })
  navigateTo({
    name: 'admin-projects-projectId',
    params: { projectId: route.params.projectId },
  })
}

async function handleDelete() {
  const confirmed = confirm(
    `Er du sikker på at du vil slette "${data.value?.achievement.name}"? Denne handlingen kan ikke angres.`,
  )

  if (!confirmed) return

  const response = await executeDelete({ id: route.params.achievementId })
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
    description: 'Utmerkelse slettet',
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
              label: data?.achievement.project.name ?? route.params.projectId,
              to: {
                name: 'admin-projects-projectId',
                params: { projectId: route.params.projectId },
              },
            },
            {
              label: 'Utmerkelser',
            },
            {
              label: data?.achievement.name ?? route.params.achievementId,
              to: {
                name: 'admin-projects-projectId-achievements-achievementId',
                params: {
                  projectId: route.params.projectId,
                  achievementId: route.params.achievementId,
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
      <template v-else-if="initialData">
        <h1 class="mb-6 text-2xl font-bold">Rediger utmerkelse</h1>
        <AdminAchievementForm
          :project-id="route.params.projectId"
          :initial-data="initialData"
          :achievement-type="achievementType"
          :is-edit-mode="true"
          :colors="data?.achievement.project.branding.colors"
          submit-label="Lagre endringer"
          :on-delete="handleDelete"
          @submit="handleSubmit"
        />
      </template>
    </UContainer>
  </div>
</template>
