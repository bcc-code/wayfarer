<script setup lang="ts">
definePageMeta({
  permission: 'projects:view',
  layout: 'admin',
})

const route = useRoute('admin-projects-projectId-challenges')
const { canEditProject } = usePermissions()
const canEdit = computed(() => canEditProject(route.params.projectId))

gql(`
  query AdminProjectChallenges($projectId: ID!) {
    challenges(first: 50, filter: { projectId: $projectId }) {
      edges {
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

const { isAuthReady } = useAuthReady()
const { data, error, fetching } = useAdminProjectChallengesQuery({
  variables: computed(() => ({ projectId: route.params.projectId })),
  pause: computed(() => !isAuthReady.value),
})

const challenges = computed(
  () => data.value?.challenges.edges.map((e) => e.node) ?? [],
)

function challengeType(typename?: string) {
  switch (typename) {
    case 'ExternalChallenge':
      return 'Ekstern'
    case 'QuizChallenge':
      return 'Quiz'
    case 'PluginChallenge':
      return 'Plugin'
    default:
      return 'Enkel'
  }
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

    <LoadingState v-if="fetching" />
    <ErrorState v-else-if="error" :error />
    <UEmpty
      v-else-if="!challenges.length"
      icon="lucide:swords"
      title="Ingen utfordringer ennå"
      description="Opprett den første utfordringen for dette prosjektet."
    />
    <UTable
      v-else
      :data="challenges"
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
    </UTable>
  </div>
</template>
