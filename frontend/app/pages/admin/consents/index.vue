<script setup lang="ts">
import type { TableColumn } from '@nuxt/ui'

definePageMeta({
  layout: 'admin',
  middleware: 'admin',
})

gql(`
  query AdminConsentsPage {
    consents {
      id
      key
      version
      title
      shortText
      publishedAt
      managementType
      managedBy
    }
  }
`)

const { isAuthReady } = useAuthReady()
const { data, fetching, error } = useAdminConsentsPageQuery({
  pause: computed(() => !isAuthReady.value),
})

const consents = computed(() => data.value?.consents ?? [])

type ConsentRow = AdminConsentsPageQuery['consents'][number]

const columns: TableColumn<ConsentRow>[] = [
  { accessorKey: 'key', header: 'Key' },
  { accessorKey: 'title', header: 'Title' },
  { accessorKey: 'version', header: 'Version' },
  { accessorKey: 'publishedAt', header: 'Published' },
  { accessorKey: 'managementType', header: 'Type' },
  { id: 'actions' },
]
</script>

<template>
  <UContainer class="py-12">
    <div class="mb-6 flex items-center justify-between">
      <h1 class="text-3xl">Consents</h1>
      <UButton icon="i-lucide-plus" :to="{ name: 'admin-consents-new' }">
        New Consent
      </UButton>
    </div>
    <ErrorState v-if="error" :error />
    <div v-else class="space-y-4">
      <UTable :data="consents" :loading="fetching" :columns>
        <template #key-cell="{ row }">
          <code class="bg-background-indent rounded px-2 py-1 text-sm">
            {{ row.original.key }}
          </code>
        </template>
        <template #title-cell="{ row }">
          <span class="font-medium">{{ row.original.title }}</span>
        </template>
        <template #version-cell="{ row }">
          <UBadge variant="soft"> v{{ row.original.version }} </UBadge>
        </template>
        <template #publishedAt-cell="{ row }">
          <span v-if="row.original.publishedAt">
            {{ formatDateTime(row.original.publishedAt) }}
          </span>
          <UBadge v-else variant="soft" color="warning"> Draft </UBadge>
        </template>
        <template #managementType-cell="{ row }">
          <UBadge
            :color="
              row.original.managementType === 'LOCAL' ? 'primary' : 'neutral'
            "
            variant="soft"
          >
            {{ row.original.managementType }}
          </UBadge>
        </template>
        <template #actions-cell="{ row }">
          <div class="flex justify-end">
            <UButton
              variant="ghost"
              :to="{
                name: 'admin-consents-consentId',
                params: { consentId: row.original.id },
              }"
            >
              Edit
            </UButton>
          </div>
        </template>
      </UTable>
      <div
        v-if="consents.length === 0 && !fetching"
        class="text-dimmed py-12 text-center"
      >
        No consents found. Create your first consent to get started.
      </div>
    </div>
  </UContainer>
</template>
