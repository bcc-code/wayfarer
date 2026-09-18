<script setup lang="ts">
definePageMeta({
  permission: 'projects:view',
  layout: 'admin',
})

const route = useRoute('admin-projects-projectId-superteams')
const { canEditProject } = usePermissions()
const { isSuperAdmin } = useAuth()
const canEdit = computed(() => canEditProject(route.params.projectId))

gql(`
  query AdminProjectSuperteams($projectId: ID!) {
    superteams(first: 50, filter: { projectId: $projectId }) {
      edges {
        node {
          id
          name
          description
          color
          imageObject {
            ...ImageFields
          }
          teams {
            id
          }
        }
      }
    }
  }
`)

const { isAuthReady } = useAuthReady()
const { data, error, fetching } = useAdminProjectSuperteamsQuery({
  variables: computed(() => ({ projectId: route.params.projectId })),
  pause: computed(() => !isAuthReady.value),
})

const superteams = computed(
  () => data.value?.superteams.edges.map((e) => e.node) ?? [],
)
</script>

<template>
  <div>
    <div class="mb-4 flex items-center justify-between gap-4">
      <h2 class="text-xl">Superlag</h2>
      <div class="flex gap-2">
        <!-- The ladder-to-heaven distribution tool. Its own route since it is a
             separate workflow, and superadmin-only. -->
        <UButton
          v-if="isSuperAdmin"
          variant="soft"
          icon="lucide:shuffle"
          :to="{
            name: 'admin-projects-projectId-superteams-distribute',
            params: { projectId: route.params.projectId },
          }"
        >
          LADD-fordeling
        </UButton>
        <UButton
          v-if="canEdit"
          icon="lucide:plus"
          :to="{
            name: 'admin-projects-projectId-superteams-new',
            params: { projectId: route.params.projectId },
          }"
        >
          Opprett superlag
        </UButton>
      </div>
    </div>

    <AdminLoadingState v-if="fetching" />
    <AdminErrorState v-else-if="error" :error />
    <UEmpty
      v-else-if="!superteams.length"
      icon="lucide:users"
      title="Ingen superlag ennå"
      description="Opprett det første superlaget for dette prosjektet."
    />
    <UTable
      v-else
      :data="superteams"
      :columns="[
        { accessorKey: 'color', header: 'Farge' },
        { accessorKey: 'imageObject', header: 'Bilde' },
        { accessorKey: 'name', header: 'Navn' },
        { accessorKey: 'teams', header: 'Lag' },
        { id: 'actions' },
      ]"
    >
      <template #color-cell="{ row }">
        <div
          v-if="row.original.color"
          class="size-6 rounded-full border"
          :style="{ backgroundColor: row.original.color }"
        />
      </template>
      <template #imageObject-cell="{ row }">
        <img
          v-if="row.original.imageObject?.url"
          :src="row.original.imageObject.url"
          height="32"
          width="32"
          class="bg-muted size-8 rounded"
        />
      </template>
      <template #teams-cell="{ row }">
        {{ row.original.teams.length }} lag
      </template>
      <template #actions-cell="{ row }">
        <div class="flex justify-end">
          <UButton
            variant="ghost"
            size="sm"
            :to="{
              name: 'admin-projects-projectId-superteams-superTeamId',
              params: {
                projectId: route.params.projectId,
                superTeamId: row.original.id,
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
