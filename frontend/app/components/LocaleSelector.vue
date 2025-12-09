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
  <DesignDrawer>
    <slot :selected-locale="locales.find((l) => l.code === selectedLocale)" />
    <template #content="{ close }">
      <PageLayout :title="$t('settings.language')">
        <template #action>
          <DesignIconButton icon="lucide:x" @click="close" />
        </template>
      </PageLayout>
    </template>
  </DesignDrawer>
</template>
