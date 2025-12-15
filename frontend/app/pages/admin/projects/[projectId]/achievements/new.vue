<script setup lang="ts">
definePageMeta({
  layout: 'admin',
  middleware: 'admin',
})

gql(`
  query AdminProjectAchievementsNewPage($projectId: ID!) {
    project(id: $projectId) {
      id
      name
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

const { executeMutation } = useCreateSimpleAchievementMutation()

async function handleSubmit(formData: {
  name: string
  descriptionPending: string
  descriptionCompleted: string
  notificationText: string
  imagePending?: string
  imageCompleted?: string
  points: number
  hidden: boolean
}) {
  const response = await executeMutation({
    input: {
      name: formData.name,
      descriptionPending: formData.descriptionPending,
      descriptionCompleted: formData.descriptionCompleted,
      notificationText: formData.notificationText,
      imagePending: formData.imagePending ?? '',
      imageCompleted: formData.imageCompleted ?? '',
      points: formData.points,
      hidden: formData.hidden,
      projectId: route.params.projectId,
    },
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
    description: 'Achievement created successfully',
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
              label: data?.project.name ?? route.params.projectId,
              to: {
                name: 'admin-projects-projectId',
                params: { projectId: route.params.projectId },
              },
            },
            {
              label: 'Achievements',
            },
            {
              label: 'New',
            },
          ]"
        />
      </UContainer>
    </div>
    <UContainer class="py-12">
      <h1 class="mb-6 text-2xl font-bold">Create Achievement</h1>
      <AdminAchievementForm
        submit-label="Create Achievement"
        @submit="handleSubmit"
      />
    </UContainer>
  </div>
</template>
