<script setup lang="ts">
import { useAcceptConsentMutation } from '~/api/generated'

type PendingConsent =
  ConsentsDialogQuery['me']['consentStatus']['pendingConsents'][number]
type AcceptedConsent =
  ConsentsDialogQuery['me']['consentStatus']['acceptedConsents'][number]
type RejectedConsent =
  ConsentsDialogQuery['me']['consentStatus']['rejectedConsents'][number]

defineProps<{
  title: string
  consents: PendingConsent[] | AcceptedConsent[] | RejectedConsent[]
}>()

const emit = defineEmits<{
  (e: 'update'): void
}>()

const { executeMutation: acceptConsent } = useAcceptConsentMutation()
const { executeMutation: rejectConsent } = useRejectConsentMutation()

function handleAccept(consentId: string, close: () => void) {
  acceptConsent({ consentId }).then(({ error }) => {
    if (error) {
      console.error(error)
      return
    }
    close()
    emit('update')
  })
}

function handleReject(consentId: string, close: () => void) {
  rejectConsent({ consentId }).then(({ error }) => {
    if (error) {
      console.error(error)
      return
    }
    close()
    emit('update')
  })
}
</script>

<template>
  <div>
    <p class="px-3 py-2 text-title">{{ title }}</p>
    <DesignPanel>
      <template v-for="(consent, index) in consents" :key="consent.id">
        <UModal
          :ui="{
            content:
              'bg-background-raised rounded-list ring-border-default divide-border-default h-[600px]',
            title: 'text-label text-text-default',
            description: 'text-caption text-text-hint',
            body: 'flex flex-col',
            overlay: 'bg-background-indent',
          }"
          :title="
            consent.__typename == 'Consent'
              ? consent.title
              : consent.consent.title
          "
          :description="
            consent.__typename == 'Consent'
              ? consent.managedBy
                ? consent.managedBy
                : ''
              : consent.consent.managedBy
                ? consent.consent.managedBy
                : ''
          "
        >
          <button
            class="flex items-center justify-between gap-2.5 px-3 py-2 text-start w-full"
          >
            <div v-if="consent.__typename == 'Consent'">
              <p class="text-label">{{ consent.title }}</p>
              <p class="text-caption text-text-hint">{{ consent.managedBy }}</p>
            </div>
            <div v-if="consent.__typename == 'UserConsent'">
              <p class="text-label">{{ consent.consent.title }}</p>
              <p class="text-caption text-text-hint">
                {{ consent.consent.managedBy }}
              </p>
            </div>
            <DesignButton size="small" variant="secondary" class="grow-0">
              Read
            </DesignButton>
          </button>
          <template #close>
            <DesignIconButton icon="lucide:x" class="ml-auto" />
          </template>
          <template #body="{ close }">
            <div
              v-if="consent.__typename == 'Consent'"
              class="prose prose-sm text-text-default grow pb-default"
              v-html="consent.body.html"
            />
            <div
              v-if="consent.__typename == 'UserConsent'"
              class="prose prose-sm text-text-default grow pb-default"
              v-html="consent.consent.body.html"
            />
            <div class="flex flex-col gap-small w-full mt-auto">
              <DesignButton
                variant="primary"
                :disabled="
                  consent.__typename == 'UserConsent' &&
                  consent.action == ConsentAction.Accepted
                "
                @click="handleAccept(consent.id, close)"
              >
                {{ $t('consent.acceptButton') }}
              </DesignButton>
              <DesignButton
                variant="tertiary"
                :disabled="
                  consent.__typename == 'UserConsent' &&
                  consent.action == ConsentAction.Rejected
                "
                @click="handleReject(consent.id, close)"
              >
                {{ $t('consent.rejectButton') }}
              </DesignButton>
            </div>
          </template>
        </UModal>
        <hr v-if="index !== consents.length - 1" />
      </template>
    </DesignPanel>
  </div>
</template>
