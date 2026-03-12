<script setup lang="ts">
import type { TranslationStatusFragment } from '~/api/generated'
import { dbLanguageToLocale, localeToFlagEmoji } from '~/utils/languageMapping'

const props = defineProps<{
  translationStatus: TranslationStatusFragment[]
  fieldName: string
}>()

const translatedLanguages = computed(() => {
  return props.translationStatus
    .filter((ts) => ts.fields.includes(props.fieldName))
    .map((ts) => {
      const locale = dbLanguageToLocale(ts.languageCode)
      return {
        languageCode: ts.languageCode,
        locale,
        flag: localeToFlagEmoji(locale),
      }
    })
})
</script>

<template>
  <UPopover v-if="translatedLanguages.length > 0">
    <button
      type="button"
      class="text-muted hover:text-default inline-flex items-center gap-0.5 text-xs"
    >
      Oversatt til {{ translatedLanguages.length }} språk
    </button>
    <template #content>
      <div class="flex flex-col gap-1 p-2">
        <div
          v-for="lang in translatedLanguages"
          :key="lang.languageCode"
          class="flex items-center gap-2 text-sm"
        >
          <span>{{ lang.flag }}</span>
          <span>{{ lang.languageCode }}</span>
        </div>
      </div>
    </template>
  </UPopover>
</template>
