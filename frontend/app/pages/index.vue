<script setup lang="ts">
// Age group filter for leaderboard rank
function getAgeRangeFilter(age: number | null | undefined) {
  if (age == null) return undefined
  if (age >= 13 && age <= 19) return { min: 13, max: 19 } // U18
  if (age >= 20 && age <= 37) return { min: 20, max: 37 } // O18
  return undefined // Outside defined age groups
}

const { isAuthReady } = useAuthReady()
const { me } = useAuth()

const {
  data,
  error,
  fetching,
  stale,
  executeQuery: refresh,
} = useProfilePageQuery({
  variables: computed(() => {
    const ageRange = getAgeRangeFilter(me.value?.age)
    if (!ageRange) return {}
    return { ageFilter: { ageRange } }
  }),
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

const isWindowFocused = useWindowFocus()
watch(isWindowFocused, (focused) => {
  if (focused) {
    refresh()
  }
})

// Hidden Treasures link based on locale
const { locale } = useI18n()
const hiddenTreasuresLink = computed(() => {
  const langCode = getHiddenTreasureLocale(locale.value)
  return `https://app.hiddentreasures.org/${langCode}/podcasts/hidden-treasures-podcast`
})
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
    <TransitionGroup
      v-else-if="data"
      tag="div"
      class="space-y-list-section-gap p-list-outside"
      enter-active-class="transition duration-300 ease-out"
      enter-from-class="scale-95 opacity-0"
      enter-to-class="scale-100 opacity-100"
      leave-active-class="transition duration-300 ease-out absolute left-0 right-0"
      leave-from-class="scale-100 opacity-100"
      leave-to-class="scale-95 opacity-0"
      move-class="transition duration-300 ease-out"
    >
      <ProjectInfoBanner
        v-if="data.myCurrentProject.infoMessage"
        key="info-message"
        :project-id="data.myCurrentProject.id"
        :info-message="data.myCurrentProject.infoMessage"
        :info-message-start="data.myCurrentProject.infoMessageStart"
        :info-message-end="data.myCurrentProject.infoMessageEnd"
      />
      <ProfileProjectCard
        v-if="data.myCurrentProject"
        key="current-project"
        :project-name="data.myCurrentProject.name"
        :banner="data.myCurrentProject.branding.banner"
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
        <div v-else class="px-default pb-default">
          <NuxtLink :to="hiddenTreasuresLink" target="_blank">
            <DesignButton size="large" class="w-full">
              {{ $t('goToHiddenTreasures') }}
              <IconArrowRight class="size-5" />
            </DesignButton>
          </NuxtLink>
        </div>
      </ProfileProjectCard>
      <UserFeedback
        key="user-feedback"
        :project-id="data.myCurrentProject?.id"
        class="mt-default"
      />
    </TransitionGroup>
  </PageLayout>
</template>
