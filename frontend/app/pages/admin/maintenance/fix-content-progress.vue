<script setup lang="ts">
import type { TableColumn } from '@nuxt/ui'

definePageMeta({
  layout: 'admin',
  middleware: ['admin', 'superadmin'],
})

gql(`
  query MaintenanceContentProgressPreview($first: Int) {
    previewMissingContentProgress(first: $first) {
      totalUsers
      totalEvents
      affectedUsers {
        user {
          id
          name
        }
        eventCount
      }
    }
  }
`)

gql(`
  mutation FixMissingContentProgressAsync {
    fixMissingContentProgressAsync {
      id
      operationType
      status
      totalCount
      processedCount
      successCount
      failureCount
      errorMessage
    }
  }
`)

gql(`
  query BulkJobStatus($id: ID!) {
    bulkJob(id: $id) {
      id
      status
      totalCount
      processedCount
      successCount
      failureCount
      errorMessage
      completedAt
    }
  }
`)

const { isAuthReady } = useAuthReady()
const {
  data,
  fetching,
  error,
  executeQuery: refetch,
} = useMaintenanceContentProgressPreviewQuery({
  variables: { first: 50 },
  pause: computed(() => !isAuthReady.value),
})

const preview = computed(() => data.value?.previewMissingContentProgress)
const affectedUsers = computed(() => preview.value?.affectedUsers ?? [])

type UserRow = NonNullable<
  MaintenanceContentProgressPreviewQuery['previewMissingContentProgress']
>['affectedUsers'][number]

const columns: TableColumn<UserRow>[] = [
  { accessorKey: 'user.name', header: 'Bruker' },
  { accessorKey: 'eventCount', header: 'Manglende fremgang' },
]

const { executeMutation: fix, fetching: fixing } =
  useFixMissingContentProgressAsyncMutation()

const showConfirmModal = ref(false)
const jobId = ref<string | null>(null)
const jobStatus = ref<string | null>(null)
const jobResult = ref<{
  successCount: number
  failureCount: number
  totalCount: number
} | null>(null)
const fixComplete = ref(false)

const { executeQuery: checkJobStatus } = useBulkJobStatusQuery({
  variables: computed(() => ({ id: jobId.value ?? '' })),
  pause: computed(() => !jobId.value),
  requestPolicy: 'network-only',
})

async function pollJobStatus() {
  if (!jobId.value) return

  const result = await checkJobStatus()
  const job = result.data?.bulkJob
  if (!job) return

  jobStatus.value = job.status

  if (job.status === 'COMPLETED' || job.status === 'FAILED') {
    jobResult.value = {
      successCount: job.successCount,
      failureCount: job.failureCount,
      totalCount: job.totalCount,
    }
    fixComplete.value = true
    refetch({ requestPolicy: 'network-only' })
  } else {
    // Poll again after 1 second
    setTimeout(pollJobStatus, 1000)
  }
}

async function handleFix() {
  showConfirmModal.value = false
  jobId.value = null
  jobStatus.value = null
  jobResult.value = null
  fixComplete.value = false

  const result = await fix({})
  if (result.data?.fixMissingContentProgressAsync) {
    const job = result.data.fixMissingContentProgressAsync
    jobId.value = job.id
    jobStatus.value = job.status
    // Start polling for job completion
    pollJobStatus()
  }
}

const toast = useToast()

watch(fixComplete, (complete) => {
  if (complete && jobResult.value) {
    if (jobStatus.value === 'COMPLETED') {
      toast.add({
        title: 'Reparasjon fullfort',
        description: `Behandlet ${jobResult.value.successCount} av ${jobResult.value.totalCount} hendelser.`,
        color: 'success',
      })
    } else {
      toast.add({
        title: 'Reparasjon feilet',
        description: 'Se jobbstatus for detaljer.',
        color: 'error',
      })
    }
  }
})
</script>

<template>
  <UContainer class="py-12">
    <div class="mb-8">
      <div class="mb-4 flex items-center gap-2">
        <UButton
          variant="ghost"
          icon="lucide:arrow-left"
          to="/admin/maintenance"
        />
        <h1 class="text-3xl">Fix innholdsfremgang</h1>
      </div>
      <p class="text-muted">
        Dette verktøyet finner brukere som har fullført eksterne
        innholdsoppgaver (f.eks. lest artikler, lyttet til podcasts), men som
        mangler tilhørende fremgangsregistreringer. Det oppretter manglende
        registreringer og tildeler prestasjoner der alle krav er oppfylt.
      </p>
    </div>

    <ErrorState v-if="error" :error />

    <template v-else>
      <div class="mb-8 grid gap-4 md:grid-cols-3">
        <UCard>
          <div class="flex flex-col gap-1">
            <span class="text-muted text-sm">Berørte brukere</span>
            <span v-if="fetching" class="text-2xl font-bold">...</span>
            <span v-else class="text-2xl font-bold">
              {{ preview?.totalUsers ?? 0 }}
            </span>
          </div>
        </UCard>
        <UCard>
          <div class="flex flex-col gap-1">
            <span class="text-muted text-sm">Manglende registreringer</span>
            <span v-if="fetching" class="text-2xl font-bold">...</span>
            <span v-else class="text-2xl font-bold">
              {{ preview?.totalEvents ?? 0 }}
            </span>
          </div>
        </UCard>
        <UCard>
          <div class="flex flex-col gap-1">
            <span class="text-muted text-sm">Status</span>
            <UBadge
              v-if="preview?.totalUsers === 0"
              variant="soft"
              color="success"
            >
              Alt i orden
            </UBadge>
            <UBadge v-else-if="fetching" variant="soft" color="neutral">
              Laster...
            </UBadge>
            <UBadge v-else variant="soft" color="warning">
              Trenger reparasjon
            </UBadge>
          </div>
        </UCard>
      </div>

      <div
        v-if="jobId && !fixComplete"
        class="bg-info/10 border-info mb-8 rounded-lg border p-4"
      >
        <div class="flex items-center gap-2">
          <UIcon name="lucide:loader-2" class="text-info h-5 w-5 animate-spin" />
          <span>Behandler... Status: {{ jobStatus }}</span>
        </div>
      </div>

      <div
        v-if="fixComplete && jobResult?.successCount === 0"
        class="bg-success/10 border-success mb-8 rounded-lg border p-4"
      >
        <div class="flex items-center gap-2">
          <UIcon name="lucide:check-circle" class="text-success h-5 w-5" />
          <span>Ingen data å reparere.</span>
        </div>
      </div>

      <div
        v-if="fixComplete && jobStatus === 'COMPLETED' && jobResult && jobResult.successCount > 0"
        class="bg-success/10 border-success mb-8 rounded-lg border p-4"
      >
        <div class="flex flex-col gap-2">
          <div class="flex items-center gap-2">
            <UIcon name="lucide:check-circle" class="text-success h-5 w-5" />
            <span class="font-medium">Reparasjon fullfort</span>
          </div>
          <ul class="text-muted list-inside list-disc text-sm">
            <li>{{ jobResult.successCount }} hendelser behandlet</li>
            <li v-if="jobResult.failureCount > 0">
              {{ jobResult.failureCount }} feil
            </li>
          </ul>
        </div>
      </div>

      <div
        v-if="fixComplete && jobStatus === 'FAILED'"
        class="bg-error/10 border-error mb-8 rounded-lg border p-4"
      >
        <div class="flex items-center gap-2">
          <UIcon name="lucide:x-circle" class="text-error h-5 w-5" />
          <span>Reparasjon feilet.</span>
        </div>
      </div>

      <div class="mb-6 flex items-center justify-between">
        <h2 class="text-xl">Berørte brukere (topp 50)</h2>
        <UButton
          v-if="preview && preview.totalUsers > 0"
          :loading="fixing"
          color="primary"
          @click="showConfirmModal = true"
        >
          Reparer
        </UButton>
      </div>

      <UTable :data="affectedUsers" :loading="fetching" :columns>
        <template #user.name-cell="{ row }">
          <NuxtLink
            :to="{
              name: 'admin-users-userId',
              params: { userId: row.original.user.id },
            }"
            class="hover:text-primary font-medium hover:underline"
          >
            {{ row.original.user.name }}
          </NuxtLink>
        </template>
        <template #eventCount-cell="{ row }">
          <UBadge variant="soft" color="warning">
            {{ row.original.eventCount }}
          </UBadge>
        </template>
      </UTable>

      <div
        v-if="affectedUsers.length === 0 && !fetching"
        class="text-muted py-12 text-center"
      >
        Ingen brukere med manglende fremgangsregistreringer funnet.
      </div>
    </template>

    <UModal v-model:open="showConfirmModal">
      <template #content>
        <UCard>
          <template #header>
            <div class="flex items-center gap-2">
              <UIcon name="lucide:alert-triangle" class="text-warning h-5 w-5" />
              <span class="font-medium">Bekreft reparasjon</span>
            </div>
          </template>
          <p class="text-muted">
            Er du sikker på at du vil kjøre reparasjonen? Dette vil opprette
            {{ preview?.totalEvents ?? 0 }} manglende fremgangsregistreringer
            for {{ preview?.totalUsers ?? 0 }} brukere, og tildele prestasjoner
            der alle krav er oppfylt.
          </p>
          <template #footer>
            <div class="flex justify-end gap-2">
              <UButton
                variant="ghost"
                :disabled="fixing"
                @click="showConfirmModal = false"
              >
                Avbryt
              </UButton>
              <UButton color="primary" :loading="fixing" @click="handleFix">
                Kjør reparasjon
              </UButton>
            </div>
          </template>
        </UCard>
      </template>
    </UModal>
  </UContainer>
</template>
