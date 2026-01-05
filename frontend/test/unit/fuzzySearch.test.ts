import { describe, it, expect } from 'vitest'
import { fuzzyMatch, fuzzySearch } from '../../app/utils/fuzzySearch'

describe('fuzzySearch', () => {
  describe('fuzzyMatch', () => {
    describe('exact matches', () => {
      it('should return highest score for exact match', () => {
        expect(fuzzyMatch('hello', 'hello')).toBe(1000)
      })

      it('should be case insensitive for exact match', () => {
        expect(fuzzyMatch('Hello', 'hello')).toBe(1000)
        expect(fuzzyMatch('HELLO', 'hello')).toBe(1000)
        expect(fuzzyMatch('hello', 'HELLO')).toBe(1000)
      })
    })

    describe('contains matches', () => {
      it('should return good score for contains match', () => {
        const score = fuzzyMatch('ello', 'hello')
        expect(score).toBeGreaterThanOrEqual(500)
        expect(score).toBeLessThan(1000)
      })

      it('should give bonus for match at start', () => {
        const startScore = fuzzyMatch('hel', 'hello world')
        const middleScore = fuzzyMatch('wor', 'hello world')
        expect(startScore).toBeGreaterThan(middleScore!)
      })

      it('should give bonus for word boundary match', () => {
        const wordBoundaryScore = fuzzyMatch('wor', 'hello world')
        const middleScore = fuzzyMatch('llo', 'hello world')
        expect(wordBoundaryScore).toBeGreaterThan(middleScore!)
      })

      it('should handle substring at different positions', () => {
        expect(fuzzyMatch('test', 'this is a test')).toBeGreaterThan(0)
        expect(fuzzyMatch('is', 'this is a test')).toBeGreaterThan(0)
      })
    })

    describe('fuzzy matches', () => {
      it('should match characters in order', () => {
        expect(fuzzyMatch('hw', 'hello world')).not.toBeNull()
        expect(fuzzyMatch('hlo', 'hello')).not.toBeNull()
        expect(fuzzyMatch('ep1', 'Episode 1')).not.toBeNull()
      })

      it('should return null when characters are not in order', () => {
        expect(fuzzyMatch('wh', 'hello world')).toBeNull()
        expect(fuzzyMatch('olh', 'hello')).toBeNull()
      })

      it('should return null when query has extra characters', () => {
        expect(fuzzyMatch('hellox', 'hello')).toBeNull()
        expect(fuzzyMatch('xyz', 'hello')).toBeNull()
      })

      it('should give bonus for consecutive matches', () => {
        const consecutiveScore = fuzzyMatch('hel', 'h-e-l-l-o')
        const spreadScore = fuzzyMatch('heo', 'h-e-l-l-o')
        // Both should match, but consecutive should be higher
        expect(consecutiveScore).not.toBeNull()
        expect(spreadScore).not.toBeNull()
        // Note: 'hel' will be a substring match, 'heo' will be fuzzy
        // Let's use a better example
        const consec = fuzzyMatch('abc', 'xabcdef')
        const spread = fuzzyMatch('adf', 'xabcdef')
        expect(consec).toBeGreaterThan(spread!)
      })

      it('should give bonus for word boundary matches', () => {
        // 'ab' at word boundaries (Alice Bob) vs 'lb' not at word boundaries
        const wordBoundary = fuzzyMatch('ab', 'Alice Bob')
        // 'lb' matches l (in Alice) and b (in Bob) - only 'b' is at word boundary
        const partialBoundary = fuzzyMatch('lb', 'Alice Bob')
        expect(wordBoundary).toBeGreaterThan(partialBoundary!)
      })

      it('should give bonus for earlier matches', () => {
        const earlyMatch = fuzzyMatch('a', 'abcdefghij')
        const lateMatch = fuzzyMatch('j', 'abcdefghij')
        expect(earlyMatch).toBeGreaterThan(lateMatch!)
      })
    })

    describe('edge cases', () => {
      it('should handle empty query', () => {
        // Empty string is contained in any string, so it gets a contains match score
        // (the caller typically handles empty queries separately by returning all items)
        const score = fuzzyMatch('', 'hello')
        expect(score).toBeGreaterThanOrEqual(500) // Contains match score
      })

      it('should handle single character query', () => {
        expect(fuzzyMatch('h', 'hello')).not.toBeNull()
        expect(fuzzyMatch('x', 'hello')).toBeNull()
      })

      it('should handle query longer than target', () => {
        expect(fuzzyMatch('hello world', 'hello')).toBeNull()
      })

      it('should handle special characters', () => {
        expect(fuzzyMatch('test-1', 'test-123')).not.toBeNull()
        expect(fuzzyMatch('test_1', 'test_123')).not.toBeNull()
      })

      it('should handle unicode characters', () => {
        expect(fuzzyMatch('café', 'café')).toBe(1000)
        expect(fuzzyMatch('caf', 'café')).not.toBeNull()
      })

      it('should handle numbers', () => {
        expect(fuzzyMatch('123', '123')).toBe(1000)
        expect(fuzzyMatch('13', '123')).not.toBeNull()
      })
    })

    describe('real-world examples', () => {
      it('should match episode search patterns', () => {
        expect(fuzzyMatch('ep1', 'Episode 1: The Beginning')).not.toBeNull()
        expect(fuzzyMatch('ep 1', 'Episode 1: The Beginning')).not.toBeNull()
        expect(fuzzyMatch('beginning', 'Episode 1: The Beginning')).not.toBeNull()
      })

      it('should match bible verse patterns', () => {
        expect(fuzzyMatch('john 3', 'John 3:16')).not.toBeNull()
        expect(fuzzyMatch('jn3', 'John 3:16')).not.toBeNull()
        expect(fuzzyMatch('316', 'John 3:16')).not.toBeNull()
      })

      it('should match chapter patterns', () => {
        expect(fuzzyMatch('ch5', 'Chapter 5')).not.toBeNull()
        expect(fuzzyMatch('chap', 'Chapter 5')).not.toBeNull()
      })

      it('should rank matches at start higher', () => {
        // Both are substring matches, but 'test' at start should score higher
        const atStart = fuzzyMatch('test', 'test file')
        const inMiddle = fuzzyMatch('file', 'a test file')
        expect(atStart).toBeGreaterThan(inMiddle!)
      })
    })
  })

  describe('fuzzySearch', () => {
    interface TestItem {
      id: string
      title: string
    }

    const items: TestItem[] = [
      { id: '1', title: 'Episode 1: The Beginning' },
      { id: '2', title: 'Episode 2: The Middle' },
      { id: '3', title: 'Chapter 1' },
      { id: '4', title: 'John 3:16' },
      { id: '5', title: 'Test Document' },
    ]

    it('should return all items for empty query', () => {
      const results = fuzzySearch(items, '', (item) => [item.title])
      expect(results).toHaveLength(5)
    })

    it('should return all items for whitespace-only query', () => {
      const results = fuzzySearch(items, '   ', (item) => [item.title])
      expect(results).toHaveLength(5)
    })

    it('should filter and sort by score', () => {
      const results = fuzzySearch(items, 'ep', (item) => [item.title])
      expect(results.length).toBeGreaterThan(0)
      // Episode items should be first
      expect(results[0].title).toContain('Episode')
    })

    it('should match against multiple searchable strings', () => {
      const results = fuzzySearch(items, '1', (item) => [item.id, item.title])
      expect(results.length).toBeGreaterThan(0)
      // Should find items with '1' in id or title
      expect(results.some((r) => r.id === '1')).toBe(true)
    })

    it('should handle null/undefined in searchable strings', () => {
      const itemsWithNull = [
        { id: '1', title: 'Test', subtitle: null as string | null },
        { id: '2', title: 'Other', subtitle: undefined as string | undefined },
      ]
      const results = fuzzySearch(itemsWithNull, 'test', (item) => [
        item.title,
        item.subtitle,
      ])
      expect(results).toHaveLength(1)
      expect(results[0].title).toBe('Test')
    })

    it('should return empty array when no matches', () => {
      const results = fuzzySearch(items, 'xyz123', (item) => [item.title])
      expect(results).toHaveLength(0)
    })

    it('should sort results by best score', () => {
      const results = fuzzySearch(items, 'episode', (item) => [item.title])
      expect(results.length).toBe(2)
      // Both episode items should match
      expect(results.every((r) => r.title.includes('Episode'))).toBe(true)
    })
  })
})
