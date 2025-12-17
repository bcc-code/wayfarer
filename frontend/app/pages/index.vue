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

const pendingInternal = computed(() => {
  return data.value?.me.consentStatus.pendingConsents.filter(
    (c) => c.managementType === ConsentManagementType.Local,
  )
})

const pendingRemote = computed(() => {
  return data.value?.me.consentStatus.pendingConsents.filter(
    (c) => c.managementType === ConsentManagementType.Remote,
  )
})

watch(
  pendingInternal,
  (pending) => {
    if (pending?.length) {
      hasCompletedOnboarding.value = false
      navigateTo({ name: 'settings-consent' })
    }
  },
  { immediate: true },
)

watch(
  pendingRemote,
  (pending) => {
    if (pending) {
      if (!showBanner.value) {
        showBanner.value = true
      }
    }
  },
  { immediate: true },
)
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
    <div v-else-if="data" class="space-y-default p-list-outside">
      <ProfileProjectCard
        v-if="data.myCurrentProject"
        :project-name="data.myCurrentProject.name"
        :score="data.myCurrentProject.leaderboard.me?.score"
        :rank="data.myCurrentProject.leaderboard.me?.rank"
        :achievements="data.myCurrentProject.achievements"
      >
        <div v-if="showBanner" class="p-small">
          <ConsentCard
            v-for="consent in pendingRemote"
            :key="consent.id"
            :consent
            class="bg-background-indent!"
          />
        </div>
      </ProfileProjectCard>
    </div>
  </PageLayout>
</template>
