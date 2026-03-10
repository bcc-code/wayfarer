<script setup lang="ts">
definePageMeta({
  layout: 'admin',
  middleware: 'admin',
})

gql(`
  query AdminChallengeSessionsPage($challengeId: ID!) {
    challenge(id: $challengeId) {
      __typename
      id
      name
      project {
        id
        name
      }
      ... on QuizChallenge {
        quiz {
          id
        }
      }
    }
  }
`)

const route = useRoute(
  'admin-projects-projectId-challenges-challengeId-sessions',
)
const toast = useToast()

const { isAuthReady } = useAuthReady()
const { data, fetching, error } = useAdminChallengeSessionsPageQuery({
  variables: {
    challengeId: route.params.challengeId,
  },
  pause: computed(() => !isAuthReady.value),
})

const quizId = computed(() => {
  if (data.value?.challenge.__typename === 'QuizChallenge') {
    return data.value.challenge.quiz?.id
  }
  return undefined
})

// State filter
const stateFilterOptions = [
  { value: '', label: 'Alle' },
  { value: QuizSessionState.Draft, label: 'Utkast' },
  { value: QuizSessionState.Open, label: 'Åpen' },
  { value: QuizSessionState.Locked, label: 'Låst' },
  { value: QuizSessionState.Finished, label: 'Fullført' },
]
const stateFilter = ref('')

const {
  data: sessionsData,
  fetching: sessionsFetching,
  executeQuery: refetchSessions,
} = useAdminQuizSessionsQuery({
  variables: computed(() => ({
    quizId: quizId.value ?? '',
    state: stateFilter.value
      ? (stateFilter.value as QuizSessionState)
      : undefined,
  })),
  pause: computed(() => !isAuthReady.value || !quizId.value),
})

const sessions = computed(() => sessionsData.value?.quizSessions ?? [])

// Table columns
const columns = [
  { accessorKey: 'name', header: 'Navn' },
  { accessorKey: 'state', header: 'Status' },
  { accessorKey: 'accessCount', header: 'Tilganger' },
  { accessorKey: 'createdAt', header: 'Opprettet' },
  { id: 'actions' },
]

// State badge config
function stateBadgeColor(state: QuizSessionState) {
  switch (state) {
    case QuizSessionState.Draft:
      return 'neutral' as const
    case QuizSessionState.Open:
      return 'info' as const
    case QuizSessionState.Locked:
      return 'warning' as const
    case QuizSessionState.Finished:
      return 'success' as const
  }
}

function stateLabel(state: QuizSessionState) {
  switch (state) {
    case QuizSessionState.Draft:
      return 'Utkast'
    case QuizSessionState.Open:
      return 'Åpen'
    case QuizSessionState.Locked:
      return 'Låst'
    case QuizSessionState.Finished:
      return 'Fullført'
  }
}

// Mutations
const { executeMutation: createSession } = useCreateQuizSessionMutation()
const { executeMutation: updateSession } = useUpdateQuizSessionMutation()
const { executeMutation: deleteSession } = useDeleteQuizSessionMutation()
const { executeMutation: openSession } = useOpenQuizSessionMutation()
const { executeMutation: lockSession } = useLockQuizSessionMutation()
const { executeMutation: reopenSession } = useReopenQuizSessionMutation()
const { executeMutation: finishSession } = useFinishQuizSessionMutation()
const { executeMutation: grantAccess } = useGrantQuizSessionAccessMutation()

// Create/Edit modal
const editModalOpen = ref(false)
const editingSession = ref<{
  id: string
  name?: string | null
  openAt?: string | null
  lockAt?: string | null
  finishAt?: string | null
} | null>(null)
const editForm = reactive({
  name: '',
  openAt: '',
  lockAt: '',
  finishAt: '',
})

function openCreateModal() {
  editingSession.value = null
  editForm.name = ''
  editForm.openAt = ''
  editForm.lockAt = ''
  editForm.finishAt = ''
  editModalOpen.value = true
}

function openEditModal(session: (typeof sessions.value)[number]) {
  editingSession.value = session
  editForm.name = session.name ?? ''
  editForm.openAt = toLocalDatetimeLocal(session.openAt)
  editForm.lockAt = toLocalDatetimeLocal(session.lockAt)
  editForm.finishAt = toLocalDatetimeLocal(session.finishAt)
  editModalOpen.value = true
}

async function handleSaveSession() {
  const input = {
    name: editForm.name || undefined,
    openAt: toISOString(editForm.openAt),
    lockAt: toISOString(editForm.lockAt),
    finishAt: toISOString(editForm.finishAt),
  }

  if (editingSession.value) {
    const result = await updateSession({ id: editingSession.value.id, input })
    if (result.error) {
      toast.add({
        title: 'Feil',
        description: result.error.message,
        color: 'error',
      })
      return
    }
    toast.add({
      title: 'Suksess',
      description: 'Sesjon oppdatert',
      color: 'success',
    })
  } else {
    if (!quizId.value) return
    const result = await createSession({
      input: { quizId: quizId.value, ...input },
    })
    if (result.error) {
      toast.add({
        title: 'Feil',
        description: result.error.message,
        color: 'error',
      })
      return
    }
    toast.add({
      title: 'Suksess',
      description: 'Sesjon opprettet',
      color: 'success',
    })
  }

  editModalOpen.value = false
  refetchSessions({ requestPolicy: 'network-only' })
}

// Access modal
const accessModalOpen = ref(false)
const accessSessionId = ref<string | null>(null)
const accessForm = reactive({
  churchIds: '',
  teamIds: '',
  superTeamIds: '',
})

function openAccessModal(sessionId: string) {
  accessSessionId.value = sessionId
  accessForm.churchIds = ''
  accessForm.teamIds = ''
  accessForm.superTeamIds = ''
  accessModalOpen.value = true
}

async function handleGrantAllProjectUsers() {
  if (!accessSessionId.value) return
  const result = await grantAccess({
    input: {
      sessionId: accessSessionId.value,
      allProjectUsers: true,
    },
  })
  if (result.error) {
    toast.add({
      title: 'Feil',
      description: result.error.message,
      color: 'error',
    })
    return
  }
  toast.add({
    title: 'Suksess',
    description: `${result.data?.grantQuizSessionAccess ?? 0} tilganger gitt`,
    color: 'success',
  })
  accessModalOpen.value = false
  refetchSessions({ requestPolicy: 'network-only' })
}

async function handleGrantAccess() {
  if (!accessSessionId.value) return

  const parseIds = (str: string) => {
    const trimmed = str.trim()
    return trimmed
      ? trimmed
          .split(',')
          .map((s) => s.trim())
          .filter(Boolean)
      : undefined
  }

  const result = await grantAccess({
    input: {
      sessionId: accessSessionId.value,
      churchIds: parseIds(accessForm.churchIds),
      teamIds: parseIds(accessForm.teamIds),
      superTeamIds: parseIds(accessForm.superTeamIds),
    },
  })
  if (result.error) {
    toast.add({
      title: 'Feil',
      description: result.error.message,
      color: 'error',
    })
    return
  }
  toast.add({
    title: 'Suksess',
    description: `${result.data?.grantQuizSessionAccess ?? 0} tilganger gitt`,
    color: 'success',
  })
  accessModalOpen.value = false
  refetchSessions({ requestPolicy: 'network-only' })
}

// State transition actions
async function handleStateAction(
  action: 'open' | 'lock' | 'reopen' | 'finish',
  sessionId: string,
) {
  const mutations = {
    open: openSession,
    lock: lockSession,
    reopen: reopenSession,
    finish: finishSession,
  }
  const labels = {
    open: 'åpnet',
    lock: 'låst',
    reopen: 'gjenåpnet',
    finish: 'fullført',
  }

  const result = await mutations[action]({ id: sessionId })
  if (result.error) {
    toast.add({
      title: 'Feil',
      description: result.error.message,
      color: 'error',
    })
    return
  }
  toast.add({
    title: 'Suksess',
    description: `Sesjon ${labels[action]}`,
    color: 'success',
  })
  refetchSessions({ requestPolicy: 'network-only' })
}

async function handleDelete(sessionId: string, sessionName?: string | null) {
  const confirmed = confirm(
    `Er du sikker på at du vil slette sesjonen "${sessionName ?? sessionId}"? Denne handlingen kan ikke angres.`,
  )
  if (!confirmed) return

  const result = await deleteSession({ id: sessionId })
  if (result.error) {
    toast.add({
      title: 'Feil',
      description: result.error.message,
      color: 'error',
    })
    return
  }
  toast.add({
    title: 'Suksess',
    description: 'Sesjon slettet',
    color: 'success',
  })
  refetchSessions({ requestPolicy: 'network-only' })
}

function getDropdownItems(session: (typeof sessions.value)[number]) {
  const items: { label: string; onSelect: () => void }[] = []

  switch (session.state) {
    case QuizSessionState.Draft:
      items.push(
        { label: 'Rediger', onSelect: () => openEditModal(session) },
        {
          label: 'Åpne',
          onSelect: () => handleStateAction('open', session.id),
        },
        {
          label: 'Slett',
          onSelect: () => handleDelete(session.id, session.name),
        },
      )
      break
    case QuizSessionState.Open:
      items.push(
        { label: 'Lås', onSelect: () => handleStateAction('lock', session.id) },
        {
          label: 'Administrer tilgang',
          onSelect: () => openAccessModal(session.id),
        },
      )
      break
    case QuizSessionState.Locked:
      items.push(
        {
          label: 'Gjenåpne',
          onSelect: () => handleStateAction('reopen', session.id),
        },
        {
          label: 'Fullfør',
          onSelect: () => handleStateAction('finish', session.id),
        },
        {
          label: 'Administrer tilgang',
          onSelect: () => openAccessModal(session.id),
        },
      )
      break
  }

  return items
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
            {
              label: 'Sesjoner',
            },
          ]"
        />
      </UContainer>
    </div>
    <UContainer class="py-12">
      <LoadingState v-if="fetching" />
      <ErrorState v-else-if="error" :error />
      <template v-else-if="data">
        <div
          v-if="data.challenge.__typename !== 'QuizChallenge'"
          class="text-center py-12"
        >
          <p class="text-muted">
            Denne utfordringen er ikke en quiz-utfordring.
          </p>
          <UButton
            class="mt-4"
            :to="{
              name: 'admin-projects-projectId-challenges-challengeId',
              params: {
                projectId: route.params.projectId,
                challengeId: route.params.challengeId,
              },
            }"
          >
            Tilbake til utfordring
          </UButton>
        </div>
        <template v-else>
          <div class="mb-6 flex items-center justify-between">
            <h1 class="text-2xl font-bold">Sesjoner</h1>
            <UButton icon="lucide:plus" @click="openCreateModal">
              Opprett sesjon
            </UButton>
          </div>

          <div class="mb-4">
            <USelect
              v-model="stateFilter"
              :items="stateFilterOptions"
              value-key="value"
              label-key="label"
              class="w-48"
            />
          </div>

          <LoadingState v-if="sessionsFetching && !sessionsData" />
          <div
            v-else-if="sessions.length === 0"
            class="text-muted py-12 text-center"
          >
            Ingen sesjoner funnet.
          </div>
          <UTable v-else :data="sessions" :columns>
            <template #name-cell="{ row }">
              {{ row.original.name || '(uten navn)' }}
            </template>
            <template #state-cell="{ row }">
              <UBadge
                variant="soft"
                :color="stateBadgeColor(row.original.state)"
              >
                {{ stateLabel(row.original.state) }}
              </UBadge>
            </template>
            <template #createdAt-cell="{ row }">
              {{ formatDateTime(row.original.createdAt) }}
            </template>
            <template #actions-cell="{ row }">
              <div class="flex justify-end">
                <UDropdownMenu
                  v-if="row.original.state !== QuizSessionState.Finished"
                  :items="getDropdownItems(row.original)"
                >
                  <UButton
                    variant="ghost"
                    size="sm"
                    icon="lucide:more-horizontal"
                  />
                </UDropdownMenu>
              </div>
            </template>
          </UTable>
        </template>
      </template>
    </UContainer>

    <!-- Create/Edit modal -->
    <UModal v-model:open="editModalOpen">
      <template #content>
        <div class="p-6">
          <h3 class="mb-4 text-lg font-semibold">
            {{ editingSession ? 'Rediger sesjon' : 'Opprett sesjon' }}
          </h3>
          <div class="space-y-4">
            <UFormField name="name" label="Navn">
              <UInput v-model="editForm.name" class="w-full" />
            </UFormField>
            <UFormField name="openAt" label="Åpningstidspunkt">
              <UInput
                v-model="editForm.openAt"
                type="datetime-local"
                class="w-full"
              />
            </UFormField>
            <UFormField name="lockAt" label="Låsetidspunkt">
              <UInput
                v-model="editForm.lockAt"
                type="datetime-local"
                class="w-full"
              />
            </UFormField>
            <UFormField name="finishAt" label="Avslutningstidspunkt">
              <UInput
                v-model="editForm.finishAt"
                type="datetime-local"
                class="w-full"
              />
            </UFormField>
          </div>
          <div class="mt-6 flex justify-end gap-3">
            <UButton
              variant="ghost"
              color="neutral"
              @click="editModalOpen = false"
            >
              Avbryt
            </UButton>
            <UButton @click="handleSaveSession">
              {{ editingSession ? 'Lagre' : 'Opprett' }}
            </UButton>
          </div>
        </div>
      </template>
    </UModal>

    <!-- Access modal -->
    <UModal v-model:open="accessModalOpen">
      <template #content>
        <div class="p-6">
          <h3 class="mb-4 text-lg font-semibold">Administrer tilgang</h3>
          <div class="space-y-4">
            <UButton block @click="handleGrantAllProjectUsers">
              Gi tilgang til alle prosjektbrukere
            </UButton>
            <USeparator label="eller" />
            <UFormField name="churchIds" label="Kirke-IDer (kommaseparert)">
              <UInput
                v-model="accessForm.churchIds"
                class="w-full"
                placeholder="CH..., CH..."
              />
            </UFormField>
            <UFormField name="teamIds" label="Lag-IDer (kommaseparert)">
              <UInput
                v-model="accessForm.teamIds"
                class="w-full"
                placeholder="TM..., TM..."
              />
            </UFormField>
            <UFormField
              name="superTeamIds"
              label="Superlag-IDer (kommaseparert)"
            >
              <UInput
                v-model="accessForm.superTeamIds"
                class="w-full"
                placeholder="ST..., ST..."
              />
            </UFormField>
          </div>
          <div class="mt-6 flex justify-end gap-3">
            <UButton
              variant="ghost"
              color="neutral"
              @click="accessModalOpen = false"
            >
              Avbryt
            </UButton>
            <UButton @click="handleGrantAccess"> Gi tilgang </UButton>
          </div>
        </div>
      </template>
    </UModal>
  </div>
</template>
