<script setup lang="ts">
import type { UsePaginationReturn } from '~/composables/usePagination'

const pagination = defineModel<UsePaginationReturn>('pagination', {
  required: true,
})

// Computed properties for button states
const canGoToPrevious = computed(() => {
  return pagination.value.pageInfo.value?.hasPreviousPage ?? false
})

const canGoToNext = computed(() => {
  return pagination.value.pageInfo.value?.hasNextPage ?? false
})

// Button handlers
const goToPrevious = () => {
  if (canGoToPrevious.value) {
    pagination.value.previousPage()
  }
}

const goToNext = () => {
  if (canGoToNext.value) {
    pagination.value.nextPage()
  }
}

// Show pagination only if there's data
const shouldShowPagination = computed(() => {
  return (pagination.value.totalCount.value ?? 0) > 0
})
</script>

<template>
  <UFieldGroup v-if="shouldShowPagination">
    <UButton
      :disabled="!canGoToPrevious"
      variant="soft"
      icon="i-heroicons-chevron-left"
      aria-label="Previous page"
      square
      @click="goToPrevious"
    />
    <UButton
      :disabled="!canGoToNext"
      variant="soft"
      icon="i-heroicons-chevron-right"
      aria-label="Next page"
      square
      @click="goToNext"
    />
  </UFieldGroup>
</template>
