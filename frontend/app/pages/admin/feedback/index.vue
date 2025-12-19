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
const { data, fetching, error } = useAdminFeedbackPageQuery({
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
  { accessorKey: 'user.name', id: 'user', header: 'User' },
  { accessorKey: 'message', header: 'Message' },
  { accessorKey: 'canContactMe', header: 'Can Contact' },
  { accessorKey: 'platform', id: 'device', header: 'Device' },
  { accessorKey: 'appVersion', header: 'Version' },
  { accessorKey: 'createdAt', header: 'Submitted' },
]

function formatDate(date: string) {
  return new Date(date).toLocaleDateString('en-US', {
    year: 'numeric',
    month: 'short',
    day: 'numeric',
    hour: '2-digit',
    minute: '2-digit',
  })
}

// Expanded row state
const expandedRows = ref<Set<string>>(new Set())

function toggleRow(id: string) {
  if (expandedRows.value.has(id)) {
    expandedRows.value.delete(id)
  } else {
    expandedRows.value.add(id)
  }
}
</script>

<template>
  <UContainer class="py-12">
    <div class="mb-6 flex items-center gap-6">
      <h1 class="text-3xl">User Feedback</h1>
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
                'text-sm whitespace-normal',
                expandedRows.has(row.original.id) ? '' : 'line-clamp-2',
              ]"
            >
              {{ row.original.message }}
            </p>
            <button
              v-if="row.original.message.length > 25"
              class="text-primary text-xs hover:underline"
              @click="toggleRow(row.original.id)"
            >
              {{
                expandedRows.has(row.original.id) ? 'Show less' : 'Show more'
              }}
            </button>
          </div>
        </template>
        <template #canContactMe-cell="{ row }">
          <UBadge
            :color="row.original.canContactMe ? 'success' : 'neutral'"
            variant="soft"
          >
            {{ row.original.canContactMe ? 'Yes' : 'No' }}
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
      </UTable>
      <UEmpty
        v-if="!fetching && feedbacks?.length === 0"
        title="No feedback yet"
        description="User feedback will appear here once submitted."
      />
    </div>
  </UContainer>
</template>
