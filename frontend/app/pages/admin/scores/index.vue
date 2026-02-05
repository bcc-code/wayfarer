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
  direction: 'backward',
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
  { accessorKey: 'user.name', id: 'user', header: 'Bruker' },
  { accessorKey: 'project.name', id: 'project', header: 'Prosjekt' },
  { accessorKey: 'points', header: 'Poeng' },
  { accessorKey: 'sourceType', header: 'Kilde' },
  { accessorKey: 'reason', header: 'Grunn' },
  { accessorKey: 'createdAt', header: 'Opprettet' },
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
      title: 'Kunne ikke slette oppføring',
      description: result.error.message,
      color: 'error',
    })
  } else {
    toast.add({
      title: 'Oppføring slettet',
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

const { canDeleteScoreEntry, canManageScores } = usePermissions()
</script>

<template>
  <UContainer class="py-12">
    <div class="mb-6 flex items-center justify-between">
      <h1 class="text-3xl">Poengjusteringer</h1>
      <UButton v-if="canManageScores" :to="{ name: 'admin-scores-new' }">
        Ny justering
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
            {{ formatDateTime(row.original.createdAt) }}
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
          <Icon name="lucide:triangle-alert" class="text-error size-8" />
          <h3 class="my-2 text-lg font-semibold">Slett poengoppføring</h3>
          <p class="text-dimmed mb-6">
            Er du sikker på at du vil slette denne poengoppføringen? Denne
            handlingen kan ikke angres.
          </p>
          <div class="flex justify-end gap-3">
            <UButton variant="ghost" @click="deleteModal = false">
              Avbryt
            </UButton>
            <UButton color="error" @click="handleDelete">Slett</UButton>
          </div>
        </div>
      </template>
    </UModal>
  </UContainer>
</template>
