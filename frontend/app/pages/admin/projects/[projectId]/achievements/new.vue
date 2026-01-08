<script setup lang="ts">
import type { AchievementFormData } from '~/components/admin/achievement/AdminAchievementForm.vue'

definePageMeta({
  layout: 'admin',
  middleware: 'admin',
})

gql(`
  query AdminProjectAchievementsNewPage($projectId: ID!) {
    project(id: $projectId) {
      id
      name
      branding {
        colors {
          ...BrandingColorsFields
        }
      }
    }
  }
`)

const route = useRoute('admin-projects-projectId-achievements-new')
const toast = useToast()

const { isAuthReady } = useAuthReady()
const { data } = useAdminProjectAchievementsNewPageQuery({
  variables: {
    projectId: route.params.projectId,
  },
  pause: computed(() => !isAuthReady.value),
})

const { executeMutation: createSimple } = useCreateSimpleAchievementMutation()
const { executeMutation: createContent } = useCreateContentAchievementMutation()
const { executeMutation: createStreak } = useCreateStreakAchievementMutation()
const { executeMutation: createQuiz } = useCreateQuizAchievementMutation()

async function handleSubmit(formData: AchievementFormData) {
  let response

  const baseInput = {
    name: formData.name,
    descriptionPending: formData.descriptionPending,
    descriptionCompleted: formData.descriptionCompleted,
    notificationText: formData.notificationText,
    imagePending: formData.imagePending ?? '',
    imageCompleted: formData.imageCompleted ?? '',
    points: formData.points,
    hidden: formData.hidden,
    projectId: route.params.projectId,
  }

  switch (formData.achievementType) {
    case 'SIMPLE':
      response = await createSimple({
        input: baseInput,
      })
      break

    case 'CONTENT':
      response = await createContent({
        input: {
          ...baseInput,
          items:
            formData.items?.map((item) => ({
              externalContentId: item.externalContent.id,
            })) ?? [],
        },
      })
      break

    case 'STREAK':
      response = await createStreak({
        input: {
          ...baseInput,
          streakId: formData.streakId!,
          neededStreak: formData.neededStreak!,
        },
      })
      break

    case 'QUIZ':
      response = await createQuiz({
        input: {
          ...baseInput,
          quizId: formData.quizId!,
          minScorePercentage: formData.minScorePercentage,
          requireCompletion: formData.requireCompletion ?? true,
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
    description: 'Utmerkelse opprettet',
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
              label: 'Utmerkelser',
            },
            {
              label: 'Ny',
            },
          ]"
        />
      </UContainer>
    </div>
    <UContainer class="py-12">
      <h1 class="mb-6 text-2xl font-bold">Opprett utmerkelse</h1>
      <AdminAchievementForm
        :project-id="route.params.projectId"
        :colors="data?.project.branding.colors"
        submit-label="Opprett utmerkelse"
        @submit="handleSubmit"
      />
    </UContainer>
  </div>
</template>
