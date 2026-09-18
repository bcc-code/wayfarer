<script setup lang="ts">
const { locale, setLocale, locales } = useI18n()

const selectedLocale = computed({
  get() {
    return locale.value
  },
  set(value) {
    setLocale(value)
  },
})

const items = computed(() => {
  return locales.value.map((locale) => {
    return {
      label: locale.name,
      value: locale.code,
      icon: localeToFlagEmoji(locale.code),
      onClick() {
        selectedLocale.value = locale.code
      },
    }
  })
})
</script>

<template>
  <UDropdownMenu v-model="selectedLocale" :items>
    <UButton
      square
      color="neutral"
      variant="soft"
    >
      {{ localeToFlagEmoji(selectedLocale) }}
    </UButton>
    <template #item="{ item }">
      <div class="flex items-center gap-2 w-full text-start">
        <span>{{ item.icon }}</span>
        <span class="grow">
          {{ item.label }}
        </span>
        <Icon v-if="item.value === selectedLocale" name="lucide:check" />
      </div>
    </template>
  </UDropdownMenu>
</template>
