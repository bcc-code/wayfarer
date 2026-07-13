<script setup lang="ts">
import { AGE_RANGE_ADULT, AGE_RANGE_YOUNG } from '~/utils/constants'

function getAgeRangeFilter(age: number | null | undefined) {
  if (age == null) return undefined
  if (age >= AGE_RANGE_YOUNG.min && age <= AGE_RANGE_YOUNG.max)
    return AGE_RANGE_YOUNG
  if (age >= AGE_RANGE_ADULT.min && age <= AGE_RANGE_ADULT.max)
    return AGE_RANGE_ADULT

  // If not in any of the ranges, return a dummy range to hide rank
  return { min: 200, max: 201 }
}

const { isAuthReady } = useAuthReady()
const { me } = useAuth()
const { $pwa } = useNuxtApp()
const { track } = useAnalytics()

// Push notification prompt
const {
  isSupported: isPushSupported,
  isSubscribed,
  isInitialized: isPushInitialized,
  permission: pushPermission,
} = usePushNotifications()

const notificationPromptDismissed = useLocalStorage(
  'notificationPromptDismissed',
  false,
)

const notificationPromptOpen = ref(false)

const shouldShowNotificationPrompt = computed(() => {
  // Only show when:
  // 1. Push state is initialized (prevents flash on load)
  // 2. PWA is installed (standalone mode)
  // 3. User hasn't subscribed to notifications
  // 4. User hasn't permanently dismissed the prompt
  // 5. Notification permission isn't denied
  // 6. Push notifications are supported
  return (
    isPushInitialized.value &&
    $pwa?.isPWAInstalled &&
    isPushSupported.value &&
    !isSubscribed.value &&
    !notificationPromptDismissed.value &&
    pushPermission.value !== 'denied'
  )
})

// Auto-open/close the prompt based on conditions
watch(
  shouldShowNotificationPrompt,
  (shouldShow) => {
    if (shouldShow && !notificationPromptOpen.value) {
      notificationPromptOpen.value = true
      track(AnalyticsEvent.NotificationPromptShown)
    } else if (!shouldShow && notificationPromptOpen.value) {
      notificationPromptOpen.value = false
    }
  },
  { immediate: true },
)

function handleNotificationPromptClose() {
  notificationPromptDismissed.value = true
}

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

// Achievement celebration - auto-open drawer with confetti for uncelebrated achievements
const achievements = computed(() => data.value?.myCurrentProject?.achievements)
useAchievementCelebration(achievements)

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
const visibilityState = useDocumentVisibility()
watch([visibilityState, isWindowFocused], ([visibility, focused]) => {
  if (visibility === 'visible' || focused) {
    refresh()
  }
})

// Listen for Firestore realtime updates
useFirestoreRefresh(['ProfilePageDocument'], () => {
  refresh({ requestPolicy: 'network-only' })
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

    <template #title>
      <div class="flex gap-small items-center">
        <SuperTeamBadge
          v-if="data?.myCurrentProject.myTeam?.superTeam"
          :superteam="data.myCurrentProject.myTeam.superTeam"
          class="size-11"
        />
        <h1 class="text-text-default text-heading">{{ data?.me.name }}</h1>
      </div>
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
        :banner="data.myCurrentProject.branding.bannerImage"
        :score="data.myCurrentProject.myPoints"
        :rank="data.myCurrentProject.leaderboard.me?.rank"
        :achievements="data.myCurrentProject.achievements"
      >
        <div v-if="showBanner" class="p-small space-y-4">
          <ConsentCard
            v-for="consent in pendingRemote"
            :key="consent.id"
            :consent
            class="bg-background-indent!"
          />
        </div>
      </ProfileProjectCard>
      <div key="user-feedback" class="pt-small">
        <UserFeedback :project-id="data.myCurrentProject?.id" />
      </div>
    </TransitionGroup>

    <!-- Notification prompt for PWA users -->
    <PwaNotificationPrompt
      v-model:open="notificationPromptOpen"
      @dismiss="handleNotificationPromptClose"
      @subscribe="handleNotificationPromptClose"
    />
  </PageLayout>
</template>
