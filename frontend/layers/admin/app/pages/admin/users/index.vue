<script setup lang="ts">
import type { TableColumn } from '@nuxt/ui'

definePageMeta({
  permission: 'users:view',
  layout: 'admin',
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
            id
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

/** Churches for the filter. Rarely change, so cache-first is enough. */
gql(`
  query AdminUsersPageChurches {
    churches(first: 500) {
      edges {
        node {
          id
          name
        }
      }
    }
  }
`)

const pagination = usePagination({ defaultPageSize: 15 })

const list = useListState({
  pagination,
  filters: { churchId: '' },
})

const { isAuthReady } = useAuthReady()

const { data: churchData } = useAdminUsersPageChurchesQuery({
  pause: computed(() => !isAuthReady.value),
  requestPolicy: 'cache-first',
})

const churchItems = computed(() =>
  (churchData.value?.churches.edges ?? []).map((edge) => ({
    label: edge.node.name,
    value: edge.node.id,
  })),
)

const queryVariables = computed(() => {
  const filter: Record<string, string> = {}
  if (list.debouncedSearch.value) filter.query = list.debouncedSearch.value
  if (list.filters.churchId) filter.churchId = list.filters.churchId

  return {
    ...pagination.variables.value,
    filter: Object.keys(filter).length ? filter : undefined,
  }
})

const { data, fetching, error } = useAdminUsersPageQuery({
  variables: queryVariables,
  pause: computed(() => !isAuthReady.value),
})

watch(
  () => data.value?.users,
  (connection) => pagination.updateConnection(connection),
)

const users = computed(() => data.value?.users.edges.map((edge) => edge.node))

/** Chips need the church's name, not the id that is in the URL. */
const activeFilters = computed(() =>
  list.activeFilters.value.map((filter) => ({
    ...filter,
    label:
      churchItems.value.find((item) => item.value === filter.value)?.label ??
      filter.value,
  })),
)

const columns: TableColumn<
  AdminUsersPageQuery['users']['edges'][number]['node']
>[] = [
  { accessorKey: 'name', header: 'Navn' },
  { accessorKey: 'church.name', header: 'Menighet' },
  { accessorKey: 'roles', header: 'Roller' },
]
</script>

<template>
  <div>
    <h1 class="mb-6 text-3xl">Brukere</h1>
    <AdminErrorState v-if="error" :error />

    <AdminListView
      v-else
      v-model:search="list.search.value"
      :pagination
      :active-filters="activeFilters"
      search-placeholder="Søk etter navn eller e-post…"
      item-label="brukere"
      @clear-filter="list.clearFilter($event as 'churchId')"
      @clear-all="list.clearAll()"
    >
      <template #filters>
        <USelectMenu
          v-model="list.filters.churchId"
          :items="churchItems"
          value-key="value"
          placeholder="Alle menigheter"
          searchable
          class="w-56"
        />
      </template>

      <UTable :data="users" :loading="fetching" :columns>
        <template #name-cell="{ row }">
          <NuxtLink
            :to="{
              name: 'admin-users-userId',
              params: { userId: row.original.id },
            }"
            class="flex flex-col hover:underline"
          >
            <span>{{ row.original.name }}</span>
            <span class="text-dimmed text-xs">{{ row.original.email }}</span>
          </NuxtLink>
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
        <!--
          Distinguishes "no users" from "nothing matched" — the second is the
          common case on a searchable list and needs a way back out.
        -->
        <template #empty>
          <AdminTableEmpty
            :filtered="list.hasActiveQuery.value"
            title="Ingen brukere"
            filtered-title="Ingen brukere passer søket"
            @clear="list.clearAll()"
          />
        </template>
        <template #loading>
          <AdminTableLoading :rows="pagination.pageSize.value" />
        </template>
      </UTable>
    </AdminListView>
  </div>
</template>
