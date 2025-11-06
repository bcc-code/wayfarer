<script setup lang="ts">
import { UAvatar } from '#components'
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
const { data, fetching, error } = useAdminUsersPageQuery({
  variables: pagination.variables,
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
  { accessorKey: 'church.name', header: 'Church' },
  { accessorKey: 'roles', header: 'Roles' },
  { id: 'actions' },
]

const initials = (name: string) => {
  const splitNames = name.split(' ')
  return splitNames
    .filter(Boolean)
    .map((name) => name[0])
    .join('')
}
</script>

<template>
  <UContainer class="py-12">
    <h1 class="text-3xl mb-6">Users</h1>
    <ErrorState v-if="error" :error />
    <div v-else class="space-y-4">
      <div class="flex justify-end items-center gap-2">
        <RelayPagination v-model:pagination="pagination" />
      </div>
      <UTable :data="users" :loading="fetching" :columns>
        <template #name-cell="{ row }">
          <div class="flex items-center gap-3">
            <UAvatar
              :src="row.original.image ?? ''"
              :text="initials(row.original.name)"
              size="sm"
            />
            <div class="flex flex-col">
              <span>{{ row.original.name }}</span>
              <span class="text-xs text-dimmed">{{ row.original.email }}</span>
            </div>
          </div>
        </template>
        <template #roles-cell="{ row }">
          <div class="flex gap-1 flex-wrap">
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
              Edit
            </UButton>
          </div>
        </template>
      </UTable>
    </div>
  </UContainer>
</template>
