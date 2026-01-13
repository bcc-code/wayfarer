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

const { canManageConsents } = usePermissions()

type ConsentRow = AdminConsentsPageQuery['consents'][number]

const columns: TableColumn<ConsentRow>[] = [
  { accessorKey: 'key', header: 'Nøkkel' },
  { accessorKey: 'title', header: 'Tittel' },
  { accessorKey: 'version', header: 'Versjon' },
  { accessorKey: 'publishedAt', header: 'Publiseringsdato' },
  { accessorKey: 'managementType', header: 'Type' },
  { id: 'actions' },
]
</script>

<template>
  <UContainer class="py-12">
    <div class="mb-6 flex flex-col items-start gap-8">
      <h1 class="text-3xl">Samtykker</h1>
      <UButton
        v-if="canManageConsents"
        icon="i-lucide-plus"
        :to="{ name: 'admin-consents-new' }"
      >
        Nytt samtykke
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
          <UBadge v-else variant="soft" color="warning">Utkast</UBadge>
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
              Rediger
            </UButton>
          </div>
        </template>
      </UTable>
      <div
        v-if="consents.length === 0 && !fetching"
        class="text-dimmed py-12 text-center"
      >
        Ingen samtykker funnet. Opprett ditt første samtykke for å komme i gang.
      </div>
    </div>
  </UContainer>
</template>
