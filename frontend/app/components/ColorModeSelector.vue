<script setup lang="ts">
const { t } = useI18n()
const { track } = useAnalytics()
const colorMode = useColorMode()

const colorModes = [
  { code: 'system', name: () => t('settings.colorModes.system') },
  { code: 'dark', name: () => t('settings.colorModes.dark') },
  { code: 'light', name: () => t('settings.colorModes.light') },
]

const selectedColorMode = computed({
  get() {
    return colorMode.preference
  },
  set(value) {
    colorMode.preference = value
  },
})

watch(
  () => colorMode.preference,
  (newMode, oldMode) => {
    if (oldMode) {
      track(AnalyticsEvent.ColorModeChanged, { from: oldMode, to: newMode })
    }
  },
)
</script>

<template>
  <DesignDrawer :title="$t('settings.colorMode')">
    <slot
      :selected-color-mode="
        colorModes.find((m) => m.code === selectedColorMode)
      "
    />
    <template #content>
      <DesignPanel class="gap-list-section-inset flex flex-col">
        <template v-for="(m, index) in colorModes" :key="m.code">
          <hr v-if="index > 0" class="border-border-default mx-3" >
          <button
            class="flex items-center justify-between gap-2.5 px-3 py-2 h-12"
            @click="selectedColorMode = m.code"
          >
            <p class="text-label">{{ m.name() }}</p>
            <IconCheck v-if="selectedColorMode === m.code" class="size-6" />
          </button>
        </template>
      </DesignPanel>
    </template>
  </DesignDrawer>
</template>
