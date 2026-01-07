<script setup lang="ts">
import type { TableColumn } from '@nuxt/ui'

definePageMeta({
  layout: 'admin',
  middleware: 'admin',
})

gql(`
	query AdminUsersPage($filter: UserFilter, $first: Int, $after: String, $last: Int, $before: String) {
    users(
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
          name
          email
          image
          church {
            name
          }
          roles {
            id
            role
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
const { data, fetching, error } = useAdminUsersPageQuery({
  variables: pagination.variables,
  pause: computed(() => !isAuthReady.value),
})

// Update pagination state when data changes
watch(
  () => data.value?.users,
  (connection) => {
    pagination.updateConnection(connection)
  },
)

const users = computed(() => data.value?.users.edges.map((edge) => edge.node))

const columns: TableColumn<
  AdminUsersPageQuery['users']['edges'][number]['node']
>[] = [
  { accessorKey: 'name', header: 'Navn' },
  { accessorKey: 'church.name', header: 'Menighet' },
  { accessorKey: 'roles', header: 'Roller' },
  { id: 'actions' },
]
</script>

<template>
  <UContainer class="py-12">
    <h1 class="mb-6 text-3xl">Brukere</h1>
    <ErrorState v-if="error" :error />
    <div v-else class="space-y-4">
      <div class="flex items-center justify-between gap-2">
        <RelayPagination v-model:pagination="pagination" />
      </div>
      <UTable :data="users" :loading="fetching" :columns>
        <template #name-cell="{ row }">
          <div class="flex flex-col">
            <span>{{ row.original.name }}</span>
            <span class="text-dimmed text-xs">{{ row.original.email }}</span>
          </div>
        </template>
        <template #roles-cell="{ row }">
          <div class="flex flex-wrap gap-1">
            <UBadge
              v-for="role in row.original.roles"
              :key="role.id"
              variant="soft"
            >
              {{ role.role }}
            </UBadge>
          </div>
        </template>
        <template #actions-cell="{ row }">
          <div class="flex justify-end">
            <UButton
              variant="ghost"
              :to="{
                name: 'admin-users-userId',
                params: { userId: row.original.id },
              }"
            >
              Rediger
            </UButton>
          </div>
        </template>
      </UTable>
    </div>
  </UContainer>
</template>
