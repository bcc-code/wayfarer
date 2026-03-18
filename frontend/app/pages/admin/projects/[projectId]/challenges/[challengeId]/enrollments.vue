<script setup lang="ts">
definePageMeta({
  layout: 'admin',
  middleware: 'admin',
})

gql(`
  query AdminChallengeEnrollmentsPage($challengeId: ID!) {
    challenge(id: $challengeId) {
      id
      name
      project {
        id
        name
      }
    }
  }
`)

const route = useRoute(
  'admin-projects-projectId-challenges-challengeId-enrollments',
)
const toast = useToast()

const { isAuthReady } = useAuthReady()
const { data, fetching, error } = useAdminChallengeEnrollmentsPageQuery({
  variables: {
    challengeId: route.params.challengeId,
  },
  pause: computed(() => !isAuthReady.value),
})

const {
  data: enrollmentsData,
  fetching: enrollmentsFetching,
  executeQuery: refetchEnrollments,
} = useAdminChallengeEnrollmentsQuery({
  variables: {
    challengeId: route.params.challengeId,
  },
  pause: computed(() => !isAuthReady.value),
})

const enrollments = computed(
  () => enrollmentsData.value?.challengeEnrollments ?? [],
)

const totalEnrolled = computed(() => enrollments.value.length)
const totalCompleted = computed(
  () => enrollments.value.filter((e) => e.completedAt).length,
)
const completionRate = computed(() =>
  totalEnrolled.value > 0
    ? Math.round((totalCompleted.value / totalEnrolled.value) * 100)
    : 0,
)

// Table columns
const columns = [
  { accessorKey: 'user', header: 'Bruker' },
  { accessorKey: 'enrolledAt', header: 'Påmeldt' },
  { accessorKey: 'completedAt', header: 'Fullført' },
  { accessorKey: 'status', header: 'Status' },
  { id: 'actions' },
]

// Mutations
const { executeMutation: completeChallenge } = useCompleteChallengeMutation()
const { executeMutation: uncompleteChallenge } =
  useUncompleteChallengeMutation()
const { executeMutation: unenrollUser } =
  useUnenrollUserFromChallengeMutation()
const { executeMutation: bulkEnroll } =
  useBulkEnrollUsersInChallengeAsyncMutation()
const { executeMutation: bulkComplete } =
  useBulkCompleteChallengesAsyncMutation()

async function handleComplete(userId: string) {
  const result = await completeChallenge({
    userId,
    challengeId: route.params.challengeId,
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
    description: 'Bruker markert som fullført',
    color: 'success',
  })
  refetchEnrollments({ requestPolicy: 'network-only' })
}

async function handleUncomplete(userId: string) {
  const result = await uncompleteChallenge({
    userId,
    challengeId: route.params.challengeId,
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
    description: 'Fullføring fjernet',
    color: 'success',
  })
  refetchEnrollments({ requestPolicy: 'network-only' })
}

async function handleUnenroll(userId: string, userName: string) {
  const confirmed = confirm(`Er du sikker på at du vil avmelde "${userName}"?`)
  if (!confirmed) return

  const result = await unenrollUser({
    userId,
    challengeId: route.params.challengeId,
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
    description: 'Bruker avmeldt',
    color: 'success',
  })
  refetchEnrollments({ requestPolicy: 'network-only' })
}

function getDropdownItems(enrollment: (typeof enrollments.value)[number]) {
  const items: { label: string; onSelect: () => void }[] = []

  if (enrollment.completedAt) {
    items.push({
      label: 'Fjern fullføring',
      onSelect: () => handleUncomplete(enrollment.user.id),
    })
  } else {
    items.push({
      label: 'Merk som fullført',
      onSelect: () => handleComplete(enrollment.user.id),
    })
  }

  items.push({
    label: 'Avmeld',
    onSelect: () => handleUnenroll(enrollment.user.id, enrollment.user.name),
  })

  return items
}

// Bulk enroll modal
const enrollModalOpen = ref(false)
const enrollForm = reactive({
  churchIds: '',
  teamIds: '',
  superTeamIds: '',
})

function openEnrollModal() {
  enrollForm.churchIds = ''
  enrollForm.teamIds = ''
  enrollForm.superTeamIds = ''
  enrollModalOpen.value = true
}

const parseIds = (str: string) => {
  const trimmed = str.trim()
  return trimmed
    ? trimmed
        .split(',')
        .map((s) => s.trim())
        .filter(Boolean)
    : undefined
}

async function handleBulkEnrollAllProjectUsers() {
  const projectId = route.params.projectId
  const result = await bulkEnroll({
    target: { allProjectMembers: projectId },
    challengeId: route.params.challengeId,
  })
  if (result.error) {
    toast.add({
      title: 'Feil',
      description: result.error.message,
      color: 'error',
    })
    return
  }
  const job = result.data?.bulkEnrollUsersInChallengeAsync
  toast.add({
    title: 'Suksess',
    description: `Jobb opprettet — ${job?.totalCount ?? 0} påmeldinger behandles`,
    color: 'success',
  })
  enrollModalOpen.value = false
  refetchEnrollments({ requestPolicy: 'network-only' })
}

async function handleBulkEnroll() {
  const target: Record<string, unknown> = {}
  const churchIds = parseIds(enrollForm.churchIds)
  const teamIds = parseIds(enrollForm.teamIds)
  const superTeamIds = parseIds(enrollForm.superTeamIds)

  if (churchIds) {
    target.churchInProject = {
      churchId: churchIds[0],
      projectId: route.params.projectId,
    }
  }
  if (teamIds) target.teamIds = teamIds
  if (superTeamIds) target.superTeamIds = superTeamIds

  if (!churchIds && !teamIds && !superTeamIds) {
    toast.add({
      title: 'Feil',
      description: 'Du må fylle inn minst ett felt',
      color: 'error',
    })
    return
  }

  const result = await bulkEnroll({
    target,
    challengeId: route.params.challengeId,
  })
  if (result.error) {
    toast.add({
      title: 'Feil',
      description: result.error.message,
      color: 'error',
    })
    return
  }
  const job = result.data?.bulkEnrollUsersInChallengeAsync
  toast.add({
    title: 'Suksess',
    description: `Jobb opprettet — ${job?.totalCount ?? 0} påmeldinger behandles`,
    color: 'success',
  })
  enrollModalOpen.value = false
  refetchEnrollments({ requestPolicy: 'network-only' })
}

// Bulk complete modal
const completeModalOpen = ref(false)
const completeForm = reactive({
  churchIds: '',
  teamIds: '',
  superTeamIds: '',
})

function openCompleteModal() {
  completeForm.churchIds = ''
  completeForm.teamIds = ''
  completeForm.superTeamIds = ''
  completeModalOpen.value = true
}

async function handleBulkCompleteAllProjectUsers() {
  const projectId = route.params.projectId
  const result = await bulkComplete({
    target: { allProjectMembers: projectId },
    challengeId: route.params.challengeId,
  })
  if (result.error) {
    toast.add({
      title: 'Feil',
      description: result.error.message,
      color: 'error',
    })
    return
  }
  const job = result.data?.bulkCompleteChallengesAsync
  toast.add({
    title: 'Suksess',
    description: `Jobb opprettet — ${job?.totalCount ?? 0} fullføringer behandles`,
    color: 'success',
  })
  completeModalOpen.value = false
  refetchEnrollments({ requestPolicy: 'network-only' })
}

async function handleBulkComplete() {
  const target: Record<string, unknown> = {}
  const churchIds = parseIds(completeForm.churchIds)
  const teamIds = parseIds(completeForm.teamIds)
  const superTeamIds = parseIds(completeForm.superTeamIds)

  if (churchIds) {
    target.churchInProject = {
      churchId: churchIds[0],
      projectId: route.params.projectId,
    }
  }
  if (teamIds) target.teamIds = teamIds
  if (superTeamIds) target.superTeamIds = superTeamIds

  if (!churchIds && !teamIds && !superTeamIds) {
    toast.add({
      title: 'Feil',
      description: 'Du må fylle inn minst ett felt',
      color: 'error',
    })
    return
  }

  const result = await bulkComplete({
    target,
    challengeId: route.params.challengeId,
  })
  if (result.error) {
    toast.add({
      title: 'Feil',
      description: result.error.message,
      color: 'error',
    })
    return
  }
  const job = result.data?.bulkCompleteChallengesAsync
  toast.add({
    title: 'Suksess',
    description: `Jobb opprettet — ${job?.totalCount ?? 0} fullføringer behandles`,
    color: 'success',
  })
  completeModalOpen.value = false
  refetchEnrollments({ requestPolicy: 'network-only' })
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
              label: 'Påmeldinger',
            },
          ]"
        />
      </UContainer>
    </div>
    <UContainer class="py-12">
      <LoadingState v-if="fetching" />
      <ErrorState v-else-if="error" :error />
      <template v-else-if="data">
        <div class="mb-6 flex items-center justify-between">
          <h1 class="text-2xl font-bold">Påmeldinger</h1>
          <div class="flex gap-2">
            <UButton variant="soft" @click="openCompleteModal">
              Massefullføring
            </UButton>
            <UButton @click="openEnrollModal"> Massepåmelding </UButton>
          </div>
        </div>

        <div class="mb-6 flex gap-6">
          <div class="bg-default rounded-lg border p-4">
            <div class="text-muted text-sm">Påmeldte</div>
            <div class="text-2xl font-bold">{{ totalEnrolled }}</div>
          </div>
          <div class="bg-default rounded-lg border p-4">
            <div class="text-muted text-sm">Fullførte</div>
            <div class="text-2xl font-bold">{{ totalCompleted }}</div>
          </div>
          <div class="bg-default rounded-lg border p-4">
            <div class="text-muted text-sm">Fullføringsrate</div>
            <div class="text-2xl font-bold">{{ completionRate }}%</div>
          </div>
        </div>

        <LoadingState v-if="enrollmentsFetching && !enrollmentsData" />
        <div
          v-else-if="enrollments.length === 0"
          class="text-muted py-12 text-center"
        >
          Ingen påmeldinger funnet.
        </div>
        <UTable v-else :data="enrollments" :columns>
          <template #user-cell="{ row }">
            <div>
              <div class="font-medium">{{ row.original.user.name }}</div>
              <div class="text-muted text-sm">
                {{ row.original.user.email }}
              </div>
            </div>
          </template>
          <template #enrolledAt-cell="{ row }">
            {{ formatDateTime(row.original.enrolledAt) }}
          </template>
          <template #completedAt-cell="{ row }">
            {{
              row.original.completedAt
                ? formatDateTime(row.original.completedAt)
                : '—'
            }}
          </template>
          <template #status-cell="{ row }">
            <UBadge
              variant="soft"
              :color="row.original.completedAt ? 'success' : 'warning'"
            >
              {{ row.original.completedAt ? 'Fullført' : 'Påmeldt' }}
            </UBadge>
          </template>
          <template #actions-cell="{ row }">
            <div class="flex justify-end">
              <UDropdownMenu :items="getDropdownItems(row.original)">
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
    </UContainer>

    <!-- Bulk enroll modal -->
    <UModal v-model:open="enrollModalOpen">
      <template #content>
        <div class="p-6">
          <h3 class="mb-4 text-lg font-semibold">Massepåmelding</h3>
          <div class="space-y-4">
            <UButton block @click="handleBulkEnrollAllProjectUsers">
              Meld på alle prosjektbrukere
            </UButton>
            <USeparator label="eller" />
            <UFormField name="churchIds" label="Kirke-IDer (kommaseparert)">
              <UInput
                v-model="enrollForm.churchIds"
                class="w-full"
                placeholder="CH..., CH..."
              />
            </UFormField>
            <UFormField name="teamIds" label="Lag-IDer (kommaseparert)">
              <UInput
                v-model="enrollForm.teamIds"
                class="w-full"
                placeholder="TM..., TM..."
              />
            </UFormField>
            <UFormField
              name="superTeamIds"
              label="Superlag-IDer (kommaseparert)"
            >
              <UInput
                v-model="enrollForm.superTeamIds"
                class="w-full"
                placeholder="ST..., ST..."
              />
            </UFormField>
          </div>
          <div class="mt-6 flex justify-end gap-3">
            <UButton
              variant="ghost"
              color="neutral"
              @click="enrollModalOpen = false"
            >
              Avbryt
            </UButton>
            <UButton @click="handleBulkEnroll"> Meld på </UButton>
          </div>
        </div>
      </template>
    </UModal>

    <!-- Bulk complete modal -->
    <UModal v-model:open="completeModalOpen">
      <template #content>
        <div class="p-6">
          <h3 class="mb-4 text-lg font-semibold">Massefullføring</h3>
          <div class="space-y-4">
            <UButton block @click="handleBulkCompleteAllProjectUsers">
              Fullfør for alle prosjektbrukere
            </UButton>
            <USeparator label="eller" />
            <UFormField name="churchIds" label="Kirke-IDer (kommaseparert)">
              <UInput
                v-model="completeForm.churchIds"
                class="w-full"
                placeholder="CH..., CH..."
              />
            </UFormField>
            <UFormField name="teamIds" label="Lag-IDer (kommaseparert)">
              <UInput
                v-model="completeForm.teamIds"
                class="w-full"
                placeholder="TM..., TM..."
              />
            </UFormField>
            <UFormField
              name="superTeamIds"
              label="Superlag-IDer (kommaseparert)"
            >
              <UInput
                v-model="completeForm.superTeamIds"
                class="w-full"
                placeholder="ST..., ST..."
              />
            </UFormField>
          </div>
          <div class="mt-6 flex justify-end gap-3">
            <UButton
              variant="ghost"
              color="neutral"
              @click="completeModalOpen = false"
            >
              Avbryt
            </UButton>
            <UButton @click="handleBulkComplete"> Fullfør </UButton>
          </div>
        </div>
      </template>
    </UModal>
  </div>
</template>
