<script setup lang="ts">
import type { TableColumn } from '@nuxt/ui'

const props = withDefaults(
  defineProps<{
    consents: AdminConsentRow[]
    fetching?: boolean
    showProject?: boolean
    /** Off for the drafts table, where every row has the same status. */
    showStatus?: boolean
    canManage?: boolean
    emptyTitle?: string
    emptyDescription?: string
  }>(),
  {
    fetching: false,
    showProject: false,
    showStatus: true,
    canManage: false,
    emptyTitle: 'Ingen samtykker funnet',
    emptyDescription: 'Opprett ditt første samtykke for å komme i gang.',
  },
)

const emit = defineEmits<{ refresh: [] }>()

const columns = computed<TableColumn<AdminConsentRow>[]>(() => [
  { accessorKey: 'key', header: 'Nøkkel' },
  { accessorKey: 'title', header: 'Tittel' },
  { accessorKey: 'version', header: 'Versjon' },
  ...(props.showStatus
    ? [{ accessorKey: 'publishedAt', header: 'Status' } as const]
    : []),
  ...(props.showProject
    ? [{ accessorKey: 'project', header: 'Prosjekt' } as const]
    : []),
  { accessorKey: 'managementType', header: 'Type' },
  { id: 'actions' },
])

const { confirm } = useConfirm()
const toast = useToast()
const { executeMutation: archiveConsent } = useArchiveConsentMutation()
const { executeMutation: unarchiveConsent } = useUnarchiveConsentMutation()

async function handleArchive(consent: AdminConsentRow) {
  const confirmed = await confirm({
    title: `Arkivere "${consent.title}"?`,
    description: CONSENT_ARCHIVE_DESCRIPTION,
    confirmLabel: 'Arkiver',
    icon: 'lucide:archive',
    color: 'primary',
  })
  if (!confirmed) return

  const result = await archiveConsent({ id: consent.id })

  if (result.error) {
    toast.add({
      title: 'Kunne ikke arkivere samtykke',
      description: result.error.message,
      color: 'error',
    })
    return
  }

  toast.add({ title: 'Samtykke arkivert', color: 'success' })
  emit('refresh')
}

// No confirmation: the button next to it undoes this.
async function handleUnarchive(consent: AdminConsentRow) {
  const result = await unarchiveConsent({ id: consent.id })

  if (result.error) {
    toast.add({
      title: 'Kunne ikke gjenopprette samtykke',
      description: result.error.message,
      color: 'error',
    })
    return
  }

  toast.add({ title: 'Samtykke gjenopprettet', color: 'success' })
  emit('refresh')
}
</script>

<template>
  <UTable :data="consents" :loading="fetching" :columns="columns">
    <template #key-cell="{ row }">
      <code class="bg-elevated rounded px-2 py-1 text-sm">
        {{ row.original.key }}
      </code>
    </template>
    <template #title-cell="{ row }">
      <div class="flex items-center gap-2">
        <span class="font-medium">{{ row.original.title }}</span>
        <UBadge v-if="row.original.isRequired" variant="soft" color="error">
          Påkrevd
        </UBadge>
      </div>
    </template>
    <template #version-cell="{ row }">
      <UBadge variant="soft">v{{ row.original.version }}</UBadge>
    </template>
    <template #publishedAt-cell="{ row }">
      <AdminConsentStatusBadge :consent="row.original" />
    </template>
    <template #project-cell="{ row }">
      <NuxtLink
        v-if="row.original.project"
        class="admin-link"
        :to="{
          name: 'admin-projects-projectId',
          params: { projectId: row.original.project.id },
        }"
      >
        {{ row.original.project.name }}
      </NuxtLink>
      <span v-else class="text-muted">Alle prosjekter</span>
    </template>
    <template #managementType-cell="{ row }">
      <UBadge
        variant="soft"
        :color="consentManagementTypeColor(row.original.managementType)"
      >
        {{ consentManagementTypeLabel(row.original.managementType) }}
      </UBadge>
    </template>
    <template #actions-cell="{ row }">
      <div class="flex justify-end gap-1">
        <UButton
          icon="lucide:pencil"
          variant="ghost"
          :to="{
            name: 'admin-consents-consentId',
            params: { consentId: row.original.id },
          }"
        >
          Rediger
        </UButton>
        <template v-if="canManage">
          <UButton
            v-if="row.original.archivedAt"
            icon="lucide:archive-restore"
            variant="ghost"
            :aria-label="`Gjenopprett ${row.original.title}`"
            @click="handleUnarchive(row.original)"
          >
            Gjenopprett
          </UButton>
          <UButton
            v-else
            icon="lucide:archive"
            variant="ghost"
            color="neutral"
            :aria-label="`Arkiver ${row.original.title}`"
            @click="handleArchive(row.original)"
          >
            Arkiver
          </UButton>
        </template>
      </div>
    </template>
    <template #empty>
      <AdminTableEmpty :title="emptyTitle" :description="emptyDescription" />
    </template>
    <template #loading>
      <AdminTableLoading />
    </template>
  </UTable>
</template>
