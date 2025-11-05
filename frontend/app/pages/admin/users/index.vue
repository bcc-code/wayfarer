<script setup lang="ts">
import type { TableColumn } from '@nuxt/ui'

definePageMeta({
  layout: 'admin',
  middleware: 'admin',
})

gql(`
	query AdminUsersPage {
    users {
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

const { data, fetching, error } = useAdminUsersPageQuery()
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
    <UTable v-else-if="users" :data="users" :loading="fetching" :columns />
  </UContainer>
</template>
