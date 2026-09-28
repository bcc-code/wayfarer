<script setup lang="ts">
import type { TranslationStatusFragment } from '~/api/generated'

defineProps<{
  label: string
  translationStatus?: TranslationStatusFragment[]
  name: string
  hint?: string
  help?: string
}>()
</script>

<template>
  <UFormField :name :hint :help>
    <template #label>
      <span class="flex w-full gap-2 items-center justify-between">
        {{ label }}
        <AdminTranslationIndicator
          v-if="translationStatus?.length"
          :translation-status="translationStatus"
          :field-name="name"
        />
      </span>
    </template>
    <slot />
  </UFormField>
</template>
