<script setup lang="ts">
const { track } = useAnalytics()
const { locale, locales, setLocale } = useI18n()

const localeName = computed(() => {
  return locales.value.find((l) => l.code === locale.value)?.name
})

const localeComp = computed({
  get() {
    return locale.value
  },
  set(v) {
    track(AnalyticsEvent.LanguageChanged, { from: locale.value, to: v })
    setLocale(v)
  },
})

const colorMode = useColorMode()
const colorModes = ['system', 'dark', 'light']

watch(
  () => colorMode.preference,
  (newMode, oldMode) => {
    if (oldMode) {
      track(AnalyticsEvent.ColorModeChanged, { from: oldMode, to: newMode })
    }
  },
)

const { me } = useAuth()
const { open: consentsOpen } = useConsentsDialog()
</script>

<template>
  <PageLayout :title="$t('pages.settings')" :bottom-padding="false">
    <template #action>
      <NuxtLink :to="{ name: 'index' }">
        <DesignIconButton icon="lucide:x" />
      </NuxtLink>
    </template>

    <DesignPanel class="gap-list-section-inset flex flex-col">
      <UDropdownMenu
        :ui="{
          content:
            'bg-background-raised ring-border-default rounded-list w-(--reka-dropdown-menu-trigger-width)',
        }"
        :content="{ align: 'end', side: 'bottom', sideOffset: -4 }"
        :items="
          locales.map((l) => ({
            label: l.name,
            value: l.code,
            type: 'checkbox',
            checked: l.code == localeComp,
            onSelect: () => (localeComp = l.code),
          }))
        "
        size="xl"
        checked-icon="lucide:check"
      >
        <div class="flex items-center justify-between gap-2.5 px-3 py-2">
          <p class="text-label">{{ $t('settings.language') }}</p>
          <DesignButton size="small" variant="secondary" class="grow-0">
            {{ localeName }}
          </DesignButton>
        </div>
      </UDropdownMenu>
      <hr class="border-border-default mx-3" />
      <UDropdownMenu
        :ui="{
          content:
            'bg-background-raised ring-border-default rounded-list w-(--reka-dropdown-menu-trigger-width)',
        }"
        :content="{ align: 'end', side: 'bottom', sideOffset: -4 }"
        :items="
          colorModes.map((mode) => ({
            label: $t('settings.colorModes.' + mode),
            value: mode,
            type: 'checkbox',
            checked: mode == colorMode.preference,
            onSelect: () => (colorMode.preference = mode),
          }))
        "
      >
        <div class="flex items-center justify-between gap-2.5 px-3 py-2">
          <p class="text-label">{{ $t('settings.colorMode') }}</p>
          <DesignButton size="small" variant="secondary" class="grow-0">
            {{ $t('settings.colorModes.' + colorMode.preference) }}
          </DesignButton>
        </div>
      </UDropdownMenu>
      <hr class="border-border-default mx-3" />
      <button
        class="flex items-center justify-between gap-2.5 px-3 py-2"
        @click="consentsOpen = true"
      >
        <p class="text-label">{{ $t('settings.consents') }}</p>
        <DesignIconButton icon="lucide:chevron-right" />
      </button>
    </DesignPanel>

    <div
      v-if="me"
      class="text-text-hint text-caption p-medium mt-auto text-center"
    >
      <p>{{ me.id }}</p>
      <p>{{ me.church.id }}</p>
    </div>
  </PageLayout>
</template>
