<script setup lang="ts">
definePageMeta({
  permission: 'consents:view',
  layout: 'admin',
})

// Scoped like the consent gate: this project's consents plus the global ones.
gql(`
  query AdminProjectConsentsPage($projectId: ID!, $includeArchived: Boolean) {
    adminConsents(projectId: $projectId, includeArchived: $includeArchived) {
      id
      key
      version
      title
      publishedAt
      archivedAt
      managementType
      isRequired
      project {
        id
        name
      }
    }
  }
`)

const route = useRoute('admin-projects-projectId-consents')

const showArchived = ref(false)

const { isAuthReady } = useAuthReady()
const {
  data,
  fetching,
  error,
  executeQuery: refetch,
} = useAdminProjectConsentsPageQuery({
  variables: computed(() => ({
    projectId: route.params.projectId,
    includeArchived: showArchived.value,
  })),
  pause: computed(() => !isAuthReady.value),
})

const consents = computed(() => data.value?.adminConsents ?? [])

const { canManageConsents } = usePermissions()
</script>

<template>
  <div>
    <div class="mb-6 flex flex-wrap items-start justify-between gap-4">
      <div>
        <h1 class="text-3xl">Samtykker</h1>
        <p class="text-muted mt-1 text-sm">
          Samtykker for dette prosjektet, og de som gjelder alle prosjekter.
        </p>
      </div>
      <div class="flex items-center gap-4">
        <USwitch v-model="showArchived" label="Vis arkiverte" />
        <UButton
          v-if="canManageConsents"
          icon="i-lucide-plus"
          :to="{
            name: 'admin-consents-new',
            query: { projectId: route.params.projectId },
          }"
        >
          Nytt samtykke
        </UButton>
      </div>
    </div>
    <AdminErrorState v-if="error" :error />
    <AdminConsentSections
      v-else
      :consents
      :fetching
      show-project
      :show-archived="showArchived"
      :can-manage="canManageConsents"
      published-empty-description="Verken dette prosjektet eller fellessamtykkene har noe publisert ennå."
      @refresh="refetch({ requestPolicy: 'network-only' })"
    />
  </div>
</template>
