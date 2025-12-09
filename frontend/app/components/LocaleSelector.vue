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
    <template #content="{ close }"></template>
  </DesignDrawer>
</template>
