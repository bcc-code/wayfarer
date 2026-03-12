<script setup lang="ts">
import type { TableColumn } from '@nuxt/ui'
import { ForwardDestination } from '~/api/generated'

definePageMeta({
  layout: 'admin',
  middleware: 'admin',
})

// Subscribe to admin feedback notifications
const { subscribeAdmin, isAuthenticated } = useFirestoreSync()
watch(
  isAuthenticated,
  (authenticated) => {
    if (authenticated) {
      subscribeAdmin('feedback')
    }
  },
  { immediate: true },
)

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
          locale
          projectId
          timezone
          contextUrl
          tags
          createdAt
          handledAt
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

gql(`
  query FeedbackTags {
    feedbackTags
  }
`)

gql(`
  query FeedbackPlatforms {
    feedbackPlatforms
  }
`)

gql(`
  mutation UpdateFeedbackTags($feedbackId: ID!, $tags: [String!]!) {
    updateFeedbackTags(feedbackId: $feedbackId, tags: $tags) {
      id
      tags
    }
  }
`)

const pagination = usePagination({
  defaultPageSize: 15,
})

// Filters
const selectedTags = ref<string[]>([])
const selectedPlatform = ref<string | undefined>()
const handledFilter = ref<boolean | undefined>()

const filter = computed(() => ({
  tags: selectedTags.value.length > 0 ? selectedTags.value : undefined,
  handled: handledFilter.value,
  platform: selectedPlatform.value,
}))

const queryVariables = computed(() => ({
  ...pagination.variables.value,
  filter: filter.value,
}))

const { isAuthReady } = useAuthReady()
const { data, fetching, error, executeQuery } = useAdminFeedbackPageQuery({
  variables: queryVariables,
  pause: computed(() => !isAuthReady.value),
})

const { data: tagsData, executeQuery: refetchTags } = useFeedbackTagsQuery({
  pause: computed(() => !isAuthReady.value),
})

const { data: platformsData } = useFeedbackPlatformsQuery({
  pause: computed(() => !isAuthReady.value),
})

// Refresh when Firestore notifies of updates
useFirestoreRefresh(['AdminFeedbackPageDocument'], () => {
  executeQuery({ requestPolicy: 'network-only' })
})

// Reset pagination when filter changes
watch([selectedTags, selectedPlatform, handledFilter], () => {
  pagination.reset()
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

const allUniqueTags = computed(() => tagsData.value?.feedbackTags ?? [])
const allPlatforms = computed(
  () => platformsData.value?.feedbackPlatforms ?? [],
)

const handledOptions = [
  { label: 'Alle', value: undefined },
  { label: 'Ubehandlet', value: false },
  { label: 'Behandlet', value: true },
]

type FeedbackNode = NonNullable<typeof feedbacks.value>[number]

const columns: TableColumn<FeedbackNode>[] = [
  { accessorKey: 'user.name', id: 'user', header: 'Bruker' },
  { accessorKey: 'message', header: 'Melding' },
  { accessorKey: 'tags', id: 'tags', header: 'Tags' },
  { accessorKey: 'createdAt', header: 'Dato' },
  { id: 'actions', header: '' },
]

// Expanded message state
const expandedMessages = ref<Set<string>>(new Set())

function toggleMessage(id: string) {
  if (expandedMessages.value.has(id)) {
    expandedMessages.value.delete(id)
  } else {
    expandedMessages.value.add(id)
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

// Forward to support functionality
const { executeMutation: forwardFeedback } = useForwardFeedbackToDeskMutation()

const forwardDestinationLabels: Record<ForwardDestination, string> = {
  [ForwardDestination.BccMediaSupport]: 'BCC Media Support',
  [ForwardDestination.SsfTicket]: 'Skjulte Skatter',
}

async function handleForward(id: string, destination: ForwardDestination) {
  const result = await forwardFeedback({ feedbackId: id, destination })

  if (result.error) {
    toast.add({
      title: 'Kunne ikke videresende',
      description: result.error.message,
      color: 'error',
    })
  } else {
    toast.add({
      title: `Videresendt til ${forwardDestinationLabels[destination]}`,
      color: 'success',
    })
    executeQuery({ requestPolicy: 'network-only' })
  }
}

// Mark as handled functionality
const { executeMutation: markHandled } = useMarkFeedbackHandledMutation()

async function handleMarkHandled(id: string) {
  const result = await markHandled({ feedbackId: id })

  if (result.error) {
    toast.add({
      title: 'Kunne ikke markere som behandlet',
      description: result.error.message,
      color: 'error',
    })
  } else {
    toast.add({
      title: 'Markert som behandlet',
      color: 'success',
    })
    executeQuery({ requestPolicy: 'network-only' })
  }
}

const { canDeleteFeedback, canForwardFeedback } = usePermissions()

// Update tags functionality
const { executeMutation: updateFeedbackTags } = useUpdateFeedbackTagsMutation()

async function handleUpdateTags(feedbackId: string, tags: string[]) {
  const result = await updateFeedbackTags({
    feedbackId,
    tags,
  })

  if (result.error) {
    toast.add({
      title: 'Kunne ikke oppdatere tags',
      description: result.error.message,
      color: 'error',
    })
  } else {
    refetchTags({ requestPolicy: 'network-only' })
  }
}
</script>

<template>
  <UContainer class="py-12">
    <div class="mb-6 flex items-center gap-6">
      <h1 class="text-3xl">Tilbakemeldinger</h1>
    </div>
    <ErrorState v-if="error" :error />
    <div v-else class="space-y-4">
      <div class="flex items-center justify-between gap-2">
        <USelectMenu
          v-model="selectedTags"
          :items="allUniqueTags"
          multiple
          placeholder="Filtrer etter tags..."
          icon="lucide:tag"
          size="sm"
          class="min-w-56 max-w-80"
          :ui="{ base: 'flex-wrap' }"
        >
          <template #default="{ modelValue: _tags }">
            <UBadge
              v-for="tag in _tags"
              :key="tag"
              :label="tag"
              size="sm"
              variant="subtle"
              color="neutral"
            />
          </template>
        </USelectMenu>
        <USelectMenu
          v-model="selectedPlatform"
          :items="allPlatforms"
          placeholder="Plattform..."
          icon="lucide:monitor-smartphone"
          size="sm"
          class="min-w-40"
        />
        <USelectMenu
          :model-value="
            handledOptions.find((o) => o.value === handledFilter)?.value
          "
          :items="handledOptions"
          value-key="value"
          label-key="label"
          placeholder="Status..."
          icon="lucide:check-circle"
          size="sm"
          class="min-w-40"
          @update:model-value="handledFilter = $event"
        />
        <UButton
          v-if="
            selectedTags.length > 0 ||
            selectedPlatform !== undefined ||
            handledFilter !== undefined
          "
          variant="ghost"
          size="sm"
          color="neutral"
          label="Nullstill"
          @click="
            () => {
              selectedTags = []
              selectedPlatform = undefined
              handledFilter = undefined
            }
          "
        />
        <RelayPagination v-model:pagination="pagination" class="ml-auto" />
      </div>
      <UTable :data="feedbacks" :loading="fetching" :columns>
        <template #user-cell="{ row }">
          <NuxtLink
            :to="{
              name: 'admin-users-userId',
              params: { userId: row.original.user.id },
            }"
            class="flex flex-col group"
          >
            <span class="group-hover:underline">
              {{ row.original.user.name }}
            </span>
            <span class="text-dimmed text-xs">
              {{ row.original.user.email }}
            </span>
          </NuxtLink>
        </template>
        <template #tags-cell="{ row }">
          <UInputTags
            :model-value="row.original.tags"
            placeholder="Legg til..."
            size="xs"
            color="neutral"
            class="w-full"
            @update:model-value="handleUpdateTags(row.original.id, $event)"
          />
        </template>
        <template #message-cell="{ row }">
          <div class="max-w-lg">
            <p
              :class="[
                'text-sm whitespace-pre-wrap',
                expandedMessages.has(row.original.id) ? '' : 'line-clamp-4',
              ]"
            >
              {{ row.original.message }}
            </p>
            <button
              v-if="row.original.message.length > 100"
              class="text-primary text-xs hover:underline mt-1"
              @click="toggleMessage(row.original.id)"
            >
              {{
                expandedMessages.has(row.original.id) ? 'Vis mindre' : 'Les mer'
              }}
            </button>
          </div>
        </template>
        <template #createdAt-cell="{ row }">
          <span class="text-dimmed text-sm">
            {{ formatDateTime(row.original.createdAt) }}
          </span>
        </template>
        <template #actions-cell="{ row }">
          <div class="flex justify-end items-center gap-1">
            <UDropdownMenu
              v-if="canForwardFeedback && !row.original.handledAt"
              :items="
                Object.values(ForwardDestination).map((dest) => ({
                  label: forwardDestinationLabels[dest],
                  onSelect: () => handleForward(row.original.id, dest),
                }))
              "
            >
              <UButton
                variant="soft"
                color="neutral"
                size="sm"
                icon="lucide:send-horizontal"
                label="Videresend"
                trailing-icon="lucide:chevron-down"
              />
            </UDropdownMenu>
            <UButton
              v-if="canForwardFeedback && !row.original.handledAt"
              variant="soft"
              color="neutral"
              size="sm"
              label="Behandle"
              icon="lucide:check"
              @click="handleMarkHandled(row.original.id)"
            />
            <UBadge
              v-else-if="canForwardFeedback && row.original.handledAt"
              variant="soft"
              color="success"
              label="Behandlet"
            />
            <UPopover>
              <UButton
                variant="ghost"
                color="neutral"
                size="sm"
                icon="i-lucide-info"
              />
              <template #content>
                <div
                  class="p-3 grid grid-cols-[auto_1fr] gap-x-3 gap-y-1 text-xs min-w-56"
                >
                  <span class="text-dimmed">Kan kontaktes:</span>
                  <span>{{ row.original.canContactMe ? 'Ja' : 'Nei' }}</span>
                  <template v-if="row.original.locale">
                    <span class="text-dimmed">Språk:</span>
                    <span>{{ row.original.locale }}</span>
                  </template>
                  <template v-if="row.original.timezone">
                    <span class="text-dimmed">Tidssone:</span>
                    <span>{{ row.original.timezone }}</span>
                  </template>
                  <template v-if="row.original.projectId">
                    <span class="text-dimmed">Prosjekt:</span>
                    <span>{{ row.original.projectId }}</span>
                  </template>
                  <template v-if="row.original.platform">
                    <span class="text-dimmed">Plattform:</span>
                    <span>{{ row.original.platform }}</span>
                  </template>
                  <template
                    v-if="row.original.screenWidth && row.original.screenHeight"
                  >
                    <span class="text-dimmed">Skjerm:</span>
                    <span
                      >{{ row.original.screenWidth }}x{{
                        row.original.screenHeight
                      }}</span
                    >
                  </template>
                  <template v-if="row.original.userAgent">
                    <span class="text-dimmed">Nettleser:</span>
                    <span>{{ parseUserAgent(row.original.userAgent) }}</span>
                  </template>
                  <template v-if="row.original.appVersion">
                    <span class="text-dimmed">Versjon:</span>
                    <code>{{ row.original.appVersion }}</code>
                  </template>
                  <template v-if="row.original.contextUrl">
                    <span class="text-dimmed">Side:</span>
                    <code class="text-xs">{{ row.original.contextUrl }}</code>
                  </template>
                </div>
              </template>
            </UPopover>
            <UButton
              v-if="canDeleteFeedback"
              variant="ghost"
              size="sm"
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
          <Icon name="lucide:triangle-alert" class="text-error size-8" />
          <h3 class="my-2 text-lg font-semibold">Slett tilbakemelding</h3>
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
