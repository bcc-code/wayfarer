/**
 * Pure functions for cursor-based pagination calculations.
 * These functions are stateless and can be easily unit tested.
 */

export interface CursorPageInfo {
  hasNextPage: boolean
  hasPreviousPage: boolean
  startCursor?: string | null
  endCursor?: string | null
}

export interface CursorPaginationVariables {
  first?: number | null
  after?: string | null
  last?: number | null
  before?: string | null
}

/**
 * Build variables for navigating to the next page
 */
export function buildNextPageVariables(
  pageInfo: CursorPageInfo,
  pageSize: number,
): CursorPaginationVariables | null {
  if (!pageInfo.hasNextPage || !pageInfo.endCursor) {
    return null
  }

  return {
    first: pageSize,
    after: pageInfo.endCursor,
    last: null,
    before: null,
  }
}

/**
 * Build variables for navigating to the previous page
 */
export function buildPreviousPageVariables(
  pageInfo: CursorPageInfo,
  pageSize: number,
): CursorPaginationVariables | null {
  if (!pageInfo.hasPreviousPage || !pageInfo.startCursor) {
    return null
  }

  return {
    first: null,
    after: null,
    last: pageSize,
    before: pageInfo.startCursor,
  }
}

/**
 * Build variables for the first page
 */
export function buildFirstPageVariables(
  pageSize: number,
  initialCursor?: string | null,
): CursorPaginationVariables {
  return {
    first: pageSize,
    after: initialCursor ?? null,
    last: null,
    before: null,
  }
}

/**
 * Build variables for the last page (backward pagination - newest first)
 */
export function buildLastPageVariables(
  pageSize: number,
  initialCursor?: string | null,
): CursorPaginationVariables {
  return {
    first: null,
    after: null,
    last: pageSize,
    before: initialCursor ?? null,
  }
}

/**
 * Check if currently on the first page
 */
export function isFirstPage(pageInfo: CursorPageInfo | null): boolean {
  return !pageInfo?.hasPreviousPage
}

/**
 * Check if currently on the last page
 */
export function isLastPage(pageInfo: CursorPageInfo | null): boolean {
  return !pageInfo?.hasNextPage
}

/**
 * Validate page size
 */
export function validatePageSize(size: number): boolean {
  return size > 0
}

// ==================== Position within the result set ====================

/**
 * Keyset pagination knows how to step but not where it is: a cursor says
 * "after this row", never "row 60 of 13,477". The position is therefore
 * tracked client-side by counting steps, which is exact as long as navigation
 * only happens through the controls — which is the only way it can happen.
 *
 * This is what makes "Viser 16–30 av 13 477" possible without adding OFFSET to
 * the backend. Jumping to an arbitrary page still is not: that needs a cursor
 * we have no way to compute without walking there.
 */

export function nextOffset(offset: number, pageSize: number): number {
  return offset + pageSize
}

export function previousOffset(offset: number, pageSize: number): number {
  return Math.max(0, offset - pageSize)
}

export interface PageRange {
  /** 1-based index of the first row on screen. */
  start: number
  /** 1-based index of the last row on screen. */
  end: number
  total: number
}

/**
 * The human-readable range for the current page, or null when there is nothing
 * to describe.
 *
 * `end` is derived from the rows actually returned rather than from the page
 * size, so a short final page reads "13 471–13 477" instead of overshooting the
 * total.
 */
export function pageRange(
  offset: number,
  rowsOnPage: number,
  totalCount: number | null | undefined,
): PageRange | null {
  if (!totalCount || totalCount <= 0 || rowsOnPage <= 0) return null

  const start = offset + 1
  const end = Math.min(offset + rowsOnPage, totalCount)

  // A stale offset (data shrank under us) would otherwise print a range that
  // starts past the end.
  if (start > totalCount) return null

  return { start, end, total: totalCount }
}
