<script setup lang="ts">
import { ChallengeType } from '~/api/generated'

definePageMeta({
  permission: 'projects:view',
  layout: 'admin',
})

const route = useRoute('admin-projects-projectId-challenges')
const { canEditProject } = usePermissions()
const canEdit = computed(() => canEditProject(route.params.projectId))

// Was `first: 50` with no pagination: a project's 51st challenge simply did
// not appear, with nothing on the page to say so.
gql(`
  query AdminProjectChallenges(
    $filter: ChallengeFilter
    $first: Int
    $after: String
    $last: Int
    $before: String
  ) {
    challenges(
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
          __typename
          id
          name
          description
          imageObject {
            ...ImageFields
          }
        }
      }
    }
  }
`)

const pagination = usePagination({ defaultPageSize: 15 })

/**
 * `ChallengeFilter` has no free-text field, so this filters by type. `eventId`
 * is deliberately not offered: project events are barely used (see the note's
 * scope decisions), so it would be a permanently empty control.
 */
const list = useListState({
  pagination,
  filters: { challengeType: '' },
})

const challengeTypeLabels: Record<string, string> = {
  [ChallengeType.Simple]: 'Enkel',
  [ChallengeType.Quiz]: 'Quiz',
  [ChallengeType.External]: 'Ekstern',
  [ChallengeType.Plugin]: 'Plugin',
}

const challengeTypeItems = Object.values(ChallengeType).map((type) => ({
  label: challengeTypeLabels[type] ?? String(type),
  value: String(type),
}))

const { isAuthReady } = useAuthReady()
const { data, error, fetching } = useAdminProjectChallengesQuery({
  variables: computed(() => ({
    ...pagination.variables.value,
    filter: {
      projectId: route.params.projectId,
      ...(list.filters.challengeType
        ? { challengeType: list.filters.challengeType as ChallengeType }
        : {}),
    },
  })),
  pause: computed(() => !isAuthReady.value),
})

watch(
  () => data.value?.challenges,
  (connection) => pagination.updateConnection(connection),
)

const challenges = computed(
  () => data.value?.challenges.edges.map((e) => e.node) ?? [],
)

const activeFilters = computed(() =>
  list.activeFilters.value.map((filter) => ({
    ...filter,
    label: `Type: ${challengeTypeLabels[filter.value] ?? filter.value}`,
  })),
)

/**
 * The table has a `__typename` and the filter has a `ChallengeType`; both need
 * the same four labels, so the typename maps to the enum rather than carrying a
 * second copy of the labels that could drift from it.
 */
const typenameToChallengeType: Record<string, ChallengeType> = {
  SimpleChallenge: ChallengeType.Simple,
  QuizChallenge: ChallengeType.Quiz,
  ExternalChallenge: ChallengeType.External,
  PluginChallenge: ChallengeType.Plugin,
}

function challengeType(typename?: string) {
  const type = typename ? typenameToChallengeType[typename] : undefined
  // Unknown typenames fell back to "Enkel" before; kept, so a new challenge
  // kind added server-side degrades rather than rendering blank.
  return challengeTypeLabels[type ?? ChallengeType.Simple]
}
</script>

<template>
  <div>
    <div class="mb-4 flex items-center justify-between gap-4">
      <h2 class="text-xl">Utfordringer</h2>
      <UButton
        v-if="canEdit"
        icon="lucide:plus"
        :to="{
          name: 'admin-projects-projectId-challenges-new',
          params: { projectId: route.params.projectId },
        }"
      >
        Opprett utfordring
      </UButton>
    </div>

    <AdminErrorState v-if="error" :error />

    <AdminListView
      v-else
      :pagination
      :active-filters="activeFilters"
      :searchable="false"
      item-label="utfordringer"
      @clear-filter="list.clearFilter($event as 'challengeType')"
      @clear-all="list.clearAll()"
    >
      <template #filters>
        <USelect
          v-model="list.filters.challengeType"
          :items="challengeTypeItems"
          value-key="value"
          placeholder="Alle typer"
          icon="lucide:filter"
          class="w-48"
        />
      </template>

      <UTable
        :data="challenges"
        :loading="fetching"
        :columns="[
          { accessorKey: 'imageObject' },
          { accessorKey: 'name' },
          { accessorKey: 'description' },
          { accessorKey: 'type', header: 'Type' },
          { id: 'actions' },
        ]"
      >
        <template #imageObject-cell="{ row }">
          <img
            v-if="row.original.imageObject?.url"
            :src="row.original.imageObject.url"
            height="32"
            width="32"
            class="bg-muted size-8 rounded"
          />
        </template>
        <template #type-cell="{ row }">
          {{ challengeType(row.original.__typename) }}
        </template>
        <template #actions-cell="{ row }">
          <div class="flex justify-end">
            <UButton
              variant="ghost"
              size="sm"
              :to="{
                name: 'admin-projects-projectId-challenges-challengeId',
                params: {
                  projectId: route.params.projectId,
                  challengeId: row.original.id,
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
                  ? 'Ingen utfordringer av denne typen'
                  : 'Ingen utfordringer ennå'
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
