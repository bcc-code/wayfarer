<script setup lang="ts">
import type { UsePaginationReturn } from '../../composables/usePagination'

const props = withDefaults(
  defineProps<{
    pagination: UsePaginationReturn
    searchPlaceholder?: string
    activeFilters?: Array<{ key: string; value: string; label?: string }>
    /** Noun for the result count, e.g. "brukere". */
    itemLabel?: string
    /** False for filter inputs with no free-text field. */
    searchable?: boolean
  }>(),
  {
    searchPlaceholder: 'Søk…',
    activeFilters: () => [],
    itemLabel: 'rader',
    searchable: true,
  },
)

const search = defineModel<string>('search', { default: '' })

const emit = defineEmits<{
  clearFilter: [key: string]
  clearAll: []
}>()

const range = computed(() => props.pagination.range.value)
const pageSize = computed(() => props.pagination.pageSize.value)

const canGoPrevious = computed(() => !props.pagination.isFirstPage.value)
const canGoNext = computed(() => !props.pagination.isLastPage.value)

// `useListState` validates `?size=` against the same list.
const PAGE_SIZES = [15, 25, 50, 100]
const pageSizeItems = PAGE_SIZES.map((size) => ({
  label: String(size),
  value: size,
}))

const onPageSizeChange = (value: number) => {
  props.pagination.setPageSize(value)
}
</script>

<template>
  <div class="space-y-4">
    <!-- Filters stay inline and in the URL rather than behind an apply step,
         so a filtered list is visible and linkable. -->
    <div class="flex flex-wrap items-center gap-2">
      <UInput
        v-if="searchable"
        v-model="search"
        :placeholder="searchPlaceholder"
        icon="i-lucide-search"
        class="w-full sm:w-72"
        :ui="{ trailing: 'pe-1' }"
      >
        <template v-if="search" #trailing>
          <UButton
            color="neutral"
            variant="link"
            size="sm"
            icon="i-lucide-x"
            aria-label="Tøm søk"
            @click="search = ''"
          />
        </template>
      </UInput>

      <slot name="filters" />

      <div class="ms-auto flex items-center gap-2">
        <slot name="actions" />
      </div>
    </div>

    <!-- Chips, so a short result is never mistaken for the whole truth. -->
    <div v-if="activeFilters.length" class="flex flex-wrap items-center gap-2">
      <UBadge
        v-for="filter in activeFilters"
        :key="filter.key"
        color="neutral"
        variant="subtle"
        class="gap-1"
      >
        {{ filter.label ?? filter.value }}
        <UButton
          color="neutral"
          variant="link"
          size="xs"
          icon="i-lucide-x"
          :aria-label="`Fjern filter: ${filter.label ?? filter.value}`"
          class="p-0"
          @click="emit('clearFilter', filter.key)"
        />
      </UBadge>
      <UButton
        variant="link"
        size="xs"
        color="neutral"
        @click="emit('clearAll')"
      >
        Nullstill
      </UButton>
    </div>

    <slot />

    <!-- No numbered pages: keyset pagination cannot jump to page 5. -->
    <div
      v-if="range"
      class="flex flex-wrap items-center justify-between gap-3 text-sm"
    >
      <div class="text-muted flex items-center gap-2">
        <span>
          Viser {{ formatNumber(range.start) }}–{{ formatNumber(range.end) }} av
          {{ formatNumber(range.total) }} {{ itemLabel }}
        </span>
        <USelect
          :model-value="pageSize"
          :items="pageSizeItems"
          size="xs"
          class="w-20"
          aria-label="Rader per side"
          @update:model-value="onPageSizeChange"
        />
      </div>

      <div class="flex items-center gap-1">
        <UButton
          :disabled="!canGoPrevious"
          variant="soft"
          color="neutral"
          icon="i-lucide-chevron-left"
          label="Forrige"
          size="sm"
          @click="pagination.previousPage()"
        />
        <UButton
          :disabled="!canGoNext"
          variant="soft"
          color="neutral"
          trailing-icon="i-lucide-chevron-right"
          label="Neste"
          size="sm"
          @click="pagination.nextPage()"
        />
      </div>
    </div>
  </div>
</template>
