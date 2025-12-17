<script setup lang="ts">
definePageMeta({
  layout: 'admin',
  middleware: 'admin',
})

gql(`
  query AdminProjectAchievementPage($achievementId: ID!) {
    achievement(id: $achievementId) {
      id
      name
      descriptionPending
      descriptionCompleted
      imagePending
      imageCompleted
      notificationText
      achievedAt
      points
      hidden
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

const { executeMutation } = useUpdateAchievementMutation()
const { executeMutation: executeDelete } = useDeleteAchievementMutation()

const initialData = computed(() => {
  if (!data.value) return undefined
  const a = data.value.achievement
  return {
    name: a.name,
    descriptionPending: a.descriptionPending,
    descriptionCompleted: a.descriptionCompleted,
    notificationText: a.notificationText,
    imagePending: a.imagePending,
    imageCompleted: a.imageCompleted,
    points: a.points,
    hidden: a.hidden,
  }
})

async function handleSubmit(formData: {
  name: string
  descriptionPending: string
  descriptionCompleted: string
  imagePending?: string
  imageCompleted?: string
  points: number
  hidden: boolean
}) {
  const response = await executeMutation({
    id: route.params.achievementId,
    input: formData,
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
    title: 'Success',
    description: 'Achievement updated successfully',
    color: 'success',
  })
  navigateTo({
    name: 'admin-projects-projectId',
    params: { projectId: route.params.projectId },
  })
}

async function handleDelete() {
  const confirmed = confirm(
    `Are you sure you want to delete "${data.value?.achievement.name}"? This action cannot be undone.`,
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
    title: 'Success',
    description: 'Achievement deleted successfully',
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
              label: 'Projects',
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
              label: 'Achievements',
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
      <AdminAchievementForm
        v-else-if="initialData"
        :initial-data="initialData"
        :colors="data?.achievement.project.branding.colors"
        submit-label="Save changes"
        :on-delete="handleDelete"
        @submit="handleSubmit"
      />
    </UContainer>
  </div>
</template>
