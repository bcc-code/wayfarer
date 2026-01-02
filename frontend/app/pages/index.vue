<script setup lang="ts">
const { isAuthReady } = useAuthReady()
const { data, error, fetching, stale } = useProfilePageQuery({
  pause: computed(() => !isAuthReady.value),
})

const isInitialLoading = computed(() => fetching.value && !data.value)

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
    if (fetching.value || stale.value) return
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
    if (pending?.length) {
      if (!showBanner.value) {
        showBanner.value = true
      }
    } else {
      showBanner.value = false
    }
  },
  { immediate: true },
)

// If there is ?achievement=<id> in the url, open the achievement sheet
const { openAchievementSheet, openAchievementId } = useAchievementSheet()
const route = useRoute()
watch(
  () => route.query.achievement,
  (achievementId) => {
    if (achievementId && typeof achievementId === 'string') {
      openAchievementSheet(achievementId)
      console.log(openAchievementId.value)
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

    <div v-if="isInitialLoading" class="space-y-default p-list-outside">
      <ProfileProjectCardSkeleton />
    </div>
    <ErrorState v-else-if="error" :error />
    <div v-else-if="data" class="space-y-default p-list-outside">
      <ProfileProjectCard
        v-if="data.myCurrentProject"
        :project-name="data.myCurrentProject.name"
        :score="data.myCurrentProject.leaderboard.me?.score"
        :rank="data.myCurrentProject.leaderboard.me?.rank"
        :achievements="data.myCurrentProject.achievements"
      >
        <!-- TODO: uncomment when consent is ready from SSF side -->
        <!-- <div v-if="showBanner" class="p-small">
          <ConsentCard
            v-for="consent in pendingRemote"
            :key="consent.id"
            :consent
            class="bg-background-indent!"
          />
        </div> -->
      </ProfileProjectCard>
      <UserFeedback />
    </div>
  </PageLayout>
</template>
