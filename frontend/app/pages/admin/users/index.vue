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
  defaultPageSize: 20,
})
const { data, fetching, error } = useAdminUsersPageQuery({
  variables: pagination.variables,
})

watch(pagination.variables, () => {
  console.log('pagination', unref(pagination))
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
  { accessorKey: 'name', header: 'Name' },
  { accessorKey: 'image', header: 'Image' },
  { accessorKey: 'church.name', header: 'Church' },
  { accessorKey: 'roles', header: 'Roles' },
]
</script>

<template>
  <UContainer class="py-12">
    <h1 class="text-3xl">Users</h1>
    <ErrorState v-if="error" :error />
    <LoadingState v-else-if="fetching" />
    <div v-else-if="users" class="space-y-4">
      <div class="flex justify-end">
        <RelayPagination v-model:pagination="pagination" />
      </div>
      <UTable :data="users" :loading="fetching" :columns />
    </div>
  </UContainer>
</template>
