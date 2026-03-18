<script setup lang="ts">
definePageMeta({
  layout: 'admin',
  middleware: 'admin',
})

gql(`
  query AdminSuperTeamDetailPage($id: ID!, $projectId: ID!) {
    superteam(id: $id) {
      id
      name
      description
      color
      imageObject {
        ...ImageFields
      }
      teams {
        id
        name
        description
      }
    }
    teams(first: 200, filter: { projectId: $projectId }) {
      edges {
        node {
          id
          name
          superTeam {
            id
          }
        }
      }
    }
  }
`)

const route = useRoute('admin-projects-projectId-superteams-superTeamId')
const toast = useToast()

const { isAuthReady } = useAuthReady()
const {
  data,
  fetching,
  error,
  executeQuery: refetch,
} = useAdminSuperTeamDetailPageQuery({
  variables: {
    id: route.params.superTeamId,
    projectId: route.params.projectId,
  },
  pause: computed(() => !isAuthReady.value),
})

const { executeMutation: updateSuperTeam } = useUpdateSuperTeamMutation()
const { executeMutation: deleteSuperTeam } = useDeleteSuperTeamMutation()
const { executeMutation: assignTeams } = useAssignTeamsToSuperTeamMutation()

const state = reactive({
  name: '',
  description: '',
  imageUrl: null as string | null,
})

const hasColor = ref(false)
const colorValue = ref('#000000')

watch(data, () => {
  if (data.value) {
    const st = data.value.superteam
    state.name = st.name
    state.description = st.description
    state.imageUrl = st.imageObject?.url ?? null
    hasColor.value = !!st.color
    colorValue.value = st.color || '#000000'
  }
})

// Team assignment
const selectedTeamIds = ref<string[]>([])

watch(data, () => {
  if (data.value) {
    selectedTeamIds.value = data.value.superteam.teams.map((t) => t.id)
  }
})

const availableTeams = computed(() => {
  if (!data.value) return []
  return data.value.teams.edges
    .map((e) => e.node)
    .filter((t) => {
      // Show teams that are unassigned or assigned to this superteam
      return !t.superTeam || t.superTeam.id === route.params.superTeamId
    })
})

async function handleSave() {
  const response = await updateSuperTeam({
    id: route.params.superTeamId,
    input: {
      name: state.name,
      description: state.description,
      imageUrl: state.imageUrl || undefined,
      color: hasColor.value ? colorValue.value : undefined,
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
    description: 'Superteam oppdatert',
    color: 'success',
  })
}

async function handleAssignTeams() {
  const response = await assignTeams({
    superTeamId: route.params.superTeamId,
    teamIds: selectedTeamIds.value,
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
    description: 'Lag tilordnet',
    color: 'success',
  })

  refetch({ requestPolicy: 'network-only' })
}

async function handleDelete() {
  if (!confirm('Er du sikker på at du vil slette denne superteamen?')) return

  const response = await deleteSuperTeam({
    id: route.params.superTeamId,
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
    description: 'Superteam slettet',
    color: 'success',
  })

  navigateTo({
    name: 'admin-projects-projectId',
    params: { projectId: route.params.projectId },
    query: { tab: 'superteams' },
  })
}

function toggleTeam(teamId: string) {
  const idx = selectedTeamIds.value.indexOf(teamId)
  if (idx >= 0) {
    selectedTeamIds.value.splice(idx, 1)
  } else {
    selectedTeamIds.value.push(teamId)
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
              label: route.params.projectId,
              to: {
                name: 'admin-projects-projectId',
                params: { projectId: route.params.projectId },
              },
            },
            {
              label: 'Superteams',
            },
            {
              label: data?.superteam.name ?? route.params.superTeamId,
            },
          ]"
        />
      </UContainer>
    </div>
    <UContainer class="max-w-2xl py-12">
      <LoadingState v-if="fetching" />
      <ErrorState v-else-if="error" :error class="h-150" />
      <template v-else-if="data">
        <h1 class="mb-6 text-2xl font-bold">Rediger superteam</h1>

        <form class="space-y-6" @submit.prevent="handleSave">
          <UFormField name="name" label="Navn" required>
            <UInput v-model="state.name" class="w-full" />
          </UFormField>

          <UFormField name="description" label="Beskrivelse">
            <UTextarea v-model="state.description" class="w-full" />
          </UFormField>

          <UFormField name="imageUrl" label="Bilde">
            <AdminFileUpload v-model="state.imageUrl" />
          </UFormField>

          <UFormField name="color" label="Farge">
            <div class="flex items-center gap-3">
              <UCheckbox v-model="hasColor" />
              <template v-if="hasColor">
                <ColorPickerInput v-model="colorValue" />
              </template>
              <span v-else class="text-muted text-sm">Ingen farge valgt</span>
            </div>
          </UFormField>

          <div class="flex gap-2">
            <UButton type="submit" :disabled="!state.name">
              Lagre endringer
            </UButton>
          </div>
        </form>

        <hr class="border-default my-8" />

        <h2 class="mb-4 text-xl font-bold">Lag</h2>
        <p class="text-muted mb-4 text-sm">
          Velg hvilke lag som skal tilhøre denne superteamen.
        </p>

        <div
          v-if="availableTeams.length > 0"
          class="border-default mb-4 max-h-80 space-y-0 overflow-y-auto rounded-lg border"
        >
          <label
            v-for="team in availableTeams"
            :key="team.id"
            class="border-default flex cursor-pointer items-center gap-3 border-b px-4 py-3 last:border-b-0 hover:bg-gray-50 dark:hover:bg-gray-800"
          >
            <UCheckbox
              :model-value="selectedTeamIds.includes(team.id)"
              @update:model-value="toggleTeam(team.id)"
            />
            <span>{{ team.name }}</span>
          </label>
        </div>
        <div v-else class="text-dimmed py-4 text-sm">
          Ingen tilgjengelige lag i dette prosjektet.
        </div>

        <UButton @click="handleAssignTeams"> Lagre lagtilordning </UButton>

        <hr class="border-default my-8" />

        <div>
          <h2 class="mb-2 text-xl font-bold text-red-600">Faresone</h2>
          <p class="text-muted mb-4 text-sm">
            Sletting av superteam fjerner tilordningen til alle lag.
          </p>
          <UButton color="error" variant="soft" @click="handleDelete">
            Slett superteam
          </UButton>
        </div>
      </template>
    </UContainer>
  </div>
</template>
