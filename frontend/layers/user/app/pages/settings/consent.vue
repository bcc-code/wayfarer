<script setup lang="ts">
const { data, error, executeQuery: refetch } = useConsentsPageQuery()
const hasCompletedOnboarding = useLocalStorage('hasCompletedOnboarding', false)

// Refresh consents when locale changes
const { locale } = useI18n()
watch(locale, (newLocale) => {
  if (newLocale != locale.value) {
    nextTick(() => refetch())
  }
})

function finishOnboarding() {
  hasCompletedOnboarding.value = true
  navigateTo({ name: 'index' })
}

// A required consent counts until it is ACCEPTED, so a rejected one still
// blocks — unlike `pending`, which only means "no action yet". Blocking wins
// over `hasCompletedOnboarding`, which is per-device localStorage and clearable.
const blockingIds = computed(() => {
  const status = data.value?.me.consentStatus
  if (!status) return new Set<string>()

  const isLocal = (c: { managementType: ConsentManagementType }) =>
    c.managementType === ConsentManagementType.Local

  return new Set([
    ...status.pendingConsents
      .filter((c) => c.isRequired && isLocal(c))
      .map((c) => c.id),
    ...status.rejectedConsents
      .filter((entry) => entry.consent.isRequired && isLocal(entry.consent))
      .map((entry) => entry.consent.id),
  ])
})

const hasBlockingConsents = computed(() => blockingIds.value.size > 0)

const hasPendingInternalConsents = computed(() => {
  const pending = data.value?.me.consentStatus.pendingConsents.filter(
    (c) => c.managementType === ConsentManagementType.Local,
  )
  if (!pending) return false
  return pending.length > 0
})

// We want to preserve the order of the consents, even when data is refetched
const initialSorting = ref<string[]>([])
watch(
  data,
  (newData) => {
    if (newData) {
      // Blocking ones first: a rejected required consent would otherwise sort
      // last, leaving the user stuck on a screen whose reason is off-screen.
      // `sort` is stable, so the buckets keep their order within each half.
      initialSorting.value = [
        ...newData.me.consentStatus.pendingConsents,
        ...newData.me.consentStatus.acceptedConsents,
        ...newData.me.consentStatus.rejectedConsents,
      ]
        .map((c) => (c.__typename === 'Consent' ? c.id : c.consent.id))
        .sort(
          (a, b) =>
            Number(blockingIds.value.has(b)) - Number(blockingIds.value.has(a)),
        )
    }
  },
  {
    once: true,
  },
)

const allConsents = computed(() => {
  if (!data.value) return []
  return [
    ...data.value.me.consentStatus.pendingConsents,
    ...data.value.me.consentStatus.acceptedConsents,
    ...data.value.me.consentStatus.rejectedConsents,
  ]
})

const sortedConsents = computed(() => {
  if (!data.value) return []
  return initialSorting.value
    .map((id) =>
      allConsents.value.find((c) =>
        c.__typename === 'Consent' ? c.id === id : c.consent.id === id,
      ),
    )
    .filter(Boolean)
})

/**
 * While a required consent is unanswered the page stops being a list and
 * becomes a gate: only the blocking consents are shown. A list invites
 * browsing, and the one card that has to be dealt with should not compete with
 * optional ones the user cannot reach the app to use anyway. The rest come
 * back the moment the requirement is met.
 */
const visibleConsents = computed(() => {
  if (!hasBlockingConsents.value) return sortedConsents.value

  return sortedConsents.value.filter((c) =>
    blockingIds.value.has(c!.__typename === 'Consent' ? c!.id : c!.consent.id),
  )
})
</script>

<template>
  <PageLayout
    :title="
      hasBlockingConsents ? $t('consent.requiredTitle') : $t('pages.consents')
    "
    :bottom-padding="!hasBlockingConsents"
  >
    <template v-if="hasCompletedOnboarding && !hasBlockingConsents" #action>
      <NuxtLink :to="{ name: 'settings' }">
        <DesignIconButton icon="IconClose" />
      </NuxtLink>
    </template>

    <LocaleSelector
      v-if="!hasCompletedOnboarding || hasBlockingConsents"
      v-slot="{ selectedLocale }"
    >
      <DesignPanel
        class="mx-default border border-border-default bg-background-default! mb-default"
      >
        <div class="flex items-center justify-between gap-2.5 px-3 py-2">
          <p class="text-label">{{ $t('settings.language') }}</p>
          <DesignButton size="small" variant="secondary" class="grow-0">
            {{ selectedLocale?.name }}
          </DesignButton>
        </div>
      </DesignPanel>
    </LocaleSelector>

    <div :class="hasBlockingConsents ? 'flex grow flex-col' : ''">
      <ErrorState v-if="error && !data" :error />
      <div
        v-else-if="data"
        :class="[
          'space-y-list-section-gap',
          // Gated, the card is the screen: an even `default` gutter on every
          // side. The top one comes from the language picker's own mb-default,
          // so adding pt here would double it.
          hasBlockingConsents
            ? 'flex grow flex-col px-default pb-default'
            : {
                'px-default py-list-outside': !hasCompletedOnboarding,
                'p-list-outside': hasCompletedOnboarding,
              },
        ]"
      >
        <ConsentCard
          v-for="consent in visibleConsents"
          :key="consent!.id"
          :consent="consent!"
          :expanded="hasBlockingConsents"
          @update="refetch"
        />
      </div>
    </div>
    <div
      v-if="
        !hasCompletedOnboarding &&
        !hasPendingInternalConsents &&
        !hasBlockingConsents
      "
      class="fixed bottom-default left-default right-default"
    >
      <DesignButton
        variant="secondary"
        size="large"
        class="w-full max-w-md mx-auto"
        @click="finishOnboarding"
      >
        {{ $t('consent.continue') }}
      </DesignButton>
    </div>
  </PageLayout>
</template>
