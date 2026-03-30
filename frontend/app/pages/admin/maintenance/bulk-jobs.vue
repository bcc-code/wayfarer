<script setup lang="ts">
import type { TableColumn } from '@nuxt/ui'
import { BulkJobStatus } from '~/api/generated'

definePageMeta({
  layout: 'admin',
  middleware: ['admin', 'superadmin'],
})

gql(`
  query AdminBulkJobsPage($filter: BulkJobFilter, $first: Int, $after: String, $last: Int, $before: String) {
    bulkJobs(filter: $filter, first: $first, after: $after, last: $last, before: $before) {
      totalCount
      pageInfo {
        hasNextPage
        hasPreviousPage
        startCursor
        endCursor
      }
      edges {
        cursor
        node {
          id
          operationType
          status
          totalCount
          processedCount
          successCount
          failureCount
          errorMessage
          createdAt
          startedAt
          completedAt
        }
      }
    }
  }
`)

const toast = useToast()
const { executeMutation: retryBulkJob } = useRetryBulkJobMutation()

const pagination = usePagination({
  defaultPageSize: 20,
  direction: 'backward',
})

// Filters
const selectedStatus = ref<BulkJobStatus | undefined>()
const selectedOperationType = ref<string | undefined>()

const filter = computed(() => ({
  status: selectedStatus.value,
  operationType: selectedOperationType.value,
}))

const queryVariables = computed(() => ({
  ...pagination.variables.value,
  filter: filter.value,
}))

const { isAuthReady } = useAuthReady()
const { data, fetching, error, executeQuery } = useAdminBulkJobsPageQuery({
  variables: queryVariables,
  pause: computed(() => !isAuthReady.value),
})

// Reset pagination when filter changes
watch([selectedStatus, selectedOperationType], () => {
  pagination.reset()
})

watch(
  () => data.value?.bulkJobs,
  (connection) => {
    pagination.updateConnection(connection)
  },
)

const jobs = computed(() =>
  data.value?.bulkJobs.edges.map((edge) => edge.node),
)

// Status options
const statusOptions = [
  { label: 'Alle', value: undefined },
  { label: 'Venter', value: BulkJobStatus.Pending },
  { label: 'Behandler', value: BulkJobStatus.Processing },
  { label: 'Fullfort', value: BulkJobStatus.Completed },
  { label: 'Feilet', value: BulkJobStatus.Failed },
]

// Operation type options
const operationTypes = [
  { label: 'Alle', value: undefined },
  { label: 'Meld pa utfordring', value: 'BULK_ENROLL_CHALLENGE' },
  { label: 'Meld av utfordring', value: 'BULK_UNENROLL_CHALLENGE' },
  { label: 'Fullfore utfordring', value: 'BULK_COMPLETE_CHALLENGE' },
  { label: 'Publiser utfordring', value: 'BULK_PUBLISH_CHALLENGE' },
  { label: 'Tildel prestasjon', value: 'BULK_AWARD_ACHIEVEMENT' },
]

type BulkJobNode = NonNullable<typeof jobs.value>[number]

const columns: TableColumn<BulkJobNode>[] = [
  { accessorKey: 'operationType', header: 'Operasjon' },
  { accessorKey: 'status', header: 'Status' },
  { id: 'progress', header: 'Fremgang' },
  { id: 'result', header: 'Resultat' },
  { accessorKey: 'createdAt', header: 'Opprettet' },
  { id: 'duration', header: 'Varighet' },
  { id: 'actions', header: '' },
]

// Helper functions
function formatOperationType(type: string): string {
  const labels: Record<string, string> = {
    BULK_ENROLL_CHALLENGE: 'Meld pa utfordring',
    BULK_UNENROLL_CHALLENGE: 'Meld av utfordring',
    BULK_COMPLETE_CHALLENGE: 'Fullfore utfordring',
    BULK_PUBLISH_CHALLENGE: 'Publiser utfordring',
    BULK_AWARD_ACHIEVEMENT: 'Tildel prestasjon',
  }
  return labels[type] ?? type
}

function getStatusColor(status: BulkJobStatus): 'neutral' | 'info' | 'success' | 'error' {
  switch (status) {
    case BulkJobStatus.Pending:
      return 'neutral'
    case BulkJobStatus.Processing:
      return 'info'
    case BulkJobStatus.Completed:
      return 'success'
    case BulkJobStatus.Failed:
      return 'error'
    default:
      return 'neutral'
  }
}

function getStatusLabel(status: BulkJobStatus): string {
  switch (status) {
    case BulkJobStatus.Pending:
      return 'Venter'
    case BulkJobStatus.Processing:
      return 'Behandler'
    case BulkJobStatus.Completed:
      return 'Fullfort'
    case BulkJobStatus.Failed:
      return 'Feilet'
    default:
      return status
  }
}

function formatDuration(startedAt: string | null | undefined, completedAt: string | null | undefined): string {
  if (!startedAt) return '-'
  const start = new Date(startedAt)
  const end = completedAt ? new Date(completedAt) : new Date()
  const diffMs = end.getTime() - start.getTime()

  if (diffMs < 1000) return `${diffMs}ms`
  if (diffMs < 60000) return `${Math.round(diffMs / 1000)}s`
  if (diffMs < 3600000) return `${Math.round(diffMs / 60000)}min`
  return `${Math.round(diffMs / 3600000)}t`
}

function calculateProgress(processed: number, total: number): number {
  if (total === 0) return 0
  return Math.round((processed / total) * 100)
}

async function handleRetry(jobId: string) {
  const { error } = await retryBulkJob({ id: jobId })
  if (error) {
    toast.add({ title: 'Kunne ikke kjore jobb pa nytt', description: error.message, color: 'error' })
    return
  }
  toast.add({ title: 'Jobb opprettet pa nytt', color: 'success' })
  handleRefresh()
}

function handleRefresh() {
  executeQuery({ requestPolicy: 'network-only' })
}

function clearFilters() {
  selectedStatus.value = undefined
  selectedOperationType.value = undefined
}

const hasActiveFilters = computed(
  () => selectedStatus.value !== undefined || selectedOperationType.value !== undefined,
)
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
        <h1 class="text-3xl">Massejobber</h1>
      </div>
      <p class="text-muted">
        Oversikt over asynkrone massejobber og deres status.
      </p>
    </div>

    <ErrorState v-if="error" :error />

    <div v-else class="space-y-4">
      <div class="flex items-center gap-2">
        <USelectMenu
          :model-value="statusOptions.find((o) => o.value === selectedStatus)?.value"
          :items="statusOptions"
          value-key="value"
          label-key="label"
          placeholder="Status..."
          icon="lucide:circle-dot"
          size="sm"
          class="min-w-40"
          @update:model-value="selectedStatus = $event"
        />
        <USelectMenu
          :model-value="operationTypes.find((o) => o.value === selectedOperationType)?.value"
          :items="operationTypes"
          value-key="value"
          label-key="label"
          placeholder="Operasjon..."
          icon="lucide:layers"
          size="sm"
          class="min-w-52"
          @update:model-value="selectedOperationType = $event"
        />
        <UButton
          v-if="hasActiveFilters"
          variant="ghost"
          size="sm"
          color="neutral"
          label="Nullstill"
          @click="clearFilters"
        />
        <div class="ml-auto flex items-center gap-2">
          <UButton
            variant="soft"
            size="sm"
            icon="lucide:refresh-cw"
            label="Oppdater"
            :loading="fetching"
            @click="handleRefresh"
          />
          <RelayPagination v-model:pagination="pagination" />
        </div>
      </div>

      <UTable :data="jobs" :loading="fetching" :columns>
        <template #operationType-cell="{ row }">
          <span class="font-medium">
            {{ formatOperationType(row.original.operationType) }}
          </span>
        </template>

        <template #status-cell="{ row }">
          <UBadge :color="getStatusColor(row.original.status)" variant="soft">
            {{ getStatusLabel(row.original.status) }}
          </UBadge>
        </template>

        <template #progress-cell="{ row }">
          <div class="flex items-center gap-2 min-w-32">
            <div class="h-2 flex-1 overflow-hidden rounded-full bg-neutral-200 dark:bg-neutral-700">
              <div
                class="h-full rounded-full transition-all"
                :class="{
                  'bg-neutral-400': row.original.status === BulkJobStatus.Pending,
                  'bg-blue-500': row.original.status === BulkJobStatus.Processing,
                  'bg-green-500': row.original.status === BulkJobStatus.Completed,
                  'bg-red-500': row.original.status === BulkJobStatus.Failed,
                }"
                :style="{ width: `${calculateProgress(row.original.processedCount, row.original.totalCount)}%` }"
              />
            </div>
            <span class="text-dimmed text-xs whitespace-nowrap">
              {{ row.original.processedCount }}/{{ row.original.totalCount }}
            </span>
          </div>
        </template>

        <template #result-cell="{ row }">
          <div class="flex items-center gap-2">
            <span class="text-success text-sm">{{ row.original.successCount }}</span>
            <span class="text-dimmed">/</span>
            <span class="text-error text-sm">{{ row.original.failureCount }}</span>
            <UPopover v-if="row.original.errorMessage">
              <UButton
                variant="ghost"
                color="error"
                size="xs"
                icon="lucide:alert-circle"
              />
              <template #content>
                <div class="max-w-md p-3">
                  <p class="text-sm font-medium mb-1">Feilmelding</p>
                  <p class="text-dimmed text-sm whitespace-pre-wrap">
                    {{ row.original.errorMessage }}
                  </p>
                </div>
              </template>
            </UPopover>
          </div>
        </template>

        <template #createdAt-cell="{ row }">
          <span class="text-dimmed text-sm">
            {{ formatDateTime(row.original.createdAt) }}
          </span>
        </template>

        <template #duration-cell="{ row }">
          <span class="text-dimmed text-sm">
            {{ formatDuration(row.original.startedAt, row.original.completedAt) }}
          </span>
        </template>

        <template #actions-cell="{ row }">
          <UButton
            v-if="row.original.status === BulkJobStatus.Completed || row.original.status === BulkJobStatus.Failed"
            variant="ghost"
            size="xs"
            icon="lucide:refresh-cw"
            @click="handleRetry(row.original.id)"
          />
        </template>
      </UTable>

      <UEmpty
        v-if="!fetching && jobs?.length === 0"
        title="Ingen massejobber funnet"
        description="Det finnes ingen jobber som matcher de valgte filtrene."
      />
    </div>
  </UContainer>
</template>
