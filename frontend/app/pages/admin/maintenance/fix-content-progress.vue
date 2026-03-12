<script setup lang="ts">
import type { TableColumn } from '@nuxt/ui'

definePageMeta({
  layout: 'admin',
  middleware: ['admin', 'superadmin'],
})

gql(`
  query MaintenanceContentProgressPreview($first: Int) {
    previewMissingContentProgress(first: $first) {
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

gql(`
  mutation FixMissingContentProgress {
    fixMissingContentProgress {
      usersFixed
      progressRecordsCreated
      achievementsAwarded
    }
  }
`)

const { isAuthReady } = useAuthReady()
const {
  data,
  fetching,
  error,
  executeQuery: refetch,
} = useMaintenanceContentProgressPreviewQuery({
  variables: { first: 50 },
  pause: computed(() => !isAuthReady.value),
})

const preview = computed(() => data.value?.previewMissingContentProgress)
const affectedUsers = computed(() => preview.value?.affectedUsers ?? [])

type UserRow = NonNullable<
  MaintenanceContentProgressPreviewQuery['previewMissingContentProgress']
>['affectedUsers'][number]

const columns: TableColumn<UserRow>[] = [
  { accessorKey: 'user.name', header: 'Bruker' },
  { accessorKey: 'eventCount', header: 'Manglende fremgang' },
]

const {
  executeMutation: fix,
  fetching: fixing,
  data: fixResult,
} = useFixMissingContentProgressMutation()

const showConfirmModal = ref(false)
const fixComplete = ref(false)

async function handleFix() {
  showConfirmModal.value = false
  const result = await fix({})
  if (result.data?.fixMissingContentProgress) {
    fixComplete.value = true
    refetch({ requestPolicy: 'network-only' })
  }
}

const toast = useToast()

watch(fixComplete, (complete) => {
  if (complete && fixResult.value?.fixMissingContentProgress) {
    const result = fixResult.value.fixMissingContentProgress
    toast.add({
      title: 'Reparasjon fullfort',
      description: `Opprettet ${result.progressRecordsCreated} fremgangsregistreringer og tildelt ${result.achievementsAwarded} prestasjoner for ${result.usersFixed} brukere.`,
      color: 'success',
    })
  }
})
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
        <h1 class="text-3xl">Fix innholdsfremgang</h1>
      </div>
      <p class="text-muted">
        Dette verktøyet finner brukere som har fullført eksterne
        innholdsoppgaver (f.eks. lest artikler, lyttet til podcasts), men som
        mangler tilhørende fremgangsregistreringer. Det oppretter manglende
        registreringer og tildeler prestasjoner der alle krav er oppfylt.
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
            <span class="text-muted text-sm">Manglende registreringer</span>
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
              Trenger reparasjon
            </UBadge>
          </div>
        </UCard>
      </div>

      <div
        v-if="
          fixComplete && fixResult?.fixMissingContentProgress?.usersFixed === 0
        "
        class="bg-success/10 border-success mb-8 rounded-lg border p-4"
      >
        <div class="flex items-center gap-2">
          <UIcon name="lucide:check-circle" class="text-success h-5 w-5" />
          <span>Ingen data å reparere.</span>
        </div>
      </div>

      <div
        v-if="
          fixComplete &&
          fixResult?.fixMissingContentProgress &&
          fixResult.fixMissingContentProgress.usersFixed > 0
        "
        class="bg-success/10 border-success mb-8 rounded-lg border p-4"
      >
        <div class="flex flex-col gap-2">
          <div class="flex items-center gap-2">
            <UIcon name="lucide:check-circle" class="text-success h-5 w-5" />
            <span class="font-medium">Reparasjon fullført</span>
          </div>
          <ul class="text-muted list-inside list-disc text-sm">
            <li>
              {{ fixResult.fixMissingContentProgress.usersFixed }} brukere
              berørt
            </li>
            <li>
              {{ fixResult.fixMissingContentProgress.progressRecordsCreated }}
              fremgangsregistreringer opprettet
            </li>
            <li>
              {{ fixResult.fixMissingContentProgress.achievementsAwarded }}
              prestasjoner tildelt
            </li>
          </ul>
        </div>
      </div>

      <div class="mb-6 flex items-center justify-between">
        <h2 class="text-xl">Berørte brukere (topp 50)</h2>
        <UButton
          v-if="preview && preview.totalUsers > 0"
          :loading="fixing"
          color="primary"
          @click="showConfirmModal = true"
        >
          Reparer
        </UButton>
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
        Ingen brukere med manglende fremgangsregistreringer funnet.
      </div>
    </template>

    <UModal v-model:open="showConfirmModal">
      <template #content>
        <UCard>
          <template #header>
            <div class="flex items-center gap-2">
              <UIcon name="lucide:alert-triangle" class="text-warning h-5 w-5" />
              <span class="font-medium">Bekreft reparasjon</span>
            </div>
          </template>
          <p class="text-muted">
            Er du sikker på at du vil kjøre reparasjonen? Dette vil opprette
            {{ preview?.totalEvents ?? 0 }} manglende fremgangsregistreringer
            for {{ preview?.totalUsers ?? 0 }} brukere, og tildele prestasjoner
            der alle krav er oppfylt.
          </p>
          <template #footer>
            <div class="flex justify-end gap-2">
              <UButton
                variant="ghost"
                :disabled="fixing"
                @click="showConfirmModal = false"
              >
                Avbryt
              </UButton>
              <UButton color="primary" :loading="fixing" @click="handleFix">
                Kjør reparasjon
              </UButton>
            </div>
          </template>
        </UCard>
      </template>
    </UModal>
  </UContainer>
</template>
