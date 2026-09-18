<script setup lang="ts">
const { locale, locales, setLocale } = useI18n()
const { track } = useAnalytics()

const selectedLocale = computed({
  get() {
    return locale.value
  },
  set(value) {
    setLocale(value)
  },
})

watch(locale, (newLocale, oldLocale) => {
  if (oldLocale) {
    track(AnalyticsEvent.LanguageChanged, { from: oldLocale, to: newLocale })
  }
})
</script>

<template>
  <DesignDrawer :title="$t('settings.language')">
    <slot :selected-locale="locales.find((l) => l.code === selectedLocale)" />
    <template #content>
      <DesignPanel class="gap-list-section-inset flex flex-col">
        <template v-for="(l, index) in locales" :key="l.code">
          <hr v-if="index > 0" class="border-border-default mx-3" >
          <button
            class="flex items-center justify-between gap-2.5 px-3 py-2 h-12"
            @click="selectedLocale = l.code"
          >
            <p class="text-label">{{ l.name }}</p>
            <Icon
              v-if="selectedLocale === l.code"
              name="IconCheck"
              class="size-6"
            />
          </button>
        </template>
      </DesignPanel>
    </template>
  </DesignDrawer>
</template>
