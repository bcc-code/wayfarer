<script setup lang="ts">
import type { LeaderboardConfigFormData } from '../../../../../components/admin/leaderboard/AdminLeaderboardConfigForm.vue'

definePageMeta({
  permission: 'projects:view',
  layout: 'admin',
})

// Trailing breadcrumb crumbs; the path above them is derived from the route.
useAdminPage(() => 'Ny')

const route = useRoute('admin-projects-projectId-leaderboards-new')
const toast = useToast()
const { executeMutation } = useCreateLeaderboardConfigMutation()

async function handleSubmit(formData: LeaderboardConfigFormData) {
  const response = await executeMutation({
    input: {
      projectId: route.params.projectId,
      eventId: formData.eventId,
      name: formData.name,
      entityType: formData.entityType,
      filter: formData.filter,
      maxEntries: formData.maxEntries,
      sortOrder: formData.sortOrder,
      isActive: formData.isActive,
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
    title: 'Suksess',
    description: 'Ledertavle opprettet',
    color: 'success',
  })
  navigateTo({
    name: 'admin-projects-projectId-leaderboards',
    params: { projectId: route.params.projectId },
  })
}
</script>

<template>
  <div class="space-y-8">
    <h1 class="text-3xl font-bold">Opprett ledertavle</h1>
    <AdminLeaderboardConfigForm
      :project-id="route.params.projectId"
      submit-label="Opprett ledertavle"
      @submit="handleSubmit"
    />
  </div>
</template>
