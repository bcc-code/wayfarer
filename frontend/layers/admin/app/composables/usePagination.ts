import { ref, computed, type Ref } from 'vue'
import {
  buildNextPageVariables,
  buildPreviousPageVariables,
  buildFirstPageVariables,
  buildLastPageVariables,
  isFirstPage as isFirstPagePure,
  isLastPage as isLastPagePure,
  validatePageSize,
  nextOffset,
  previousOffset,
  pageRange,
  type PageRange,
  type CursorPageInfo,
  type CursorPaginationVariables,
} from '../utils/pagination'

export type PaginationPageInfo = CursorPageInfo
export type PaginationVariables = CursorPaginationVariables

export interface Edge<T> {
  cursor: string
  node: T
}

export interface Connection<T> {
  edges: Edge<T>[]
  pageInfo: PaginationPageInfo
  totalCount?: number
}

export interface UsePaginationOptions {
  /** Default page size for forward pagination */
  defaultPageSize?: number
  /** Initial cursor to start from */
  initialCursor?: string | null
  /** Pagination direction: 'forward' (oldest first) or 'backward' (newest first) */
  direction?: 'forward' | 'backward'
}

export interface UsePaginationReturn {
  /** Current pagination variables for GraphQL query */
  variables: Ref<PaginationVariables>
  /** Current page info from the last query result */
  pageInfo: Ref<PaginationPageInfo | null>
  /** Total count of items across all pages */
  totalCount: Ref<number | null>
  /** Current page size being used */
  pageSize: Ref<number>
  /**
   * Which rows are on screen, 1-based, or null when there is nothing to show.
   * Tracked by counting steps — keyset cursors cannot report a position.
   */
  range: Ref<PageRange | null>
  /** Whether we're currently on the first page */
  isFirstPage: Ref<boolean>
  /** Whether we're currently on the last page */
  isLastPage: Ref<boolean>
  /** Navigate to the next page */
  nextPage: () => void
  /** Navigate to the previous page */
  previousPage: () => void
  /** Navigate to the first page */
  firstPage: () => void
  /** Update pagination state with new connection data */
  updateConnection: <T>(connection: Connection<T> | null | undefined) => void
  /** Reset pagination to initial state */
  reset: () => void
  /** Set a new page size */
  setPageSize: (size: number) => void
  /**
   * Jump straight to a known position, for restoring from the URL.
   *
   * Takes the query variables rather than a page number on purpose: with
   * keyset pagination the variables *are* the position, so round-tripping them
   * is exact. There is no page number to restore — see `restore` in
   * useListState.
   */
  restore: (state: {
    after?: string | null
    before?: string | null
    offset?: number
  }) => void
  /** Rows stepped past, for persisting the position label. */
  offset: Ref<number>
}

/**
 * Composable for managing Relay-style cursor-based pagination
 *
 * @example
 * ```ts
 * const { variables, pageInfo, totalCount, nextPage, previousPage, updateConnection } = usePagination({
 *   defaultPageSize: 20
 * })
 *
 * // Use variables in your GraphQL query
 * const { data } = await useAsyncQuery(query, variables)
 *
 * // Update pagination state with the result
 * updateConnection(data.value?.items)
 *
 * // Access total count
 * console.log(`Total items: ${totalCount.value}`)
 *
 * // Navigate
 * nextPage()
 * previousPage()
 * ```
 */
export function usePagination(
  options: UsePaginationOptions = {},
): UsePaginationReturn {
  const {
    defaultPageSize = 20,
    initialCursor = null,
    direction = 'forward',
  } = options
  const isBackward = direction === 'backward'

  // Build initial variables based on direction
  const buildInitialVariables = (size: number) =>
    isBackward
      ? buildLastPageVariables(size, initialCursor)
      : buildFirstPageVariables(size, initialCursor)

  // Reactive state
  const pageSize = ref(defaultPageSize)
  const pageInfo = ref<PaginationPageInfo | null>(null)
  const totalCount = ref<number | null>(null)
  /** Rows stepped past so far. See `pageRange` in utils/pagination. */
  const offset = ref(0)
  /** Rows the last query actually returned, so a short final page reads right. */
  const rowsOnPage = ref(0)
  const variables = ref<PaginationVariables>(
    buildInitialVariables(defaultPageSize),
  )

  // Computed properties
  // For backward pagination, "first page" is the last page of data (newest items)
  const isFirstPage = computed(() =>
    isBackward
      ? isLastPagePure(pageInfo.value)
      : isFirstPagePure(pageInfo.value),
  )
  const isLastPage = computed(() =>
    isBackward
      ? isFirstPagePure(pageInfo.value)
      : isLastPagePure(pageInfo.value),
  )

  // Methods
  // For backward pagination, "next" means older items, "previous" means newer items
  function nextPage() {
    if (!pageInfo.value) return

    const nextVars = isBackward
      ? buildPreviousPageVariables(pageInfo.value, pageSize.value)
      : buildNextPageVariables(pageInfo.value, pageSize.value)
    if (nextVars) {
      variables.value = nextVars
      offset.value = nextOffset(offset.value, pageSize.value)
    }
  }

  function previousPage() {
    if (!pageInfo.value) return

    const prevVars = isBackward
      ? buildNextPageVariables(pageInfo.value, pageSize.value)
      : buildPreviousPageVariables(pageInfo.value, pageSize.value)
    if (prevVars) {
      variables.value = prevVars
      offset.value = previousOffset(offset.value, pageSize.value)
    }
  }

  function firstPage() {
    variables.value = buildInitialVariables(pageSize.value)
    pageInfo.value = null
    totalCount.value = null
    offset.value = 0
    rowsOnPage.value = 0
  }

  function updateConnection<T>(connection: Connection<T> | null | undefined) {
    if (!connection) {
      pageInfo.value = null
      totalCount.value = null
      rowsOnPage.value = 0
      return
    }

    pageInfo.value = connection.pageInfo
    totalCount.value = connection.totalCount ?? null
    rowsOnPage.value = connection.edges.length
  }

  function reset() {
    pageInfo.value = null
    totalCount.value = null
    offset.value = 0
    rowsOnPage.value = 0
    variables.value = buildInitialVariables(pageSize.value)
  }

  function setPageSize(size: number) {
    if (!validatePageSize(size)) {
      throw new Error('Page size must be greater than 0')
    }

    pageSize.value = size

    // Reset to first page with new page size
    firstPage()
  }

  /**
   * Restore an exact position. `offset` only feeds the "Viser 16–30" label; the
   * cursor is what actually selects the rows, so a stale offset mislabels a
   * correct page rather than showing the wrong one.
   */
  function restore(state: {
    after?: string | null
    before?: string | null
    offset?: number
  }) {
    if (state.after) {
      variables.value = {
        first: pageSize.value,
        after: state.after,
        last: null,
        before: null,
      }
    } else if (state.before) {
      variables.value = {
        first: null,
        after: null,
        last: pageSize.value,
        before: state.before,
      }
    } else {
      return
    }

    offset.value = Math.max(0, state.offset ?? 0)
    pageInfo.value = null
    rowsOnPage.value = 0
  }

  const range = computed(() =>
    pageRange(offset.value, rowsOnPage.value, totalCount.value),
  )

  return {
    variables,
    pageInfo,
    totalCount,
    pageSize,
    range,
    isFirstPage,
    isLastPage,
    nextPage,
    previousPage,
    firstPage,
    updateConnection,
    reset,
    setPageSize,
    restore,
    offset,
  }
}
