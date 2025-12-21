<script setup lang="ts">
import type { TableColumn } from '@nuxt/ui'

definePageMeta({
  layout: 'admin',
  middleware: 'admin',
})

gql(`
  query AdminScoresPage($filter: ScoreJournalFilter, $first: Int, $after: String, $last: Int, $before: String) {
    adminScoreJournal(
      filter: $filter
      first: $first
      after: $after
      last: $last
      before: $before
    ) {
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
          points
          sourceType
          reason
          createdAt
          user {
            id
            name
          }
          project {
            id
            name
          }
          awardedBy {
            id
            name
          }
        }
      }
    }
  }
`)

gql(`
  mutation DeleteScoreJournalEntry($id: ID!) {
    deleteScoreJournalEntry(id: $id)
  }
`)

const pagination = usePagination({
  defaultPageSize: 15,
})

const { isAuthReady } = useAuthReady()
const { data, fetching, error, executeQuery } = useAdminScoresPageQuery({
  variables: pagination.variables,
  pause: computed(() => !isAuthReady.value),
})

watch(
  () => data.value?.adminScoreJournal,
  (connection) => {
    pagination.updateConnection(connection)
  },
)

const entries = computed(() =>
  data.value?.adminScoreJournal.edges.map((edge) => edge.node),
)

const columns: TableColumn<
  AdminScoresPageQuery['adminScoreJournal']['edges'][number]['node']
>[] = [
  { accessorKey: 'user.name', id: 'user', header: 'User' },
  { accessorKey: 'project.name', id: 'project', header: 'Project' },
  { accessorKey: 'points', header: 'Points' },
  { accessorKey: 'sourceType', header: 'Source' },
  { accessorKey: 'reason', header: 'Reason' },
  { accessorKey: 'createdAt', header: 'Created' },
  { id: 'actions' },
]

// Delete functionality
const { executeMutation: deleteEntry } = useDeleteScoreJournalEntryMutation()
const toast = useToast()
const deleteModal = ref(false)
const entryToDelete = ref<string | null>(null)

function confirmDelete(id: string) {
  entryToDelete.value = id
  deleteModal.value = true
}

async function handleDelete() {
  if (!entryToDelete.value) return

  const result = await deleteEntry({ id: entryToDelete.value })

  if (result.error) {
    toast.add({
      title: 'Failed to delete entry',
      description: result.error.message,
      color: 'error',
    })
  } else {
    toast.add({
      title: 'Entry deleted',
      color: 'success',
    })
    executeQuery({ requestPolicy: 'network-only' })
  }

  deleteModal.value = false
  entryToDelete.value = null
}

function formatSourceType(type: string) {
  return type.charAt(0) + type.slice(1).toLowerCase()
}

function formatDate(date: string) {
  return new Date(date).toLocaleDateString('en-US', {
    year: 'numeric',
    month: 'short',
    day: 'numeric',
    hour: '2-digit',
    minute: '2-digit',
  })
}

const { canDeleteScoreEntry, canManageScores } = usePermissions()
</script>

<template>
  <UContainer class="py-12">
    <div class="mb-6 flex items-center justify-between">
      <h1 class="text-3xl">Score Adjustments</h1>
      <UButton v-if="canManageScores" :to="{ name: 'admin-scores-new' }">
        New adjustment
      </UButton>
    </div>
    <ErrorState v-if="error" :error />
    <div v-else class="space-y-4">
      <div class="flex items-center justify-between gap-2">
        <RelayPagination v-model:pagination="pagination" />
      </div>
      <UTable :data="entries" :loading="fetching" :columns>
        <template #user-cell="{ row }">
          <NuxtLink
            :to="{
              name: 'admin-users-userId',
              params: { userId: row.original.user.id },
            }"
            class="hover:underline"
          >
            {{ row.original.user.name }}
          </NuxtLink>
        </template>
        <template #points-cell="{ row }">
          <UBadge
            :color="row.original.points >= 0 ? 'success' : 'error'"
            variant="soft"
          >
            {{ row.original.points >= 0 ? '+' : ''
            }}{{ formatNumber(row.original.points) }}
          </UBadge>
        </template>
        <template #sourceType-cell="{ row }">
          <UBadge variant="subtle">
            {{ formatSourceType(row.original.sourceType) }}
          </UBadge>
        </template>
        <template #reason-cell="{ row }">
          <span class="text-dimmed line-clamp-1 max-w-xs text-sm">
            {{ row.original.reason || '—' }}
          </span>
        </template>
        <template #createdAt-cell="{ row }">
          <span class="text-dimmed text-sm">
            {{ formatDate(row.original.createdAt) }}
          </span>
        </template>
        <template #actions-cell="{ row }">
          <div class="flex justify-end">
            <UButton
              v-if="canDeleteScoreEntry"
              variant="ghost"
              color="error"
              icon="i-lucide-trash-2"
              @click="confirmDelete(row.original.id)"
            />
          </div>
        </template>
      </UTable>
    </div>

    <UModal v-model:open="deleteModal">
      <template #content>
        <div class="p-6">
          <h3 class="mb-4 text-lg font-semibold">Delete Score Entry</h3>
          <p class="text-dimmed mb-6">
            Are you sure you want to delete this score entry? This action cannot
            be undone.
          </p>
          <div class="flex justify-end gap-3">
            <UButton variant="ghost" @click="deleteModal = false">
              Cancel
            </UButton>
            <UButton color="error" @click="handleDelete">Delete</UButton>
          </div>
        </div>
      </template>
    </UModal>
  </UContainer>
</template>
