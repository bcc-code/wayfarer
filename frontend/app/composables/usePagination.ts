import { ref, computed, type Ref } from 'vue'

export interface PaginationPageInfo {
  hasNextPage: boolean
  hasPreviousPage: boolean
  startCursor?: string | null
  endCursor?: string | null
}

export interface Edge<T> {
  cursor: string
  node: T
}

export interface Connection<T> {
  edges: Edge<T>[]
  pageInfo: PageInfo
  totalCount?: number
}

export interface PaginationVariables {
  first?: number | null
  after?: string | null
  last?: number | null
  before?: string | null
}

export interface UsePaginationOptions {
  /** Default page size for forward pagination */
  defaultPageSize?: number
  /** Initial cursor to start from */
  initialCursor?: string | null
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
  const { defaultPageSize = 20, initialCursor = null } = options

  // Reactive state
  const pageSize = ref(defaultPageSize)
  const pageInfo = ref<PaginationPageInfo | null>(null)
  const totalCount = ref<number | null>(null)
  const variables = ref<PaginationVariables>({
    first: defaultPageSize,
    after: initialCursor,
  })

  // Computed properties
  const isFirstPage = computed(() => {
    return !pageInfo.value?.hasPreviousPage
  })

  const isLastPage = computed(() => {
    return !pageInfo.value?.hasNextPage
  })

  // Methods
  function nextPage() {
    if (!pageInfo.value?.hasNextPage || !pageInfo.value?.endCursor) {
      return
    }

    variables.value = {
      first: pageSize.value,
      after: pageInfo.value.endCursor,
      last: null,
      before: null,
    }
  }

  function previousPage() {
    if (!pageInfo.value?.hasPreviousPage || !pageInfo.value?.startCursor) {
      return
    }

    variables.value = {
      first: null,
      after: null,
      last: pageSize.value,
      before: pageInfo.value.startCursor,
    }
  }

  function firstPage() {
    variables.value = {
      first: pageSize.value,
      after: initialCursor,
      last: null,
      before: null,
    }
    pageInfo.value = null
    totalCount.value = null
  }

  function updateConnection<T>(connection: Connection<T> | null | undefined) {
    if (!connection) {
      pageInfo.value = null
      totalCount.value = null
      return
    }

    pageInfo.value = connection.pageInfo
    totalCount.value = connection.totalCount ?? null
  }

  function reset() {
    pageInfo.value = null
    totalCount.value = null
    variables.value = {
      first: pageSize.value,
      after: initialCursor,
      last: null,
      before: null,
    }
  }

  function setPageSize(size: number) {
    if (size <= 0) {
      throw new Error('Page size must be greater than 0')
    }

    pageSize.value = size

    // Reset to first page with new page size
    firstPage()
  }

  return {
    variables,
    pageInfo,
    totalCount,
    pageSize,
    isFirstPage,
    isLastPage,
    nextPage,
    previousPage,
    firstPage,
    updateConnection,
    reset,
    setPageSize,
  }
}
