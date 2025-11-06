<script setup lang="ts">
import type { UsePaginationReturn } from '~/composables/usePagination'

const pagination = defineModel<UsePaginationReturn>('pagination', {
  required: true,
})

// Track current page number (1-indexed for display)
const currentPage = ref(1)

// Calculate total pages
const totalPages = computed(() => {
  if (
    !pagination.value.totalCount.value ||
    pagination.value.pageSize.value <= 0
  ) {
    return 1
  }
  return Math.ceil(
    pagination.value.totalCount.value / pagination.value.pageSize.value,
  )
})

// Watch for navigation to update current page
// When pageInfo changes, we need to determine if we went forward or backward
const previousPageInfo = ref(pagination.value.pageInfo.value)

watch(
  () => pagination.value.pageInfo.value,
  (newPageInfo) => {
    // Reset to page 1 if pageInfo becomes null (reset was called)
    if (!newPageInfo) {
      currentPage.value = 1
      previousPageInfo.value = null
      return
    }

    previousPageInfo.value = newPageInfo
  },
)

// Handle page changes from UPagination
const handlePageChange = (newPage: number) => {
  const currentPageNum = currentPage.value

  if (newPage === currentPageNum) {
    return
  }

  // Going to first page
  if (newPage === 1) {
    pagination.value.firstPage()
    currentPage.value = 1
    return
  }

  // Going forward
  if (newPage > currentPageNum) {
    // Can only go forward one page at a time with relay pagination
    if (newPage === currentPageNum + 1) {
      pagination.value.nextPage()
      currentPage.value = newPage
    } else {
      // User tried to jump multiple pages forward - not supported with relay
      // We'll navigate as far as we can (one page)
      if (pagination.value.pageInfo.value?.hasNextPage) {
        pagination.value.nextPage()
        currentPage.value = currentPageNum + 1
      }
    }
    return
  }

  // Going backward
  if (newPage < currentPageNum) {
    // Can only go backward one page at a time with relay pagination
    if (newPage === currentPageNum - 1) {
      pagination.value.previousPage()
      currentPage.value = newPage
    } else {
      // User tried to jump multiple pages backward - not supported with relay
      // We'll navigate as far as we can (one page)
      if (pagination.value.pageInfo.value?.hasPreviousPage) {
        pagination.value.previousPage()
        currentPage.value = currentPageNum - 1
      }
    }
  }
}

// Disable pagination when there's only one page or no data
const isPaginationDisabled = computed(() => {
  return totalPages.value <= 1
})
</script>

<template>
  <UPagination
    v-if="!isPaginationDisabled"
    v-model:page="currentPage"
    :total="pagination.totalCount.value ?? 0"
    :items-per-page="pagination.pageSize.value"
    show-controls
    @update:page="handlePageChange"
  />
</template>
