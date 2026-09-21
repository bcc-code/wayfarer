<script setup lang="ts">
import type { TableColumn } from '@nuxt/ui'
import { ScoreSourceType } from '~/api/generated'

definePageMeta({
  permission: 'scores:view',
  layout: 'admin',
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

const route = useRoute('admin-projects-projectId-scores')

/**
 * `ScoreJournalFilter` has no free-text field. Its useful facet is the source:
 * separating manual adjustments from automatic awards is the common question
 * of a points journal, and `MANUAL` is the one an admin is answerable for.
 */
function formatSourceType(type: string) {
  return type.charAt(0) + type.slice(1).toLowerCase()
}

// Codegen emits a real TS enum, so the members are the values — a string
// literal list would not typecheck against it.
const sourceTypes = Object.values(ScoreSourceType)

const list = useListState({
  pagination,
  filters: { sourceType: '' },
})

const { isAuthReady } = useAuthReady()
const { data, fetching, error, executeQuery } = useAdminScoresPageQuery({
  // Merged with the pagination cursor, and computed so the journal repoints
  // when the project switcher changes the param.
  variables: computed(() => ({
    ...pagination.variables.value,
    filter: {
      projectId: route.params.projectId,
      ...(list.filters.sourceType
        ? { sourceType: list.filters.sourceType as ScoreSourceType }
        : {}),
    },
  })),
  pause: computed(() => !isAuthReady.value),
})

// `value` stays a plain string: `useListState` holds URL params, which are
// strings, and typing the items as the enum would make USelect demand a
// `ScoreSourceType` model. The cast back to the enum happens at the query.
const sourceTypeItems = computed(() =>
  sourceTypes.map((type) => ({
    label: formatSourceType(type),
    value: String(type),
  })),
)

const activeFilters = computed(() =>
  list.activeFilters.value.map((filter) => ({
    ...filter,
    label: `Kilde: ${formatSourceType(filter.value)}`,
  })),
)

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

const { canDeleteScoreEntry, canManageScores } = usePermissions()
</script>

<template>
  <div>
    <div class="mb-6 flex items-center justify-between">
      <h1 class="text-3xl">Poengjusteringer</h1>
      <UButton
        v-if="canManageScores"
        :to="{
          name: 'admin-projects-projectId-scores-new',
          params: { projectId: route.params.projectId },
        }"
      >
        Ny justering
      </UButton>
    </div>
    <AdminErrorState v-if="error" :error />
    <AdminListView
      v-else
      :pagination
      :active-filters="activeFilters"
      :searchable="false"
      item-label="oppføringer"
      @clear-filter="list.clearFilter($event as 'sourceType')"
      @clear-all="list.clearAll()"
    >
      <template #filters>
        <USelect
          v-model="list.filters.sourceType"
          :items="sourceTypeItems"
          value-key="value"
          placeholder="Alle kilder"
          icon="lucide:filter"
          class="w-48"
        />
      </template>

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
        <template #empty>
          <AdminTableEmpty
            :filtered="!!activeFilters.length"
            title="Ingen poengoppføringer i dette prosjektet"
            filtered-title="Ingen oppføringer passer filteret"
            @clear="list.clearAll()"
          />
        </template>
        <template #loading>
          <AdminTableLoading :rows="pagination.pageSize.value" />
        </template>
      </UTable>
    </AdminListView>

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
            <UButton
              variant="ghost"
              @click="
                () => {
                  deleteModal = false
                }
              "
            >
              Avbryt
            </UButton>
            <UButton color="error" @click="handleDelete">Slett</UButton>
          </div>
        </div>
      </template>
    </UModal>
  </div>
</template>
