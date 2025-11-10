<script setup>
const { locale, locales, setLocale } = useI18n()

const localeComp = computed({
  get() {
    return locale.value
  },
  set(v) {
    setLocale(v)
  },
})
</script>

<template>
  <USelect
    v-model="localeComp"
    :items="locales"
    label-key="name"
    value-key="code"
  />
</template>
