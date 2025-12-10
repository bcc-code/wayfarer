<script setup lang="ts">
const { data, fetching, error, executeQuery: refetch } = useConsentsPageQuery()
const hasCompletedOnboarding = useLocalStorage('hasCompletedOnboarding', false)

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
</script>

<template>
  <PageLayout :title="$t('pages.consents')">
    <template v-if="hasCompletedOnboarding" #action>
      <NuxtLink :to="{ name: 'settings' }">
        <DesignIconButton icon="lucide:x" />
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

    <LoadingState v-if="fetching" />
    <ErrorState v-else-if="error" :error />
    <div v-else-if="data" class="space-y-list-section-gap px-default">
      <ConsentCard
        v-for="consent in data.me.consentStatus.pendingConsents"
        :key="consent.id"
        :consent
        @update="refetch"
      />
      <ConsentCard
        v-for="consent in data.me.consentStatus.acceptedConsents"
        :key="consent.id"
        :consent
        @update="refetch"
      />
      <ConsentCard
        v-for="consent in data.me.consentStatus.rejectedConsents"
        :key="consent.id"
        :consent
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
        class="w-full"
        @click="finishOnboarding"
      >
        {{ $t('consent.continue') }}
      </DesignButton>
    </div>
  </PageLayout>
</template>
