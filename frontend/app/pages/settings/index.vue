<script setup lang="ts">
const { me } = useAuth()
const { $pwa } = useNuxtApp()

const {
  subscribe,
  unsubscribe,
  isSubscribed,
  isSupported: isPushSupported,
  isLoading: isPushLoading,
} = usePushNotifications()

async function toggleNotifications(enabled: boolean) {
  if (enabled) {
    await subscribe()
  } else {
    await unsubscribe()
  }
}
</script>

<template>
  <PageLayout :title="$t('pages.settings')" :bottom-padding="false">
    <template #action>
      <NuxtLink :to="{ name: 'index' }">
        <DesignIconButton icon="IconClose" />
      </NuxtLink>
    </template>

    <div class="space-y-list-section-gap p-list-outside">
      <DesignPanel class="gap-list-section-inset flex flex-col">
        <LocaleSelector v-slot="{ selectedLocale }">
          <div class="flex items-center justify-between gap-2.5 px-3 py-2">
            <p class="text-label">{{ $t('settings.language') }}</p>
            <DesignButton size="small" variant="secondary" class="grow-0">
              {{ selectedLocale?.name }}
            </DesignButton>
          </div>
        </LocaleSelector>
        <hr class="border-border-default mx-3" />
        <ColorModeSelector v-slot="{ selectedColorMode }">
          <div class="flex items-center justify-between gap-2.5 px-3 py-2">
            <p class="text-label">{{ $t('settings.colorMode') }}</p>
            <DesignButton size="small" variant="secondary" class="grow-0">
              {{ selectedColorMode?.name() }}
            </DesignButton>
          </div>
        </ColorModeSelector>
        <template v-if="isPushSupported">
          <hr class="border-border-default mx-3" />
          <label class="flex items-center justify-between gap-2.5 px-3 py-2">
            <p class="text-label">{{ $t('settings.notifications') }}</p>
            <DesignSwitch
              :model-value="isSubscribed"
              :disabled="isPushLoading"
              @update:model-value="toggleNotifications"
            />
          </label>
        </template>
      </DesignPanel>
      <DesignPanel class="gap-list-section-inset flex flex-col">
        <button
          class="flex items-center justify-between gap-2.5 px-3 py-2 h-12 disabled:opacity-25 disabled:cursor-not-allowed"
          :disabled="!($pwa?.showInstallPrompt && !$pwa?.isPWAInstalled)"
          @click="$pwa?.install()"
        >
          <p class="text-label">{{ $t('settings.addToHomeScreen') }}</p>
          <Icon name="IconChevronRight" class="size-6" />
        </button>
        <hr class="border-border-default mx-3" />
        <NuxtLink
          to="https://bcc.media/privacy"
          class="flex items-center justify-between gap-2.5 px-3 py-2 h-12"
        >
          <p class="text-label">{{ $t('settings.privacyPolicy') }}</p>
          <Icon name="IconChevronRight" class="size-6" />
        </NuxtLink>
        <hr class="border-border-default mx-3" />
        <NuxtLink
          :to="{ name: 'settings-consent' }"
          class="flex items-center justify-between gap-2.5 px-3 py-2 h-12"
        >
          <p class="text-label">{{ $t('settings.consents') }}</p>
          <Icon name="IconChevronRight" class="size-6" />
        </NuxtLink>
      </DesignPanel>
      <div
        v-if="me"
        class="text-text-hint text-caption p-medium mt-auto text-center"
      >
        <p>{{ me.id }}</p>
        <p>{{ me.church.id }}</p>
      </div>
    </div>
  </PageLayout>
</template>
