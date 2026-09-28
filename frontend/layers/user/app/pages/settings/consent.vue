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
      initialSorting.value = [
        ...newData.me.consentStatus.pendingConsents,
        ...newData.me.consentStatus.acceptedConsents,
        ...newData.me.consentStatus.rejectedConsents,
      ].map((c) => (c.__typename === 'Consent' ? c.id : c.consent.id))
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
</script>

<template>
  <PageLayout :title="$t('pages.consents')">
    <template v-if="hasCompletedOnboarding" #action>
      <NuxtLink :to="{ name: 'settings' }">
        <DesignIconButton icon="IconClose" />
      </NuxtLink>
    </template>

    <LocaleSelector v-if="!hasCompletedOnboarding" v-slot="{ selectedLocale }">
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

    <ErrorState v-if="error" :error />
    <div
      v-else-if="data"
      :class="[
        'space-y-list-section-gap',
        {
          'px-default py-list-outside': !hasCompletedOnboarding,
          'p-list-outside': hasCompletedOnboarding,
        },
      ]"
    >
      <ConsentCard
        v-for="consent in sortedConsents"
        :key="consent!.id"
        :consent="consent!"
        @update="refetch"
      />
    </div>
    <div
      v-if="!hasCompletedOnboarding && !hasPendingInternalConsents"
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
