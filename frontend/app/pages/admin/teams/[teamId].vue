<script setup lang="ts">
definePageMeta({
  layout: 'admin',
  middleware: 'admin',
})

gql(`
  query AdminTeamPage($id: ID!) {
    team(id: $id) {
      id
      name
      description
      joinCode
      members {
        id
        name
        isTeamLead
        joinedAt
        user {
          id
          email
          image
        }
        church {
          id
          name
        }
      }
      parentProject {
        id
        name
      }
      superTeam {
        id
        name
      }
    }
  }
`)

const route = useRoute('admin-teams-teamId')

const { canManageTeam } = usePermissions()
const canEdit = computed(() => canManageTeam(route.params.teamId))

const { isAuthReady } = useAuthReady()
const {
  data,
  fetching,
  error,
  executeQuery: refetch,
} = useAdminTeamPageQuery({
  variables: {
    id: route.params.teamId,
  },
  pause: computed(() => !isAuthReady.value),
})

const { executeMutation: updateTeam } = useUpdateTeamMutation()
const { executeMutation: removeTeamMembers } = useRemoveTeamMembersMutation()
const { executeMutation: regenerateJoinCode } = useRegenerateJoinCodeMutation()
const { executeMutation: assignTeamLead } = useAssignTeamLeadMutation()
const { executeMutation: deleteTeam } = useDeleteTeamMutation()
const toast = useToast()

// Edit mode state
const isEditing = ref(false)
const editState = reactive({
  name: '',
  description: '',
})

function startEditing() {
  if (data.value) {
    editState.name = data.value.team.name
    editState.description = data.value.team.description
    isEditing.value = true
  }
}

function cancelEditing() {
  isEditing.value = false
}

async function saveChanges() {
  const result = await updateTeam({
    id: route.params.teamId,
    input: {
      name: editState.name,
      description: editState.description,
    },
  })

  if (result.error) {
    toast.add({
      title: 'Failed to update team',
      description: result.error.message,
      color: 'error',
    })
    return
  }

  toast.add({
    title: 'Team updated',
    color: 'success',
  })

  isEditing.value = false
  refetch({ requestPolicy: 'network-only' })
}

async function handleRemoveMember(userId: string, userName: string) {
  const result = await removeTeamMembers({
    teamId: route.params.teamId,
    userIds: [userId],
  })

  if (result.error) {
    toast.add({
      title: 'Failed to remove member',
      description: result.error.message,
      color: 'error',
    })
    return
  }

  toast.add({
    title: 'Member removed',
    description: `${userName} has been removed from the team`,
    color: 'success',
  })

  refetch({ requestPolicy: 'network-only' })
}

async function handleRegenerateJoinCode() {
  const result = await regenerateJoinCode({
    teamId: route.params.teamId,
  })

  if (result.error) {
    toast.add({
      title: 'Failed to regenerate join code',
      description: result.error.message,
      color: 'error',
    })
    return
  }

  toast.add({
    title: 'Join code regenerated',
    color: 'success',
  })

  refetch({ requestPolicy: 'network-only' })
}

async function handleAssignTeamLead(userId: string, userName: string) {
  const result = await assignTeamLead({
    teamId: route.params.teamId,
    userId,
  })

  if (result.error) {
    toast.add({
      title: 'Failed to assign team lead',
      description: result.error.message,
      color: 'error',
    })
    return
  }

  toast.add({
    title: 'Team lead assigned',
    description: `${userName} is now the team lead`,
    color: 'success',
  })

  await refetch({ requestPolicy: 'network-only' })
}

async function handleDeleteTeam() {
  const result = await deleteTeam({
    id: route.params.teamId,
  })

  if (result.error) {
    toast.add({
      title: 'Failed to delete team',
      description: result.error.message,
      color: 'error',
    })
    return
  }

  toast.add({
    title: 'Team deleted',
    color: 'success',
  })

  navigateTo({ name: 'admin-teams' })
}

function copyJoinCode() {
  if (data.value) {
    navigator.clipboard.writeText(data.value.team.joinCode)
    toast.add({
      title: 'Join code copied',
      color: 'success',
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
            { label: 'Teams', to: { name: 'admin-teams' } },
            {
              label: data?.team.name ?? route.params.teamId,
              to: {
                name: 'admin-teams-teamId',
                params: { teamId: route.params.teamId },
              },
            },
          ]"
        />
      </UContainer>
    </div>
    <UContainer class="py-12">
      <LoadingState v-if="fetching" />
      <ErrorState v-else-if="error" :error />
      <div v-else-if="data" class="space-y-6">
        <!-- Team Header -->
        <div class="flex items-start justify-between">
          <div>
            <h1 class="text-3xl font-bold">{{ data.team.name }}</h1>
            <p class="text-dimmed">{{ data.team.description }}</p>
          </div>
          <div v-if="canEdit" class="flex gap-2">
            <UButton v-if="!isEditing" variant="soft" @click="startEditing">
              Edit
            </UButton>
            <UButton variant="soft" color="error" @click="handleDeleteTeam">
              Delete
            </UButton>
          </div>
        </div>

        <!-- Edit Form -->
        <UCard v-if="isEditing">
          <template #header>
            <h2 class="text-xl font-semibold">Edit Team</h2>
          </template>
          <div class="space-y-4">
            <UFormField label="Name">
              <UInput v-model="editState.name" class="w-full" />
            </UFormField>
            <UFormField label="Description">
              <UTextarea
                v-model="editState.description"
                class="w-full"
                autoresize
              />
            </UFormField>
          </div>
          <template #footer>
            <div class="flex justify-end gap-3">
              <UButton variant="ghost" @click="cancelEditing">Cancel</UButton>
              <UButton @click="saveChanges">Save Changes</UButton>
            </div>
          </template>
        </UCard>

        <!-- Team Info -->
        <dl class="text-sm">
          <div class="border-default flex gap-6 border-b py-2">
            <dt class="text-muted w-24 shrink-0">Team ID</dt>
            <dd class="font-mono">{{ data.team.id }}</dd>
          </div>
          <div class="border-default flex gap-6 border-b py-2">
            <dt class="text-muted w-24 shrink-0">Project</dt>
            <dd class="font-medium">{{ data.team.parentProject.name }}</dd>
          </div>
          <div class="border-default flex gap-6 border-b py-2">
            <dt class="text-muted w-24 shrink-0">Super Team</dt>
            <dd v-if="data.team.superTeam" class="font-medium">
              {{ data.team.superTeam.name }}
            </dd>
            <dd v-else class="text-muted">None</dd>
          </div>
          <div class="border-default flex gap-6 border-b py-2">
            <dt class="text-muted w-24 shrink-0">Members</dt>
            <dd class="font-medium">{{ data.team.members.length }}</dd>
          </div>
          <div class="flex items-center gap-6 py-2">
            <dt class="text-muted w-24 shrink-0">Join Code</dt>
            <dd class="flex items-center gap-2">
              <code class="bg-background-indent rounded px-2 py-1">{{
                data.team.joinCode
              }}</code>
              <UButton
                variant="ghost"
                size="xs"
                icon="i-lucide-copy"
                @click="copyJoinCode"
              />
              <UButton
                v-if="canEdit"
                variant="ghost"
                size="xs"
                icon="i-lucide-refresh-cw"
                @click="handleRegenerateJoinCode"
              />
            </dd>
          </div>
        </dl>

        <!-- Members Card -->
        <UCard>
          <template #header>
            <div class="flex items-center justify-between">
              <h2 class="text-xl font-semibold">
                Members ({{ data.team.members.length }})
              </h2>
            </div>
          </template>

          <div v-if="data.team.members.length > 0" class="space-y-2">
            <div
              v-for="member in data.team.members"
              :key="member.id"
              class="border-default flex items-center justify-between rounded-md border p-3"
            >
              <div class="flex items-center gap-3">
                <UAvatar
                  :src="member.user.image ?? ''"
                  :text="getInitials(member.name)"
                  size="sm"
                />
                <div>
                  <div class="flex items-center gap-2">
                    <span class="font-medium">{{ member.name }}</span>
                    <UBadge
                      v-if="member.isTeamLead"
                      variant="soft"
                      size="xs"
                      color="primary"
                    >
                      Lead
                    </UBadge>
                  </div>
                  <div class="text-muted text-xs">
                    {{ member.church.name }}
                  </div>
                </div>
              </div>
              <div v-if="canEdit" class="flex items-center gap-2">
                <UButton
                  v-if="!member.isTeamLead"
                  variant="ghost"
                  size="xs"
                  @click="handleAssignTeamLead(member.user.id, member.name)"
                >
                  Make Lead
                </UButton>
                <UButton
                  icon="i-lucide-trash-2"
                  color="error"
                  variant="ghost"
                  size="sm"
                  @click="handleRemoveMember(member.user.id, member.name)"
                />
              </div>
            </div>
          </div>
          <div v-else class="text-dimmed">No members</div>
        </UCard>
      </div>
    </UContainer>
  </div>
</template>
