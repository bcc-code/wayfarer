import { describe, it, expect } from 'vitest'
import { usePagination } from '../../app/composables/usePagination'
import type { Connection } from '../../app/composables/usePagination'

describe('usePagination', () => {
  describe('initialization', () => {
    it('should initialize with default page size', () => {
      const pagination = usePagination()

      expect(pagination.pageSize.value).toBe(20)
      expect(pagination.variables.value).toEqual({
        first: 20,
        after: null,
      })
    })

    it('should initialize with custom page size', () => {
      const pagination = usePagination({ defaultPageSize: 50 })

      expect(pagination.pageSize.value).toBe(50)
      expect(pagination.variables.value).toEqual({
        first: 50,
        after: null,
      })
    })

    it('should initialize with initial cursor', () => {
      const pagination = usePagination({
        defaultPageSize: 20,
        initialCursor: 'cursor123',
      })

      expect(pagination.variables.value).toEqual({
        first: 20,
        after: 'cursor123',
      })
    })

    it('should have null pageInfo and totalCount initially', () => {
      const pagination = usePagination()

      expect(pagination.pageInfo.value).toBeNull()
      expect(pagination.totalCount.value).toBeNull()
    })

    it('should compute isFirstPage as true initially', () => {
      const pagination = usePagination()

      expect(pagination.isFirstPage.value).toBe(true)
    })

    it('should compute isLastPage as true initially', () => {
      const pagination = usePagination()

      expect(pagination.isLastPage.value).toBe(true)
    })
  })

  describe('updateConnection', () => {
    it('should update pageInfo and totalCount from connection', () => {
      const pagination = usePagination()

      const connection: Connection<{ id: string }> = {
        edges: [
          { cursor: 'c1', node: { id: '1' } },
          { cursor: 'c2', node: { id: '2' } },
        ],
        pageInfo: {
          hasNextPage: true,
          hasPreviousPage: false,
          startCursor: 'c1',
          endCursor: 'c2',
        },
        totalCount: 100,
      }

      pagination.updateConnection(connection)

      expect(pagination.pageInfo.value).toEqual(connection.pageInfo)
      expect(pagination.totalCount.value).toBe(100)
    })

    it('should handle null connection', () => {
      const pagination = usePagination()

      pagination.updateConnection(null)

      expect(pagination.pageInfo.value).toBeNull()
      expect(pagination.totalCount.value).toBeNull()
    })

    it('should handle undefined connection', () => {
      const pagination = usePagination()

      pagination.updateConnection(undefined)

      expect(pagination.pageInfo.value).toBeNull()
      expect(pagination.totalCount.value).toBeNull()
    })

    it('should handle connection without totalCount', () => {
      const pagination = usePagination()

      const connection: Connection<{ id: string }> = {
        edges: [],
        pageInfo: {
          hasNextPage: false,
          hasPreviousPage: false,
        },
      }

      pagination.updateConnection(connection)

      expect(pagination.totalCount.value).toBeNull()
    })
  })

  describe('nextPage', () => {
    it('should update variables for next page', () => {
      const pagination = usePagination({ defaultPageSize: 20 })

      const connection: Connection<{ id: string }> = {
        edges: [{ cursor: 'c1', node: { id: '1' } }],
        pageInfo: {
          hasNextPage: true,
          hasPreviousPage: false,
          startCursor: 'c1',
          endCursor: 'c2',
        },
        totalCount: 100,
      }

      pagination.updateConnection(connection)
      pagination.nextPage()

      expect(pagination.variables.value).toEqual({
        first: 20,
        after: 'c2',
        last: null,
        before: null,
      })
    })

    it('should not update variables if no next page', () => {
      const pagination = usePagination({ defaultPageSize: 20 })

      const connection: Connection<{ id: string }> = {
        edges: [{ cursor: 'c1', node: { id: '1' } }],
        pageInfo: {
          hasNextPage: false,
          hasPreviousPage: false,
          startCursor: 'c1',
          endCursor: 'c2',
        },
        totalCount: 20,
      }

      pagination.updateConnection(connection)
      const beforeVariables = { ...pagination.variables.value }
      pagination.nextPage()

      expect(pagination.variables.value).toEqual(beforeVariables)
    })

    it('should not update variables if no endCursor', () => {
      const pagination = usePagination({ defaultPageSize: 20 })

      const connection: Connection<{ id: string }> = {
        edges: [],
        pageInfo: {
          hasNextPage: true,
          hasPreviousPage: false,
        },
        totalCount: 100,
      }

      pagination.updateConnection(connection)
      const beforeVariables = { ...pagination.variables.value }
      pagination.nextPage()

      expect(pagination.variables.value).toEqual(beforeVariables)
    })
  })

  describe('previousPage', () => {
    it('should update variables for previous page', () => {
      const pagination = usePagination({ defaultPageSize: 20 })

      const connection: Connection<{ id: string }> = {
        edges: [{ cursor: 'c2', node: { id: '2' } }],
        pageInfo: {
          hasNextPage: false,
          hasPreviousPage: true,
          startCursor: 'c1',
          endCursor: 'c2',
        },
        totalCount: 100,
      }

      pagination.updateConnection(connection)
      pagination.previousPage()

      expect(pagination.variables.value).toEqual({
        first: null,
        after: null,
        last: 20,
        before: 'c1',
      })
    })

    it('should not update variables if no previous page', () => {
      const pagination = usePagination({ defaultPageSize: 20 })

      const connection: Connection<{ id: string }> = {
        edges: [{ cursor: 'c1', node: { id: '1' } }],
        pageInfo: {
          hasNextPage: true,
          hasPreviousPage: false,
          startCursor: 'c1',
          endCursor: 'c2',
        },
        totalCount: 100,
      }

      pagination.updateConnection(connection)
      const beforeVariables = { ...pagination.variables.value }
      pagination.previousPage()

      expect(pagination.variables.value).toEqual(beforeVariables)
    })

    it('should not update variables if no startCursor', () => {
      const pagination = usePagination({ defaultPageSize: 20 })

      const connection: Connection<{ id: string }> = {
        edges: [],
        pageInfo: {
          hasNextPage: false,
          hasPreviousPage: true,
        },
        totalCount: 100,
      }

      pagination.updateConnection(connection)
      const beforeVariables = { ...pagination.variables.value }
      pagination.previousPage()

      expect(pagination.variables.value).toEqual(beforeVariables)
    })
  })

  describe('firstPage', () => {
    it('should reset to first page', () => {
      const pagination = usePagination({ defaultPageSize: 20 })

      // Navigate to second page
      const connection: Connection<{ id: string }> = {
        edges: [{ cursor: 'c1', node: { id: '1' } }],
        pageInfo: {
          hasNextPage: true,
          hasPreviousPage: false,
          startCursor: 'c1',
          endCursor: 'c2',
        },
        totalCount: 100,
      }

      pagination.updateConnection(connection)
      pagination.nextPage()

      // Go back to first page
      pagination.firstPage()

      expect(pagination.variables.value).toEqual({
        first: 20,
        after: null,
        last: null,
        before: null,
      })
      expect(pagination.pageInfo.value).toBeNull()
      expect(pagination.totalCount.value).toBeNull()
    })

    it('should respect initial cursor when going to first page', () => {
      const pagination = usePagination({
        defaultPageSize: 20,
        initialCursor: 'initial123',
      })

      pagination.firstPage()

      expect(pagination.variables.value.after).toBe('initial123')
    })
  })

  describe('reset', () => {
    it('should reset all state to initial values', () => {
      const pagination = usePagination({ defaultPageSize: 20 })

      // Set some state
      const connection: Connection<{ id: string }> = {
        edges: [{ cursor: 'c1', node: { id: '1' } }],
        pageInfo: {
          hasNextPage: true,
          hasPreviousPage: false,
          startCursor: 'c1',
          endCursor: 'c2',
        },
        totalCount: 100,
      }

      pagination.updateConnection(connection)
      pagination.nextPage()

      // Reset
      pagination.reset()

      expect(pagination.variables.value).toEqual({
        first: 20,
        after: null,
        last: null,
        before: null,
      })
      expect(pagination.pageInfo.value).toBeNull()
      expect(pagination.totalCount.value).toBeNull()
    })
  })

  describe('setPageSize', () => {
    it('should update page size and reset to first page', () => {
      const pagination = usePagination({ defaultPageSize: 20 })

      pagination.setPageSize(50)

      expect(pagination.pageSize.value).toBe(50)
      expect(pagination.variables.value).toEqual({
        first: 50,
        after: null,
        last: null,
        before: null,
      })
    })

    it('should throw error for invalid page size', () => {
      const pagination = usePagination()

      expect(() => pagination.setPageSize(0)).toThrow(
        'Page size must be greater than 0',
      )
      expect(() => pagination.setPageSize(-1)).toThrow(
        'Page size must be greater than 0',
      )
    })
  })

  describe('computed properties', () => {
    it('should compute isFirstPage correctly', () => {
      const pagination = usePagination()

      // Initially true
      expect(pagination.isFirstPage.value).toBe(true)

      // After setting pageInfo with hasPreviousPage = false
      pagination.updateConnection({
        edges: [],
        pageInfo: {
          hasNextPage: true,
          hasPreviousPage: false,
        },
      })
      expect(pagination.isFirstPage.value).toBe(true)

      // After setting pageInfo with hasPreviousPage = true
      pagination.updateConnection({
        edges: [],
        pageInfo: {
          hasNextPage: true,
          hasPreviousPage: true,
        },
      })
      expect(pagination.isFirstPage.value).toBe(false)
    })

    it('should compute isLastPage correctly', () => {
      const pagination = usePagination()

      // Initially true
      expect(pagination.isLastPage.value).toBe(true)

      // After setting pageInfo with hasNextPage = false
      pagination.updateConnection({
        edges: [],
        pageInfo: {
          hasNextPage: false,
          hasPreviousPage: true,
        },
      })
      expect(pagination.isLastPage.value).toBe(true)

      // After setting pageInfo with hasNextPage = true
      pagination.updateConnection({
        edges: [],
        pageInfo: {
          hasNextPage: true,
          hasPreviousPage: true,
        },
      })
      expect(pagination.isLastPage.value).toBe(false)
    })
  })
})
