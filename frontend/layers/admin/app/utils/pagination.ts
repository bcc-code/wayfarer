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

// A cursor says "after this row", never "row 60 of 13 477", so the position is
// counted client-side. Exact, since navigation only happens through the
// controls. Jumping to an arbitrary page still needs OFFSET.

export function nextOffset(offset: number, pageSize: number): number {
  return offset + pageSize
}

export function previousOffset(offset: number, pageSize: number): number {
  return Math.max(0, offset - pageSize)
}

export interface PageRange {
  /** Both 1-based. */
  start: number
  end: number
  total: number
}

// `end` comes from the rows actually returned, so a short final page does not
// overshoot the total.
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
