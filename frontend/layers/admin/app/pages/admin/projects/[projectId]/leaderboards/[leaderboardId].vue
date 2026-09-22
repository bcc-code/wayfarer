<script setup lang="ts">
import type { LeaderboardConfigFormData } from '../../../../../components/admin/leaderboard/AdminLeaderboardConfigForm.vue'

definePageMeta({
  permission: 'projects:view',
  layout: 'admin',
})

gql(`
  query AdminProjectLeaderboardPage($id: ID!) {
    leaderboardConfig(id: $id) {
      ...LeaderboardConfigFields
    }
  }
`)

const route = useRoute('admin-projects-projectId-leaderboards-leaderboardId')

const { isAuthReady } = useAuthReady()
const { data, fetching, error } = useAdminProjectLeaderboardPageQuery({
  variables: computed(() => ({ id: route.params.leaderboardId })),
  pause: computed(() => !isAuthReady.value),
})

const config = computed(() => data.value?.leaderboardConfig)

// Supplies the trailing breadcrumb crumb and the navbar title; everything
// above it is derived from the route.
useAdminPage(() => config.value?.name)

const toast = useToast()
const { confirm } = useConfirm()
const { executeMutation: executeUpdate } = useUpdateLeaderboardConfigMutation()
const { executeMutation: executeDelete } = useDeleteLeaderboardConfigMutation()

const listRoute = computed(() => ({
  name: 'admin-projects-projectId-leaderboards' as const,
  params: { projectId: route.params.projectId },
}))

/**
 * Every field is resent, including `filter: null` to clear one:
 * `UpdateLeaderboardConfigInput` is full-replace, not a patch. `eventId` is
 * absent from that input entirely — a config's event scope is fixed when it is
 * created.
 */
async function handleSubmit(formData: LeaderboardConfigFormData) {
  const response = await executeUpdate({
    id: route.params.leaderboardId,
    input: {
      name: formData.name,
      entityType: formData.entityType,
      filter: formData.filter,
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
    description: 'Ledertavle oppdatert',
    color: 'success',
  })
  navigateTo(listRoute.value)
}

async function handleDelete() {
  const confirmed = await confirm({
    title: `Slette "${config.value?.name}"?`,
    description: 'Denne handlingen kan ikke angres.',
    icon: 'lucide:triangle-alert',
  })

  if (!confirmed) return

  const response = await executeDelete({ id: route.params.leaderboardId })
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
    description: 'Ledertavle slettet',
    color: 'success',
  })
  navigateTo(listRoute.value)
}
</script>

<template>
  <div class="space-y-8">
    <AdminQueryState :fetching :error>
      <div v-if="config" class="space-y-8">
        <div class="flex flex-wrap items-center justify-between gap-4">
          <h1 class="text-3xl font-bold">Rediger ledertavle</h1>
          <UButton
            color="error"
            variant="soft"
            icon="lucide:trash-2"
            @click="handleDelete"
          >
            Slett
          </UButton>
        </div>
        <AdminLeaderboardConfigForm
          :project-id="route.params.projectId"
          :initial-data="config"
          is-edit-mode
          submit-label="Lagre endringer"
          @submit="handleSubmit"
        />
      </div>
    </AdminQueryState>
  </div>
</template>
