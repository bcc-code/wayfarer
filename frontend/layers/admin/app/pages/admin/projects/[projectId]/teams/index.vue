<script setup lang="ts">
import type { TableColumn } from '@nuxt/ui'

definePageMeta({
  permission: 'teams:view',
  layout: 'admin',
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

/** Superteams for the filter. Few per project, so one unpaginated read. */
gql(`
  query AdminTeamsPageSuperTeams($projectId: ID!) {
    superteams(first: 100, filter: { projectId: $projectId }) {
      edges {
        node {
          id
          name
        }
      }
    }
  }
`)

const route = useRoute('admin-projects-projectId-teams')

const pagination = usePagination({
  defaultPageSize: 15,
})

/**
 * `TeamFilter` has no free-text field, so this list filters by superteam only.
 * `NO_SUPER_TEAM` is a sentinel rather than a second control: the API exposes
 * `noSuperTeam: Boolean` alongside `superTeamId`, and folding both into one
 * select gives three states — all / a specific superteam / unassigned — which
 * is the question worth asking before running the distribution tool.
 */
const NO_SUPER_TEAM = '__none__'

const list = useListState({
  pagination,
  filters: { superTeam: '' },
})

const { isAuthReady } = useAuthReady()

const { data: superTeamData } = useAdminTeamsPageSuperTeamsQuery({
  variables: computed(() => ({ projectId: route.params.projectId })),
  pause: computed(() => !isAuthReady.value),
  requestPolicy: 'cache-first',
})

const superTeamItems = computed(() => [
  { label: 'Uten superlag', value: NO_SUPER_TEAM },
  ...(superTeamData.value?.superteams.edges ?? []).map((edge) => ({
    label: edge.node.name,
    value: edge.node.id,
  })),
])

const teamFilter = computed(() => {
  const selected = list.filters.superTeam
  return {
    projectId: route.params.projectId,
    ...(selected === NO_SUPER_TEAM
      ? { noSuperTeam: true }
      : selected
        ? { superTeamId: selected }
        : {}),
  }
})

const { data, fetching, error } = useAdminTeamsPageQuery({
  // Merged with the pagination cursor rather than replacing it. Computed so the
  // list repoints when the project switcher changes the param — this route no
  // longer remounts, it lives under the persistent [projectId].vue parent.
  variables: computed(() => ({
    ...pagination.variables.value,
    filter: teamFilter.value,
  })),
  pause: computed(() => !isAuthReady.value),
})

const activeFilters = computed(() =>
  list.activeFilters.value.map((filter) => ({
    ...filter,
    label: `Superlag: ${
      superTeamItems.value.find((item) => item.value === filter.value)?.label ??
      filter.value
    }`,
  })),
)

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
  { accessorKey: 'superTeam.name', id: 'superTeam', header: 'Superlag' },
  { accessorKey: 'members', header: 'Medlemmer' },
  { id: 'actions' },
]
</script>

<template>
  <div>
    <h1 class="mb-6 text-3xl">Lag</h1>
    <AdminErrorState v-if="error" :error />
    <AdminListView
      v-else
      :pagination
      :active-filters="activeFilters"
      :searchable="false"
      item-label="lag"
      @clear-filter="list.clearFilter($event as 'superTeam')"
      @clear-all="list.clearAll()"
    >
      <template #filters>
        <USelectMenu
          v-model="list.filters.superTeam"
          :items="superTeamItems"
          value-key="value"
          placeholder="Alle superlag"
          icon="lucide:users"
          class="w-56"
        />
      </template>

      <UTable :data="teams" :loading="fetching" :columns>
        <template #name-cell="{ row }">
          <div class="flex flex-col">
            <span class="font-medium">{{ row.original.name }}</span>
            <span class="text-dimmed line-clamp-1 text-xs">{{
              row.original.description
            }}</span>
          </div>
        </template>
        <template #superTeam-cell="{ row }">
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
                name: 'admin-projects-projectId-teams-teamId',
                params: {
                  projectId: route.params.projectId,
                  teamId: row.original.id,
                },
              }"
            >
              Rediger
            </UButton>
          </div>
        </template>
        <template #empty>
          <div class="py-6 text-center">
            <p class="text-muted text-sm">
              {{
                activeFilters.length
                  ? 'Ingen lag passer filteret'
                  : 'Ingen lag i dette prosjektet'
              }}
            </p>
            <UButton
              v-if="activeFilters.length"
              variant="link"
              size="sm"
              @click="list.clearAll()"
            >
              Nullstill filter
            </UButton>
          </div>
        </template>
      </UTable>
    </AdminListView>
  </div>
</template>
