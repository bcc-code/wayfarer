<script setup lang="ts">
import type { TableColumn } from '@nuxt/ui'

definePageMeta({
  layout: 'admin',
  middleware: ['admin', 'superadmin'],
})

// Hardcoded achievement ID for now
const ACHIEVEMENT_ID = 'AC01KCCNMRWECKQG6W0WCXN4BZ2R'

gql(`
  query MaintenanceScoreJournalPreview($achievementId: ID!, $first: Int) {
    previewMissingScoreJournal(achievementId: $achievementId, first: $first) {
      totalUsers
      totalEvents
      affectedUsers {
        user {
          id
          name
        }
        eventCount
      }
    }
  }
`)

const { isAuthReady } = useAuthReady()
const { data, fetching, error } = useMaintenanceScoreJournalPreviewQuery({
  variables: { achievementId: ACHIEVEMENT_ID, first: 50 },
  pause: computed(() => !isAuthReady.value),
})

const preview = computed(() => data.value?.previewMissingScoreJournal)
const affectedUsers = computed(() => preview.value?.affectedUsers ?? [])

type UserRow = NonNullable<
  MaintenanceScoreJournalPreviewQuery['previewMissingScoreJournal']
>['affectedUsers'][number]

const columns: TableColumn<UserRow>[] = [
  { accessorKey: 'user.name', header: 'Bruker' },
  { accessorKey: 'eventCount', header: 'Manglende oppforinger' },
]
</script>

<template>
  <UContainer class="py-12">
    <div class="mb-8">
      <div class="mb-4 flex items-center gap-2">
        <UButton
          variant="ghost"
          icon="lucide:arrow-left"
          to="/admin/maintenance"
        />
        <h1 class="text-3xl">Check points journal</h1>
      </div>
      <p class="text-muted">
        Dette verktøyet viser brukere som har fullført eksterne
        innholdsoppgaver, men som mangler tilhørende poengjournal-oppforinger.
      </p>
      <p class="text-muted mt-2 text-sm">
        Prestasjon-ID: <code class="bg-muted rounded px-1">{{ ACHIEVEMENT_ID }}</code>
      </p>
    </div>

    <ErrorState v-if="error" :error />

    <template v-else>
      <div class="mb-8 grid gap-4 md:grid-cols-3">
        <UCard>
          <div class="flex flex-col gap-1">
            <span class="text-muted text-sm">Berørte brukere</span>
            <span v-if="fetching" class="text-2xl font-bold">...</span>
            <span v-else class="text-2xl font-bold">
              {{ preview?.totalUsers ?? 0 }}
            </span>
          </div>
        </UCard>
        <UCard>
          <div class="flex flex-col gap-1">
            <span class="text-muted text-sm">Manglende oppforinger</span>
            <span v-if="fetching" class="text-2xl font-bold">...</span>
            <span v-else class="text-2xl font-bold">
              {{ preview?.totalEvents ?? 0 }}
            </span>
          </div>
        </UCard>
        <UCard>
          <div class="flex flex-col gap-1">
            <span class="text-muted text-sm">Status</span>
            <UBadge
              v-if="preview?.totalUsers === 0"
              variant="soft"
              color="success"
            >
              Alt i orden
            </UBadge>
            <UBadge v-else-if="fetching" variant="soft" color="neutral">
              Laster...
            </UBadge>
            <UBadge v-else variant="soft" color="warning">
              Mangler oppforinger
            </UBadge>
          </div>
        </UCard>
      </div>

      <div class="mb-6">
        <h2 class="text-xl">Berørte brukere (topp 50)</h2>
      </div>

      <UTable :data="affectedUsers" :loading="fetching" :columns>
        <template #user.name-cell="{ row }">
          <NuxtLink
            :to="{
              name: 'admin-users-userId',
              params: { userId: row.original.user.id },
            }"
            class="hover:text-primary font-medium hover:underline"
          >
            {{ row.original.user.name }}
          </NuxtLink>
        </template>
        <template #eventCount-cell="{ row }">
          <UBadge variant="soft" color="warning">
            {{ row.original.eventCount }}
          </UBadge>
        </template>
      </UTable>

      <div
        v-if="affectedUsers.length === 0 && !fetching"
        class="text-muted py-12 text-center"
      >
        Ingen brukere med manglende poengjournal-oppforinger funnet.
      </div>
    </template>
  </UContainer>
</template>
