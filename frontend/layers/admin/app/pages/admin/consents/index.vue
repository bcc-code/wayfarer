<script setup lang="ts">
definePageMeta({
  permission: 'consents:view',
  layout: 'admin',
})

// `adminConsents`, not `consents`: the latter is published-only, so it hides drafts.
gql(`
  query AdminConsentsPage($includeArchived: Boolean) {
    adminConsents(includeArchived: $includeArchived) {
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

const showArchived = ref(false)

const { isAuthReady } = useAuthReady()
const {
  data,
  fetching,
  error,
  executeQuery: refetch,
} = useAdminConsentsPageQuery({
  variables: computed(() => ({ includeArchived: showArchived.value })),
  pause: computed(() => !isAuthReady.value),
})

const consents = computed(() => data.value?.adminConsents ?? [])

const { canManageConsents } = usePermissions()
</script>

<template>
  <div>
    <div class="mb-6 flex flex-wrap items-center justify-between gap-4">
      <h1 class="text-3xl">Samtykker</h1>
      <div class="flex items-center gap-4">
        <USwitch v-model="showArchived" label="Vis arkiverte" />
        <UButton
          v-if="canManageConsents"
          icon="i-lucide-plus"
          :to="{ name: 'admin-consents-new' }"
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
      @refresh="refetch({ requestPolicy: 'network-only' })"
    />
  </div>
</template>
