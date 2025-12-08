<script
  setup
  lang="ts"
  generic="TConsent extends PendingConsent | AcceptedConsent | RejectedConsent"
>
const props = defineProps<{
  consent: TConsent
  dismissible?: boolean
}>()

const emit = defineEmits<{
  update: []
}>()

const { track } = useAnalytics()
const { executeMutation: acceptConsent } = useAcceptConsentMutation()
const { executeMutation: rejectConsent } = useRejectConsentMutation()

function handleAccept() {
  const consentId = props.consent.id
  acceptConsent({ consentId }).then(({ error }) => {
    if (error) {
      console.error(error)
      return
    }
    track(AnalyticsEvent.ConsentAccepted, { consent_id: consentId })
    emit('update')
  })
}

function handleReject() {
  const consentId = props.consent.id
  rejectConsent({ consentId }).then(({ error }) => {
    if (error) {
      console.error(error)
      return
    }
    track(AnalyticsEvent.ConsentRejected, { consent_id: consentId })
    emit('update')
  })
}

const title = computed(() => {
  if (props.consent.__typename == 'UserConsent') {
    return props.consent.consent.title
  }
  return props.consent.title
})

const body = computed(() => {
  if (props.consent.__typename == 'UserConsent') {
    return props.consent.consent.body.html
  }
  return props.consent.body.html
})

const status = computed(() => {
  if (props.consent.__typename == 'UserConsent') {
    return props.consent.action
  }
  return 'pending'
})

const changing = ref(false)
</script>

<template>
  <DesignPanel class="p-default! space-y-3">
    <h3 class="text-title text-text-default">{{ title }}</h3>
    <div class="text-label text-text-muted" v-html="body" />

    <template v-if="status === 'pending' || changing">
      <button
        class="text-label text-accent-contrast py-2 flex items-center gap-1"
        @click="handleReject"
      >
        {{ $t('consent.iDontConsent') }}
        <Icon name="lucide:arrow-right" />
      </button>
      <DesignButton
        variant="primary"
        size="large"
        class="w-full"
        @click="handleAccept"
      >
        {{ $t('consent.giveConsent') }}
      </DesignButton>
    </template>
    <template v-else-if="status === ConsentAction.Accepted">
      <div class="flex justify-between items-center">
        <span class="text-label text-accent-contrast flex items-center gap-1">
          <Icon name="lucide:check" class="size-6" />
          {{ $t('consent.accepted') }}
        </span>
        <DesignButton
          size="small"
          variant="secondary"
          class="grow-0"
          @click="changing = true"
        >
          {{ $t('consent.change') }}
        </DesignButton>
      </div>
    </template>
    <template v-else-if="status === ConsentAction.Rejected"></template>
  </DesignPanel>
</template>
