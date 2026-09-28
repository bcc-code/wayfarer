<script setup lang="ts">
definePageMeta({
  permission: 'projects:view',
  layout: 'admin',
})

const route = useRoute('admin-projects-projectId-events')
const { canEditProject } = usePermissions()
const canEdit = computed(() => canEditProject(route.params.projectId))

// New: `events/new` and `events/[eventId]` previously had no list linking to
// them and were reachable only by typing the URL.
gql(`
  query AdminProjectEvents($projectId: ID!) {
    events(first: 50, filter: { projectId: $projectId }) {
      edges {
        node {
          id
          name
          description
          startDate
          endDate
        }
      }
    }
  }
`)

const { isAuthReady } = useAuthReady()
const { data, error, fetching } = useAdminProjectEventsQuery({
  variables: computed(() => ({ projectId: route.params.projectId })),
  pause: computed(() => !isAuthReady.value),
})

const events = computed(() => data.value?.events.edges.map((e) => e.node) ?? [])
</script>

<template>
  <div>
    <div class="mb-4 flex items-center justify-between gap-4">
      <h2 class="text-xl">Arrangement</h2>
      <UButton
        v-if="canEdit"
        icon="lucide:plus"
        :to="{
          name: 'admin-projects-projectId-events-new',
          params: { projectId: route.params.projectId },
        }"
      >
        Opprett arrangement
      </UButton>
    </div>

    <AdminErrorState v-if="error" :error />
    <UTable
      v-else
      :data="events"
      :loading="fetching"
      :columns="[
        { accessorKey: 'name', header: 'Navn' },
        { accessorKey: 'description', header: 'Beskrivelse' },
        { accessorKey: 'dates', header: 'Periode' },
        { id: 'actions' },
      ]"
    >
      <template #dates-cell="{ row }">
        {{ formatDateRange(row.original.startDate, row.original.endDate) }}
      </template>
      <template #actions-cell="{ row }">
        <div class="flex justify-end">
          <UButton
            variant="ghost"
            size="sm"
            :to="{
              name: 'admin-projects-projectId-events-eventId',
              params: {
                projectId: route.params.projectId,
                eventId: row.original.id,
              },
            }"
          >
            Rediger
          </UButton>
        </div>
      </template>
      <template #empty>
        <AdminTableEmpty
          icon="lucide:calendar"
          title="Ingen arrangement ennå"
          description="Opprett det første arrangementet for dette prosjektet."
        />
      </template>
      <template #loading>
        <AdminTableLoading />
      </template>
    </UTable>
  </div>
</template>
