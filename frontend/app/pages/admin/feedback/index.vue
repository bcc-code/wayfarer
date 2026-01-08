<script setup lang="ts">
import type { TableColumn } from '@nuxt/ui'

definePageMeta({
  layout: 'admin',
  middleware: 'admin',
})

gql(`
  query AdminFeedbackPage($filter: FeedbackFilter, $first: Int, $after: String, $last: Int, $before: String) {
    feedback(filter: $filter, first: $first, after: $after, last: $last, before: $before) {
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
          message
          canContactMe
          userAgent
          platform
          screenWidth
          screenHeight
          appVersion
          createdAt
          user {
            id
            name
            email
          }
        }
      }
    }
  }
`)

const pagination = usePagination({
  defaultPageSize: 15,
})

const { isAuthReady } = useAuthReady()
const { data, fetching, error, executeQuery } = useAdminFeedbackPageQuery({
  variables: pagination.variables,
  pause: computed(() => !isAuthReady.value),
})

watch(
  () => data.value?.feedback,
  (connection) => {
    pagination.updateConnection(connection)
  },
)

const feedbacks = computed(() =>
  data.value?.feedback.edges.map((edge) => edge.node),
)

type FeedbackNode = NonNullable<typeof feedbacks.value>[number]

const columns: TableColumn<FeedbackNode>[] = [
  { accessorKey: 'user.name', id: 'user', header: 'Bruker' },
  { accessorKey: 'message', header: 'Melding' },
  { accessorKey: 'canContactMe', header: 'Kan kontaktes' },
  { accessorKey: 'platform', id: 'device', header: 'Enhet' },
  { accessorKey: 'appVersion', header: 'Versjon' },
  { accessorKey: 'createdAt', header: 'Innsendingsdato' },
  { id: 'actions' },
]

// Expanded row state
const expandedRows = ref<Set<string>>(new Set())

function toggleRow(id: string) {
  if (expandedRows.value.has(id)) {
    expandedRows.value.delete(id)
  } else {
    expandedRows.value.add(id)
  }
}

// Delete functionality
const { executeMutation: deleteFeedback } = useDeleteFeedbackMutation()
const toast = useToast()
const deleteModal = ref(false)
const feedbackToDelete = ref<string | null>(null)

function confirmDelete(id: string) {
  feedbackToDelete.value = id
  deleteModal.value = true
}

async function handleDelete() {
  if (!feedbackToDelete.value) return

  const result = await deleteFeedback({ id: feedbackToDelete.value })

  if (result.error) {
    toast.add({
      title: 'Kunne ikke slette tilbakemelding',
      description: result.error.message,
      color: 'error',
    })
  } else {
    toast.add({
      title: 'Tilbakemelding slettet',
      color: 'success',
    })
    executeQuery({ requestPolicy: 'network-only' })
  }

  deleteModal.value = false
  feedbackToDelete.value = null
}

const { canDeleteFeedback } = usePermissions()
</script>

<template>
  <UContainer class="py-12">
    <div class="mb-6 flex items-center gap-6">
      <h1 class="text-3xl">Tilbakemeldinger</h1>
    </div>
    <ErrorState v-if="error" :error />
    <div v-else class="space-y-4">
      <div class="flex items-center justify-between gap-2">
        <RelayPagination v-model:pagination="pagination" />
      </div>
      <UTable :data="feedbacks" :loading="fetching" :columns>
        <template #user-cell="{ row }">
          <div class="flex flex-col">
            <NuxtLink
              :to="{
                name: 'admin-users-userId',
                params: { userId: row.original.user.id },
              }"
              class="hover:underline"
            >
              {{ row.original.user.name }}
            </NuxtLink>
            <span class="text-dimmed text-xs">{{
              row.original.user.email
            }}</span>
          </div>
        </template>
        <template #message-cell="{ row }">
          <div class="max-w-md">
            <p
              :class="[
                'text-sm whitespace-pre-wrap',
                expandedRows.has(row.original.id) ? '' : 'line-clamp-4',
              ]"
            >
              {{ row.original.message }}
            </p>
            <button
              v-if="row.original.message.length > 100"
              class="text-primary text-xs hover:underline"
              @click="toggleRow(row.original.id)"
            >
              {{ expandedRows.has(row.original.id) ? 'Vis mindre' : 'Vis mer' }}
            </button>
          </div>
        </template>
        <template #canContactMe-cell="{ row }">
          <UBadge
            :color="row.original.canContactMe ? 'success' : 'neutral'"
            variant="soft"
          >
            {{ row.original.canContactMe ? 'Ja' : 'Nei' }}
          </UBadge>
        </template>
        <template #device-cell="{ row }">
          <div class="text-dimmed space-y-1 text-xs">
            <div v-if="row.original.platform">
              {{ row.original.platform }}
            </div>
            <div v-if="row.original.screenWidth && row.original.screenHeight">
              {{ row.original.screenWidth }}x{{ row.original.screenHeight }}
            </div>
            <div
              v-if="row.original.userAgent"
              class="max-w-xs truncate"
              :title="row.original.userAgent"
            >
              {{ row.original.userAgent }}
            </div>
            <span
              v-if="
                !row.original.platform &&
                !row.original.userAgent &&
                !row.original.screenWidth
              "
            >
              —
            </span>
          </div>
        </template>
        <template #appVersion-cell="{ row }">
          <code v-if="row.original.appVersion" class="text-xs">
            {{ row.original.appVersion }}
          </code>
          <span v-else class="text-dimmed">—</span>
        </template>
        <template #createdAt-cell="{ row }">
          <span class="text-dimmed text-sm">
            {{ formatDate(row.original.createdAt) }}
          </span>
        </template>
        <template #actions-cell="{ row }">
          <div class="flex justify-end">
            <UButton
              v-if="canDeleteFeedback"
              variant="ghost"
              color="error"
              icon="i-lucide-trash-2"
              @click="confirmDelete(row.original.id)"
            />
          </div>
        </template>
      </UTable>
      <UEmpty
        v-if="!fetching && feedbacks?.length === 0"
        title="Ingen tilbakemeldinger ennå"
        description="Tilbakemeldinger fra brukere vises her når de blir sendt inn."
      />
    </div>

    <UModal v-model:open="deleteModal">
      <template #content>
        <div class="p-6">
          <h3 class="mb-4 text-lg font-semibold">Slett tilbakemelding</h3>
          <p class="text-dimmed mb-6">
            Er du sikker på at du vil slette denne tilbakemeldingen? Denne
            handlingen kan ikke angres.
          </p>
          <div class="flex justify-end gap-3">
            <UButton
              color="neutral"
              variant="ghost"
              @click="deleteModal = false"
            >
              Avbryt
            </UButton>
            <UButton color="error" @click="handleDelete">
              Ja, jeg vil slette
            </UButton>
          </div>
        </div>
      </template>
    </UModal>
  </UContainer>
</template>
