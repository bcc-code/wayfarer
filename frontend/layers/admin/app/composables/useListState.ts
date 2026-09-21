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
 * and a filtered list can be sent to someone.
 *
 * **The page position is included**, as the pagination variables rather than a
 * page number. With keyset pagination the variables *are* the position, so
 * round-tripping `after`/`before` restores the exact same rows — and a keyset
 * cursor is more stable than a page number would be: it means "the rows after
 * user X" whatever is inserted or deleted before it, and it survives user X
 * being deleted (the SQL compares, it does not look up). What a cursor cannot
 * carry is the *position label*, so `from` rides along for that; a stale
 * `from` mislabels a correct page rather than showing the wrong one.
 *
 * Params are split into two groups because they behave differently: changing
 * what you are querying (`q`, filters) must reset the position, while changing
 * where you are (`after`, `before`, `from`, `size`) must not — resetting on a
 * position change would make paging impossible.
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

  // Restore the position before the first query runs, so a deep link fetches
  // the right page once rather than fetching page 1 and then jumping.
  pagination.restore({
    after: readParam('after') || null,
    before: readParam('before') || null,
    offset: Number(readParam('from')) || 0,
  })

  /** What is being queried. A change here invalidates the position. */
  const queryParams = computed(() => {
    const params: Record<string, string> = {}
    if (debouncedSearch.value) params.q = debouncedSearch.value
    for (const key of filterKeys) {
      if (filters[key]) params[key] = filters[key]
    }
    return params
  })

  /** Where in the result set we are. Changing these must not reset anything. */
  const positionParams = computed(() => {
    const params: Record<string, string> = {}
    const { after, before } = pagination.variables.value

    if (after) params.after = after
    if (before) params.before = before
    if (pagination.offset.value > 0) {
      params.from = String(pagination.offset.value)
    }
    if (pagination.pageSize.value !== defaultPageSize) {
      params.size = String(pagination.pageSize.value)
    }
    return params
  })

  const writeUrl = () => {
    // `replace`, not `push`: typing into a search box should not fill the
    // history with a step per keystroke-batch.
    void router.replace({
      query: { ...queryParams.value, ...positionParams.value },
    })
  }

  watch(
    queryParams,
    (params, previous) => {
      const changed = Object.keys({ ...params, ...previous }).some(
        (key) => params[key] !== previous?.[key],
      )
      // Reset first, so the position params written below are the reset ones
      // rather than a cursor into the previous result set.
      if (changed) pagination.reset()
      writeUrl()
    },
    { deep: true },
  )

  watch(positionParams, writeUrl, { deep: true })

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
