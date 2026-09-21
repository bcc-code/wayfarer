<script setup lang="ts">
import { ConsentAction, ConsentManagementType } from '~/api/generated'

gql(`
  mutation AdminSetUserConsent($userId: ID!, $consentId: ID!, $action: ConsentAction!) {
    adminSetUserConsent(userId: $userId, consentId: $consentId, action: $action) {
      id
      action
    }
  }
`)

/**
 * A user's consent history as one sorted list, and the withdrawal it allows.
 *
 * The API returns **two different shapes** for the same concept:
 * `pendingConsents` are bare `Consent`s, while `acceptedConsents` and
 * `rejectedConsents` are `UserConsent`s wrapping one, carrying an `actionDate`
 * and the `managementType` that decides whether the consent can be withdrawn
 * here at all. `ConsentRow` flattens the three so the template renders one
 * loop rather than three near-identical blocks.
 */
interface ConsentSummary {
  id: string
  key: string
  title: string
  version: number
}

interface UserConsentEntry {
  id: string
  actionDate: string
  consent: ConsentSummary & { managementType?: ConsentManagementType }
}

const props = defineProps<{
  userId: string
  consentStatus: {
    pendingConsents: ConsentSummary[]
    acceptedConsents: UserConsentEntry[]
    rejectedConsents: UserConsentEntry[]
  }
}>()

/** The page owns the query, so it refetches when a consent changes. */
const emit = defineEmits<{ changed: [] }>()

type ConsentRowStatus = 'pending' | 'accepted' | 'rejected'

interface ConsentRow {
  /** Unique across the three source lists, which can share ids. */
  rowKey: string
  status: ConsentRowStatus
  title: string
  version: number
  consentKey: string
  /** Only accepted/rejected rows have a decision date. */
  actionDate?: string
  /** Only locally-managed consents can be withdrawn from here. */
  removableConsentId?: string
}

const STATUS_ORDER: Record<ConsentRowStatus, number> = {
  pending: 0,
  accepted: 1,
  rejected: 2,
}

const STATUS_LABELS: Record<ConsentRowStatus, string> = {
  pending: 'Ventende',
  accepted: 'Akseptert',
  rejected: 'Avvist',
}

const STATUS_COLORS: Record<ConsentRowStatus, 'warning' | 'success' | 'error'> =
  {
    pending: 'warning',
    accepted: 'success',
    rejected: 'error',
  }

/**
 * Grouped by sorting rather than by heading: every row already carries a status
 * badge, so sub-headings named "Ventende / Akseptert / Avvist" said the same
 * word twice. Pending sorts first because it is the only status that wants
 * someone to act.
 */
const rows = computed<ConsentRow[]>(() => {
  const status = props.consentStatus

  const all: ConsentRow[] = [
    ...status.pendingConsents.map((consent) => ({
      rowKey: `pending-${consent.id}`,
      status: 'pending' as const,
      title: consent.title,
      version: consent.version,
      consentKey: consent.key,
    })),
    ...status.acceptedConsents.map((item) => ({
      rowKey: `accepted-${item.id}`,
      status: 'accepted' as const,
      title: item.consent.title,
      version: item.consent.version,
      consentKey: item.consent.key,
      actionDate: item.actionDate,
      removableConsentId:
        item.consent.managementType === ConsentManagementType.Local
          ? item.consent.id
          : undefined,
    })),
    ...status.rejectedConsents.map((item) => ({
      rowKey: `rejected-${item.id}`,
      status: 'rejected' as const,
      title: item.consent.title,
      version: item.consent.version,
      consentKey: item.consent.key,
      actionDate: item.actionDate,
    })),
  ]

  return all.sort(
    (a, b) =>
      STATUS_ORDER[a.status] - STATUS_ORDER[b.status] ||
      a.title.localeCompare(b.title, 'nb'),
  )
})

const { executeMutation: setUserConsent } = useAdminSetUserConsentMutation()
const toast = useToast()

const showRemoveModal = ref(false)
const consentToRemove = ref<{ id: string; title: string } | null>(null)

function openRemoveModal(consentId: string, consentTitle: string) {
  consentToRemove.value = { id: consentId, title: consentTitle }
  showRemoveModal.value = true
}

function closeRemoveModal() {
  showRemoveModal.value = false
  consentToRemove.value = null
}

async function handleRemove() {
  if (!consentToRemove.value) return

  const { id, title } = consentToRemove.value
  const result = await setUserConsent({
    userId: props.userId,
    consentId: id,
    action: ConsentAction.Rejected,
  })

  if (result.error) {
    toast.add({
      title: 'Kunne ikke fjerne samtykke',
      description: result.error.message,
      color: 'error',
    })
    return
  }

  toast.add({
    title: 'Samtykke fjernet',
    description: `Fjernet samtykke for "${title}"`,
    color: 'success',
  })

  closeRemoveModal()
  emit('changed')
}
</script>

<template>
  <AdminSection title="Samtykker" :count="rows.length">
    <div v-if="rows.length" class="space-y-1">
      <div
        v-for="row in rows"
        :key="row.rowKey"
        class="flex flex-wrap items-center justify-between gap-3 py-3"
      >
        <div class="flex min-w-0 items-center gap-3">
          <UBadge variant="soft" :color="STATUS_COLORS[row.status]">
            {{ STATUS_LABELS[row.status] }}
          </UBadge>
          <div class="min-w-0">
            <span class="font-medium">{{ row.title }}</span>
            <span class="text-dimmed ml-2 text-xs">v{{ row.version }}</span>
          </div>
        </div>

        <div class="ms-auto flex items-center gap-3">
          <UButton
            v-if="row.removableConsentId"
            color="neutral"
            variant="soft"
            size="sm"
            @click="openRemoveModal(row.removableConsentId, row.title)"
          >
            Fjern samtykke
          </UButton>
          <div class="text-right">
            <code class="text-dimmed text-xs">{{ row.consentKey }}</code>
            <div v-if="row.actionDate" class="text-dimmed text-xs">
              {{ formatDateTime(row.actionDate) }}
            </div>
          </div>
        </div>
      </div>
    </div>
    <p v-else class="text-dimmed text-sm">Ingen samtykkeaktivitet</p>

    <UModal v-model:open="showRemoveModal">
      <template #header>
        <h3 class="text-lg font-semibold">Fjern samtykke</h3>
      </template>

      <template #body>
        <p>
          Er du sikker på at du vil fjerne samtykke for
          <strong>{{ consentToRemove?.title }}</strong
          >?
        </p>
        <p class="text-dimmed mt-2 text-sm">
          Samtykket vil bli markert som avvist.
        </p>
      </template>

      <template #footer>
        <div class="flex w-full justify-end gap-3">
          <UButton variant="ghost" color="neutral" @click="closeRemoveModal">
            Avbryt
          </UButton>
          <UButton color="error" @click="handleRemove">Fjern samtykke</UButton>
        </div>
      </template>
    </UModal>
  </AdminSection>
</template>
