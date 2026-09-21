import { computed, reactive, ref, watch, type Ref } from 'vue'
import { refDebounced } from '@vueuse/core'
import type { UsePaginationReturn } from './usePagination'

export interface UseListStateOptions<F extends Record<string, string>> {
  /** The list's pagination, reset automatically whenever the query changes. */
  pagination: UsePaginationReturn
  /** Filter keys with their default (empty) values. Keys become URL params. */
  filters?: F
  /** Debounce on the free-text search, in ms. */
  debounceMs?: number
  /** Page sizes offered in the footer. */
  pageSizes?: number[]
}

export interface UseListStateReturn<F extends Record<string, string>> {
  search: Ref<string>
  /** Debounced search — this is what belongs in query variables. */
  debouncedSearch: Ref<string>
  filters: F
  pageSizes: number[]
  /** Filters (not search) that are currently set, for chips in the toolbar. */
  activeFilters: Ref<Array<{ key: keyof F & string; value: string }>>
  hasActiveQuery: Ref<boolean>
  clearFilter: (key: keyof F & string) => void
  clearAll: () => void
}

/**
 * Search, filters and page size for an admin list, kept in the URL.
 *
 * Two things this owns that every list page would otherwise re-implement, and
 * one of them is a correctness fix rather than convenience:
 *
 * **Pagination resets when the query changes.** Today each page wires its own
 * `watch(debouncedSearch, () => pagination.reset())`. A page that adds a second
 * filter and forgets to extend that watch will ask for page 4 of a result set
 * that now has one page, and get an empty table with no error. Owning the reset
 * here means no page can forget it.
 *
 * **State lives in the URL**, so a refresh keeps context, browser back works,
 * and a filtered list can be sent to someone. Deliberately excluded: the cursor.
 * Relay cursors are opaque and shift as data changes, so `?after=…` produces an
 * ugly link that silently lands somewhere else later. The valuable thing to
 * share is the query, not the page number.
 */
export function useListState<F extends Record<string, string>>(
  options: UseListStateOptions<F>,
): UseListStateReturn<F> {
  const {
    pagination,
    filters: filterDefaults = {} as F,
    debounceMs = 300,
    pageSizes = [15, 25, 50, 100],
  } = options

  const route = useRoute()
  const router = useRouter()

  const readParam = (key: string): string => {
    const value = route.query[key]
    return typeof value === 'string' ? value : ''
  }

  const search = ref(readParam('q'))
  const debouncedSearch = refDebounced(search, debounceMs)

  const filterKeys = Object.keys(filterDefaults) as Array<keyof F & string>
  const filters = reactive({ ...filterDefaults }) as F
  for (const key of filterKeys) {
    const fromUrl = readParam(key)
    if (fromUrl) filters[key] = fromUrl as F[keyof F & string]
  }

  // The page's own default, captured before the URL can override it, so `size`
  // is written only when it actually differs from what the page would do
  // anyway. Comparing against `pageSizes[0]` instead would put `?size=20` on
  // every URL of a list that defaults to 20 but offers 15 first.
  const defaultPageSize = pagination.pageSize.value

  // A page size in the URL is only honoured if it is one we offer — otherwise a
  // hand-edited `?size=100000` would become a query for every row in the table.
  const sizeFromUrl = Number(readParam('size'))
  if (pageSizes.includes(sizeFromUrl)) {
    pagination.setPageSize(sizeFromUrl)
  }

  /** Everything that changes the result set, as URL params. */
  const queryParams = computed(() => {
    const params: Record<string, string> = {}
    if (debouncedSearch.value) params.q = debouncedSearch.value
    for (const key of filterKeys) {
      if (filters[key]) params[key] = filters[key]
    }
    if (pagination.pageSize.value !== defaultPageSize) {
      params.size = String(pagination.pageSize.value)
    }
    return params
  })

  watch(
    queryParams,
    (params, previous) => {
      // `replace`, not `push`: typing into a search box should not fill the
      // history with a step per keystroke-batch.
      void router.replace({ query: { ...params } })

      // Only reset when the *result set* changed. Page size resets itself via
      // `setPageSize`, and resetting again here would double-fire the query.
      const changed = Object.keys({ ...params, ...previous }).some(
        (key) => key !== 'size' && params[key] !== previous?.[key],
      )
      if (changed) pagination.reset()
    },
    { deep: true },
  )

  const activeFilters = computed(() =>
    filterKeys
      .filter((key) => !!filters[key])
      // `F extends Record<string, string>` does not narrow `F[key]` to
      // `string` through the index, so the cast states what the constraint
      // already guarantees.
      .map((key) => ({ key, value: filters[key] as string })),
  )

  const hasActiveQuery = computed(
    () => !!search.value || activeFilters.value.length > 0,
  )

  function clearFilter(key: keyof F & string) {
    filters[key] = '' as F[keyof F & string]
  }

  function clearAll() {
    search.value = ''
    for (const key of filterKeys) clearFilter(key)
  }

  return {
    search,
    debouncedSearch,
    filters,
    pageSizes,
    activeFilters,
    hasActiveQuery,
    clearFilter,
    clearAll,
  }
}
