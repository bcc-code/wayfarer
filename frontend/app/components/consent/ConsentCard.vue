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

const initialStatus = computed(() => {
  if (props.consent.__typename == 'UserConsent') {
    return props.consent.action
  }
  return 'pending'
})

const localStatus = ref<'pending' | ConsentAction>(initialStatus.value)

function handleAccept() {
  const consentId =
    props.consent.__typename === 'UserConsent'
      ? props.consent.consent.id
      : props.consent.id
  acceptConsent({ consentId }).then(({ error }) => {
    if (error) {
      console.error(error)
      return
    }
    track(AnalyticsEvent.ConsentAccepted, { consent_id: consentId })
    localStatus.value = ConsentAction.Accepted
    changing.value = false
    emit('update')
  })
}

function handleReject() {
  const consentId =
    props.consent.__typename === 'UserConsent'
      ? props.consent.consent.id
      : props.consent.id
  rejectConsent({ consentId }).then(({ error }) => {
    if (error) {
      console.error(error)
      return
    }
    track(AnalyticsEvent.ConsentRejected, { consent_id: consentId })
    localStatus.value = ConsentAction.Rejected
    changing.value = false
    emit('update')
  })
}

const title = computed(() => {
  if (props.consent.__typename == 'UserConsent') {
    return props.consent.consent.title
  }
  return props.consent.title
})

const shortText = computed(() => {
  if (props.consent.__typename == 'UserConsent') {
    return props.consent.consent.shortText
  }
  return props.consent.shortText
})

const managedBy = computed(() => {
  if (props.consent.__typename == 'UserConsent') {
    return props.consent.consent.managedBy
  }
  return props.consent.managedBy
})

const isRemote = computed(() => {
  if (props.consent.__typename == 'UserConsent') {
    return props.consent.consent.managementType === ConsentManagementType.Remote
  }
  return props.consent.managementType === ConsentManagementType.Remote
})

const url = computed(() => {
  if (props.consent.__typename == 'UserConsent') {
    return props.consent.consent.url
  }
  return props.consent.url
})

const changing = ref(false)
</script>

<template>
  <DesignPanel class="p-default! space-y-3">
    <div>
      <p v-if="managedBy" class="text-caption text-text-muted">
        {{ managedBy }}
      </p>
      <h3 class="text-title text-text-default pr-2">{{ title }}</h3>
    </div>
    <div class="text-label text-text-muted" v-html="shortText" />

    <template v-if="isRemote">
      <template v-if="localStatus === ConsentAction.Accepted">
        <div class="flex justify-between items-center">
          <span class="text-label text-accent-positive flex items-center gap-1">
            <Icon name="IconCheck" class="size-6" />
            {{ $t('consent.accepted') }}
          </span>
          <NuxtLink :to="url" external target="_blank">
            <DesignButton size="small" variant="secondary" class="grow-0">
              {{ $t('consent.change') }}
            </DesignButton>
          </NuxtLink>
        </div>
      </template>
      <template v-else>
        <NuxtLink :to="url" external target="_blank">
          <DesignButton size="large" class="w-full">
            <span>{{ $t('consent.goToConsent') }}</span>
            <Icon name="IconExternalLink" class="size-5" />
          </DesignButton>
        </NuxtLink>
      </template>
    </template>
    <template v-else>
      <template v-if="localStatus === 'pending' || changing">
        <ConsentDetails :consent>
          <button
            class="text-label text-accent-contrast py-2 flex items-center gap-1"
          >
            {{ $t('consent.readButton') }}
            <Icon name="IconArrowRight" class="size-5" />
          </button>
        </ConsentDetails>
        <DesignButton
          variant="primary"
          size="large"
          class="w-full"
          @click="handleAccept"
        >
          {{ $t('consent.acceptButton') }}
        </DesignButton>
        <DesignButton
          variant="tertiary"
          size="small"
          class="w-full"
          @click="handleReject"
        >
          {{ $t('consent.rejectButton') }}
        </DesignButton>
      </template>
      <template v-else-if="localStatus === ConsentAction.Accepted">
        <ConsentDetails :consent>
          <button
            class="text-label text-accent-contrast py-2 flex items-center gap-1"
          >
            {{ $t('consent.readButton') }}
            <Icon name="IconArrowRight" class="size-5" />
          </button>
        </ConsentDetails>
        <div class="flex justify-between items-center">
          <span class="text-label text-accent-positive flex items-center gap-1">
            <Icon name="IconCheck" class="size-6" />
            {{ $t('consent.accepted') }}
          </span>
          <DesignDrawer :title="$t('consent.changeConsent')">
            <DesignButton size="small" variant="secondary" class="grow-0">
              {{ $t('consent.change') }}
            </DesignButton>
            <template #content>
              <p class="text-label text-text-default px-default">
                {{
                  $t('consent.changeDescription', {
                    email: 'support@bcc.media',
                  })
                }}
              </p>
            </template>
          </DesignDrawer>
        </div>
      </template>
      <template v-else-if="localStatus === ConsentAction.Rejected">
        <ConsentDetails :consent>
          <button
            class="text-label text-accent-contrast py-2 flex items-center gap-1"
          >
            {{ $t('consent.readButton') }}
            <Icon name="IconArrowRight" class="size-5" />
          </button>
        </ConsentDetails>
        <div class="flex justify-between items-center">
          <span class="text-label text-accent-negative flex items-center gap-1">
            <Icon name="IconClose" class="size-6" />
            {{ $t('consent.rejected') }}
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
    </template>
  </DesignPanel>
</template>
