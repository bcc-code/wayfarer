<script setup lang="ts">
const { locale, locales, setLocale } = useI18n()

const selectedLocale = computed({
  get() {
    return locale.value
  },
  set(value) {
    setLocale(value)
  },
})
</script>

<template>
  <DesignDrawer :title="$t('settings.language')">
    <slot :selected-locale="locales.find((l) => l.code === selectedLocale)" />
    <template #content>
      <DesignButton
        v-for="l in locales"
        :key="l.code"
        variant="secondary"
        size="medium"
        class="w-full"
        @click="selectedLocale = l.code"
      >
        <span class="grow text-start">{{ l.name }}</span>
        <Icon v-if="selectedLocale === l.code" name="lucide:check" />
      </DesignButton>
    </template>
  </DesignDrawer>
</template>
