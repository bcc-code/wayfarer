<script setup lang="ts">
import type { UsePaginationReturn } from '../../composables/usePagination'

const props = withDefaults(
  defineProps<{
    pagination: UsePaginationReturn
    searchPlaceholder?: string
    /** Filter chips shown in the toolbar; pair with `useListState`. */
    activeFilters?: Array<{ key: string; value: string; label?: string }>
    /** Noun for the result count, e.g. "brukere". */
    itemLabel?: string
    /** Hide the search box for lists that are not searchable yet. */
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

/**
 * Only the sizes the page offers. `useListState` validates the URL against the
 * same list, so a hand-edited `?size=` cannot widen the query either.
 */
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
    <!--
      Toolbar: search and filters on the left, page actions on the right. The
      shape is the one every data-table reference agrees on; the filter *panel*
      pattern from consumer apps is deliberately not copied, because an admin
      list wants its state visible and linkable rather than hidden behind an
      apply step.
    -->
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

    <!--
      Active filters stay visible as chips rather than living only inside their
      controls — with three or more facets it is otherwise easy to forget a
      list is filtered and read a short result as the whole truth.
    -->
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

    <!--
      Footer: position on the left, controls on the right — the layout all four
      reference dashboards use. Numbered pages are deliberately absent: the API
      is keyset-paginated, so there is no way to jump to page 5 without walking
      there. See notes/admin-ux-improvements.md.
    -->
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
