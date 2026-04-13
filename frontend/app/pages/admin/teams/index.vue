<script setup lang="ts">
import type { TableColumn } from '@nuxt/ui'

definePageMeta({
  layout: 'admin',
  middleware: 'superadmin',
})

gql(`
  query AdminTeamsPage($filter: TeamFilter, $first: Int, $after: String, $last: Int, $before: String) {
    teams(
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
          description
          members {
            id
          }
          parentProject {
            id
            name
          }
          superTeam {
            id
            name
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
const { data, fetching, error } = useAdminTeamsPageQuery({
  variables: pagination.variables,
  pause: computed(() => !isAuthReady.value),
})

watch(
  () => data.value?.teams,
  (connection) => {
    pagination.updateConnection(connection)
  },
)

const teams = computed(() => data.value?.teams.edges.map((edge) => edge.node))

const columns: TableColumn<
  AdminTeamsPageQuery['teams']['edges'][number]['node']
>[] = [
  { accessorKey: 'name', header: 'Navn' },
  { accessorKey: 'parentProject.name', id: 'project', header: 'Prosjekt' },
  { accessorKey: 'superTeam.name', id: 'superTeam', header: 'Superlag' },
  { accessorKey: 'members', header: 'Medlemmer' },
  { id: 'actions' },
]
</script>

<template>
  <UContainer class="py-12">
    <h1 class="mb-6 text-3xl">Lag</h1>
    <ErrorState v-if="error" :error />
    <div v-else class="space-y-4">
      <div class="flex items-center justify-between gap-2">
        <RelayPagination v-model:pagination="pagination" />
      </div>
      <UTable :data="teams" :loading="fetching" :columns>
        <template #name-cell="{ row }">
          <div class="flex flex-col">
            <span class="font-medium">{{ row.original.name }}</span>
            <span class="text-dimmed line-clamp-1 text-xs">{{
              row.original.description
            }}</span>
          </div>
        </template>
        <template #superTeam="{ row }">
          <span v-if="row.original.superTeam">{{
            row.original.superTeam.name
          }}</span>
          <span v-else class="text-dimmed">—</span>
        </template>
        <template #members-cell="{ row }">
          <UBadge variant="soft">
            {{ row.original.members.length }}
          </UBadge>
        </template>
        <template #actions-cell="{ row }">
          <div class="flex justify-end">
            <UButton
              variant="ghost"
              :to="{
                name: 'admin-teams-teamId',
                params: { teamId: row.original.id },
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
