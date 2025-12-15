<script setup lang="ts">
const { isAuthReady } = useAuthReady()
const { data, error, fetching } = useProfilePageQuery({
  pause: computed(() => !isAuthReady.value),
})

// Consent banner
const showBanner = useLocalStorage('showBanner', false, {
  listenToStorageChanges: true,
})
const hasCompletedOnboarding = useLocalStorage('hasCompletedOnboarding', false)

// Watch for pending consents and redirect if user hasn't completed onboarding
watch(
  () => data.value?.me.consentStatus.pendingConsents.length,
  (pending) => {
    if (pending) {
      if (!showBanner.value) {
        showBanner.value = true
      }
      // Redirect to consent page if user hasn't completed onboarding
      if (!hasCompletedOnboarding.value) {
        navigateTo({ name: 'settings-consent' })
      }
    }
  },
  { immediate: true },
)

const remotePendingConsents = computed(() => {
  return data.value?.me.consentStatus.pendingConsents.filter(
    (c) => c.managementType === ConsentManagementType.Remote,
  )
})
</script>

<template>
  <PageLayout :title="data?.me.name">
    <template #action>
      <NuxtLink :to="{ name: 'settings' }">
        <DesignIconButton icon="IconSettings" />
      </NuxtLink>
    </template>

    <LoadingState v-if="fetching" />
    <ErrorState v-else-if="error" :error />
    <div v-else-if="data" class="space-y-list-section-gap p-list-outside">
      <ProfileProjectCard
        v-if="data.myCurrentProject"
        :project-name="data.myCurrentProject.name"
        :score="data.myCurrentProject.leaderboard.me?.score"
        :rank="data.myCurrentProject.leaderboard.me?.rank"
        :achievements="data.myCurrentProject.achievements"
      >
        <div v-if="showBanner" class="p-small">
          <ConsentCard
            v-for="consent in remotePendingConsents"
            :key="consent.id"
            :consent
            class="bg-background-indent!"
          />
        </div>
      </ProfileProjectCard>
    </div>
  </PageLayout>
</template>
