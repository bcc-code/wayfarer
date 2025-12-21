<script setup lang="ts">
const { me } = useAuth()
const { track } = useAnalytics()

const {
  subscribe,
  unsubscribe,
  isSubscribed,
  isSupported: isPushSupported,
  isLoading: isPushLoading,
  permission: pushPermission,
  error: pushError,
} = usePushNotifications()

async function toggleNotifications(enabled: boolean) {
  if (enabled) {
    await subscribe()
  } else {
    await unsubscribe()
  }
  track(AnalyticsEvent.PushNotificationsToggled, {
    enabled,
  })
}
</script>

<template>
  <PageLayout :title="$t('pages.settings')" :bottom-padding="false">
    <template #action>
      <NuxtLink :to="{ name: 'index' }">
        <DesignIconButton icon="IconClose" />
      </NuxtLink>
    </template>

    <div class="space-y-list-section-gap p-list-outside grow flex flex-col">
      <DesignPanel class="gap-list-section-inset flex flex-col">
        <LocaleSelector v-slot="{ selectedLocale }">
          <div class="flex items-center justify-between gap-2.5 px-3 py-2">
            <p class="text-label">{{ $t('settings.language') }}</p>
            <DesignButton size="small" variant="secondary" class="grow-0">
              {{ selectedLocale?.name }}
            </DesignButton>
          </div>
        </LocaleSelector>
        <!-- <hr class="border-border-default mx-3" />
        <ColorModeSelector v-slot="{ selectedColorMode }">
          <div class="flex items-center justify-between gap-2.5 px-3 py-2">
            <p class="text-label">{{ $t('settings.colorMode') }}</p>
            <DesignButton size="small" variant="secondary" class="grow-0">
              {{ selectedColorMode?.name() }}
            </DesignButton>
          </div>
        </ColorModeSelector> -->
        <hr class="border-border-default mx-3" />
        <label class="flex items-center justify-between gap-2.5 px-3 py-2">
          <p class="text-label">{{ $t('settings.notifications') }}</p>
          <DesignSwitch
            :model-value="isSubscribed"
            :disabled="
              isPushLoading ||
              pushPermission === 'denied' ||
              !$pwa?.isPWAInstalled ||
              !isPushSupported
            "
            :loading="isPushLoading"
            @update:model-value="toggleNotifications"
          />
        </label>
        <p
          v-if="!$pwa?.isPWAInstalled"
          class="text-caption text-text-hint px-3 pb-2"
        >
          {{ $t('settings.notificationsOnlyAvailableWhenInstalled') }}
        </p>
        <p
          v-else-if="pushPermission === 'denied'"
          class="text-text-hint text-caption px-3 pb-2"
        >
          {{ $t('settings.notificationsBlocked') }}
        </p>
        <p
          v-else-if="pushError"
          class="text-accent-negative text-caption px-3 pb-2"
        >
          {{ pushError.message }}
        </p>
      </DesignPanel>
      <DesignPanel class="gap-list-section-inset flex flex-col">
        <NuxtLink
          :to="{ name: 'settings-add-to-home' }"
          class="flex items-center justify-between gap-2.5 px-3 py-2 h-12 disabled:opacity-25 disabled:cursor-not-allowed"
        >
          <p class="text-label">{{ $t('settings.addToHomeScreen') }}</p>
          <IconChevronRight class="size-6" />
        </NuxtLink>
        <hr class="border-border-default mx-3" />
        <NuxtLink
          to="https://bcc.media/personvern"
          class="flex items-center justify-between gap-2.5 px-3 py-2 h-12"
        >
          <p class="text-label">{{ $t('settings.privacyPolicy') }}</p>
          <IconChevronRight class="size-6" />
        </NuxtLink>
        <hr class="border-border-default mx-3" />
        <NuxtLink
          :to="{ name: 'settings-consent' }"
          class="flex items-center justify-between gap-2.5 px-3 py-2 h-12"
        >
          <p class="text-label">{{ $t('settings.consents') }}</p>
          <IconChevronRight class="size-6" />
        </NuxtLink>
      </DesignPanel>
      <UserFeedback />
      <div
        v-if="me"
        class="text-text-hint text-tiny p-medium mt-auto text-center space-y-1"
      >
        <p>{{ me.id }}</p>
        <p>{{ me.church.id }}</p>
        <p>{{ useRuntimeConfig().public.appVersion }}</p>
      </div>
    </div>
  </PageLayout>
</template>
