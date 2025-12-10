<script setup lang="ts">
const { isAuthReady } = useAuthReady()
const { data, error, fetching } = useProfilePageQuery({
  pause: computed(() => !isAuthReady.value),
})

// Consent banner
const showBanner = useLocalStorage('showBanner', false, {
  listenToStorageChanges: true,
})
watch(
  () => data.value?.me.consentStatus.pendingConsents.length,
  (pending) => {
    if (pending && !showBanner.value) {
      showBanner.value = true
    }
  },
)

const hasCompletedOnboarding = useLocalStorage('hasCompletedOnboarding', false)
onMounted(() => {
  if (!hasCompletedOnboarding.value && showBanner.value) {
    navigateTo({ name: 'settings-consent' })
  }
})
</script>

<template>
  <PageLayout :title="data?.me.name">
    <template #action>
      <NuxtLink :to="{ name: 'settings' }">
        <DesignIconButton icon="lucide:settings" />
      </NuxtLink>
    </template>

    <LoadingState v-if="fetching" />
    <ErrorState v-else-if="error" :error />
    <div v-else-if="data" class="space-y-list-section-gap">
      <ProfileProjectCard
        v-if="data.myCurrentProject"
        :project-name="data.myCurrentProject.name"
        :score="data.myCurrentProject.leaderboard.me?.score"
        :rank="data.myCurrentProject.leaderboard.me?.rank"
        :achievements="data.myCurrentProject.achievements"
      >
        <div v-if="showBanner" class="p-small">
          <ConsentCard
            v-for="consent in data.me.consentStatus.pendingConsents.filter(
              (c) => c.managementType === ConsentManagementType.Remote,
            )"
            :key="consent.id"
            :consent
            class="bg-background-indent!"
          />
        </div>
      </ProfileProjectCard>
    </div>
  </PageLayout>
</template>
