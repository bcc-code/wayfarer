import { computed, reactive, ref, watch, type Ref } from 'vue'
import { refDebounced } from '@vueuse/core'
import type { UsePaginationReturn } from './usePagination'

export interface UseListStateOptions<F extends Record<string, string>> {
  pagination: UsePaginationReturn
  /** Keys become URL params. */
  filters?: F
  debounceMs?: number
  pageSizes?: number[]
}

export interface UseListStateReturn<F extends Record<string, string>> {
  search: Ref<string>
  /** Use this in query variables, not `search`. */
  debouncedSearch: Ref<string>
  filters: F
  pageSizes: number[]
  /** Set filters, excluding the free-text search. */
  activeFilters: Ref<Array<{ key: keyof F & string; value: string }>>
  hasActiveQuery: Ref<boolean>
  clearFilter: (key: keyof F & string) => void
  clearAll: () => void
}

/**
 * Search, filters and page position for an admin list, kept in the URL.
 *
 * Owns the pagination reset so no page can forget it: asking for page 4 of a
 * result set that now has one page gives an empty table with no error.
 *
 * The position is stored as the pagination variables, not a page number — with
 * keyset pagination the variables *are* the position. `from` rides along only
 * to label it ("Viser 16–30"), since a cursor cannot carry that.
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

  // Captured before the URL can override it, so `size` is only written when it
  // differs from the page's own default.
  const defaultPageSize = pagination.pageSize.value

  // Only a size we offer, or a hand-edited `?size=100000` queries every row.
  const sizeFromUrl = Number(readParam('size'))
  if (pageSizes.includes(sizeFromUrl)) {
    pagination.setPageSize(sizeFromUrl)
  }

  // Before the first query, so a deep link fetches the right page once.
  pagination.restore({
    after: readParam('after') || null,
    before: readParam('before') || null,
    offset: Number(readParam('from')) || 0,
  })

  /** A change here invalidates the position. */
  const queryParams = computed(() => {
    const params: Record<string, string> = {}
    if (debouncedSearch.value) params.q = debouncedSearch.value
    for (const key of filterKeys) {
      if (filters[key]) params[key] = filters[key]
    }
    return params
  })

  /** Changing these must not reset anything, or paging is impossible. */
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
    // `replace`, not `push`: no history entry per keystroke batch.
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
      // Before writing the URL, so it carries the reset position.
      if (changed) pagination.reset()
      writeUrl()
    },
    { deep: true },
  )

  watch(positionParams, writeUrl, { deep: true })

  const activeFilters = computed(() =>
    filterKeys
      .filter((key) => !!filters[key])
      // The constraint guarantees `string`; the index signature loses it.
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
