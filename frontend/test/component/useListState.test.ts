// @vitest-environment nuxt
import { describe, it, expect, vi, beforeEach } from 'vitest'
import { mockNuxtImport } from '@nuxt/test-utils/runtime'
import { reactive, ref } from 'vue'
import { useListState } from '../../layers/admin/app/composables/useListState'
import { usePagination } from '../../layers/admin/app/composables/usePagination'

const route = reactive({ query: {} as Record<string, string> })
const replace = vi.fn()

mockNuxtImport('useRoute', () => () => route)

/**
 * Mocking `useRouter` replaces it for Nuxt's own plugins too, and
 * `navigation-repaint.client` calls `router.beforeResolve` from an idle
 * callback — which fake timers then fire. The no-ops keep that plugin happy;
 * only `replace` is under test.
 */
mockNuxtImport('useRouter', () => () => ({
  replace,
  push: vi.fn(),
  beforeResolve: () => () => {},
  beforeEach: () => () => {},
  afterEach: () => () => {},
  onError: () => () => {},
  isReady: () => Promise.resolve(),
  currentRoute: ref(route),
  resolve: (to: unknown) => ({ href: '#', ...(to as object) }),
  options: {},
}))

/** A pagination with page 1 loaded, ready to navigate. */
function loadedPageOne() {
  const pagination = usePagination({ defaultPageSize: 15 })
  pagination.updateConnection({
    edges: Array.from({ length: 15 }, (_, i) => ({
      cursor: `c${i}`,
      node: {},
    })),
    pageInfo: {
      hasNextPage: true,
      hasPreviousPage: false,
      startCursor: 'c0',
      endCursor: 'c14',
    },
    totalCount: 100,
  })
  return pagination
}

/** Page 1 loaded and then stepped forward, so a reset is observable. */
function pageTwo() {
  const pagination = loadedPageOne()
  pagination.nextPage()
  return pagination
}

describe('useListState', () => {
  beforeEach(() => {
    route.query = {}
    replace.mockClear()
  })

  // Real timers with a near-zero debounce, rather than fake timers: VueUse's
  // `refDebounced` schedules through its own watcher, so advancing fake timers
  // does not reliably propagate to the dependent computed within a tick.

  it('reads the search term from the URL', () => {
    route.query = { q: 'sigve' }
    const list = useListState({ pagination: usePagination() })

    expect(list.search.value).toBe('sigve')
  })

  it('reads declared filters from the URL and ignores unknown params', () => {
    route.query = { churchId: 'CH1', bogus: 'x' }
    const list = useListState({
      pagination: usePagination(),
      filters: { churchId: '' },
    })

    expect(list.filters.churchId).toBe('CH1')
    expect('bogus' in list.filters).toBe(false)
  })

  // A hand-edited `?size=100000` would otherwise become a query for every row.
  it('only honours a page size it offers', () => {
    route.query = { size: '100000' }
    const pagination = usePagination({ defaultPageSize: 15 })
    useListState({ pagination })

    expect(pagination.pageSize.value).toBe(15)

    route.query = { size: '50' }
    const other = usePagination({ defaultPageSize: 15 })
    useListState({ pagination: other })

    expect(other.pageSize.value).toBe(50)
  })

  // The reason this composable owns the pagination: a page that adds a filter
  // and forgets its own reset watcher asks for page 4 of a one-page result and
  // gets an empty table with no error.
  it('resets pagination when the search changes', async () => {
    const pagination = pageTwo()
    expect(pagination.range.value?.start).toBe(16)

    const list = useListState({ pagination, debounceMs: 0 })
    list.search.value = 'sigve'

    await vi.waitFor(() => expect(pagination.variables.value.after).toBeFalsy())
  })

  it('resets pagination when a filter changes', async () => {
    const pagination = pageTwo()
    const list = useListState({ pagination, filters: { churchId: '' } })

    list.filters.churchId = 'CH1'
    await nextTick()

    expect(pagination.variables.value.after).toBeFalsy()
  })

  it('writes the query to the URL without stacking history entries', async () => {
    const list = useListState({
      pagination: usePagination(),
      debounceMs: 0,
    })

    list.search.value = 'sigve'

    await vi.waitFor(() =>
      expect(replace.mock.calls.at(-1)?.[0]).toEqual({
        query: { q: 'sigve' },
      }),
    )
  })

  it('drops emptied values from the URL rather than writing blanks', async () => {
    route.query = { q: 'sigve', churchId: 'CH1' }
    const list = useListState({
      pagination: usePagination(),
      filters: { churchId: '' },
      debounceMs: 0,
    })

    list.clearAll()

    await vi.waitFor(() =>
      expect(replace.mock.calls.at(-1)?.[0]).toEqual({ query: {} }),
    )
  })

  // The position is persisted as the pagination variables, not a page number:
  // with keyset pagination the variables *are* the position, so the round trip
  // is exact.
  it('restores a forward position from the URL', () => {
    route.query = { after: 'Y3Vyc29yLTE0', from: '15' }
    const pagination = usePagination({ defaultPageSize: 15 })
    useListState({ pagination })

    expect(pagination.variables.value).toMatchObject({
      first: 15,
      after: 'Y3Vyc29yLTE0',
    })
    expect(pagination.offset.value).toBe(15)
  })

  it('restores a backward position from the URL', () => {
    route.query = { before: 'Y3Vyc29yLTAw', from: '30' }
    const pagination = usePagination({ defaultPageSize: 15 })
    useListState({ pagination })

    expect(pagination.variables.value).toMatchObject({
      last: 15,
      before: 'Y3Vyc29yLTAw',
    })
    expect(pagination.offset.value).toBe(30)
  })

  it('starts at the beginning when the URL carries no position', () => {
    const pagination = usePagination({ defaultPageSize: 15 })
    useListState({ pagination })

    expect(pagination.variables.value.after).toBeFalsy()
    expect(pagination.offset.value).toBe(0)
  })

  it('writes the position to the URL as you page', async () => {
    const pagination = loadedPageOne()
    useListState({ pagination, debounceMs: 0 })

    pagination.nextPage()

    await vi.waitFor(() => {
      const query = replace.mock.calls.at(-1)?.[0]?.query
      expect(query?.after).toBe('c14')
      expect(query?.from).toBe('15')
    })
  })

  // Paging must not trip the reset, or the list could never leave page 1.
  it('does not reset when only the position changes', async () => {
    const pagination = loadedPageOne()
    useListState({ pagination, debounceMs: 0 })

    pagination.nextPage()
    await vi.waitFor(() => expect(replace).toHaveBeenCalled())

    expect(pagination.variables.value.after).toBe('c14')
    expect(pagination.offset.value).toBe(15)
  })

  // A cursor into the previous result set would be meaningless.
  it('drops the position when the query changes', async () => {
    const pagination = loadedPageOne()
    const list = useListState({ pagination, debounceMs: 0 })

    pagination.nextPage()
    await vi.waitFor(() =>
      expect(replace.mock.calls.at(-1)?.[0]?.query?.after).toBe('c14'),
    )

    list.search.value = 'sigve'

    await vi.waitFor(() => {
      const query = replace.mock.calls.at(-1)?.[0]?.query
      expect(query?.q).toBe('sigve')
      expect(query?.after).toBeUndefined()
      expect(query?.from).toBeUndefined()
    })
  })

  it('reports active filters for chips, excluding the free-text search', () => {
    route.query = { q: 'sigve', churchId: 'CH1' }
    const list = useListState({
      pagination: usePagination(),
      filters: { churchId: '' },
    })

    expect(list.activeFilters.value).toEqual([
      { key: 'churchId', value: 'CH1' },
    ])
    expect(list.hasActiveQuery.value).toBe(true)
  })

  it('clears a single filter without touching the others', () => {
    route.query = { churchId: 'CH1', q: 'sigve' }
    const list = useListState({
      pagination: usePagination(),
      filters: { churchId: '' },
    })

    list.clearFilter('churchId')

    expect(list.filters.churchId).toBe('')
    expect(list.search.value).toBe('sigve')
  })
})
