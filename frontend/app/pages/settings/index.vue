<script setup lang="ts">
const { me } = useAuth()

const notificationsEnabled = ref(false)
</script>

<template>
  <PageLayout :title="$t('pages.settings')" :bottom-padding="false">
    <template #action>
      <NuxtLink :to="{ name: 'index' }">
        <DesignIconButton icon="lucide:x" />
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
        <hr class="border-border-default mx-3" />
        <button class="flex items-center justify-between gap-2.5 px-3 py-2">
          <p class="text-label">{{ $t('settings.notifications') }}</p>
          <DesignSwitch v-model="notificationsEnabled" />
        </button>
      </DesignPanel>
      <DesignPanel class="gap-list-section-inset flex flex-col">
        <button
          class="flex items-center justify-between gap-2.5 px-3 py-2 h-12"
        >
          <p class="text-label">{{ $t('settings.addToHomeScreen') }}</p>
          <Icon name="lucide:chevron-right" class="size-6" />
        </button>
        <hr class="border-border-default mx-3" />
        <NuxtLink
          to="https://bcc.media/privacy"
          class="flex items-center justify-between gap-2.5 px-3 py-2 h-12"
        >
          <p class="text-label">{{ $t('settings.privacyPolicy') }}</p>
          <Icon name="lucide:chevron-right" class="size-6" />
        </NuxtLink>
        <hr class="border-border-default mx-3" />
        <NuxtLink
          :to="{ name: 'settings-consent' }"
          class="flex items-center justify-between gap-2.5 px-3 py-2 h-12"
        >
          <p class="text-label">{{ $t('settings.consents') }}</p>
          <Icon name="lucide:chevron-right" class="size-6" />
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
