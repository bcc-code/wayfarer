/**
 * Fuzzy search implementation for matching user queries against target strings.
 * Returns a score indicating match quality, or null if no match.
 *
 * Scoring:
 * - Exact match: 1000 points
 * - Contains match: 500+ points (with bonuses for start/word boundary)
 * - Fuzzy match: Variable points based on consecutive matches, word boundaries, and position
 */
export function fuzzyMatch(query: string, target: string): number | null {
  const queryLower = query.toLowerCase()
  const targetLower = target.toLowerCase()

  // Exact match gets highest score
  if (targetLower === queryLower) return 1000

  // Contains match gets good score
  if (targetLower.includes(queryLower)) {
    // Bonus for match at start
    const startBonus = targetLower.startsWith(queryLower) ? 100 : 0
    // Bonus for word boundary match
    const wordBoundaryBonus = targetLower
      .split(/\s+/)
      .some((word) => word.startsWith(queryLower))
      ? 50
      : 0
    return 500 + startBonus + wordBoundaryBonus
  }

  // Fuzzy match: all query chars must appear in order
  let queryIndex = 0
  let score = 0
  let lastMatchIndex = -1
  let consecutiveMatches = 0

  for (
    let i = 0;
    i < targetLower.length && queryIndex < queryLower.length;
    i++
  ) {
    if (targetLower[i] === queryLower[queryIndex]) {
      // Bonus for consecutive matches
      if (lastMatchIndex === i - 1) {
        consecutiveMatches++
        score += consecutiveMatches * 5
      } else {
        consecutiveMatches = 0
      }

      // Bonus for word boundary matches (after space, dash, underscore)
      const isWordBoundary =
        i === 0 || [' ', '-', '_'].includes(targetLower.charAt(i - 1))
      if (isWordBoundary) {
        score += 10
      }

      // Small bonus for earlier matches
      score += Math.max(0, 10 - Math.floor(i / 5))

      lastMatchIndex = i
      queryIndex++
    }
  }

  // Return null if not all query chars were matched
  if (queryIndex < queryLower.length) return null

  return score
}

/**
 * Search and score an array of items using fuzzy matching.
 * Returns items sorted by score (highest first).
 */
export function fuzzySearch<T>(
  items: T[],
  query: string,
  getSearchableStrings: (item: T) => (string | null | undefined)[],
): T[] {
  const trimmedQuery = query.trim()

  if (!trimmedQuery) {
    return items
  }

  // Score each item
  const scoredItems = items
    .map((item) => {
      const searchableStrings = getSearchableStrings(item)
      const scores = searchableStrings
        .filter((s): s is string => !!s)
        .map((s) => fuzzyMatch(trimmedQuery, s))
        .filter((s): s is number => s !== null)

      const bestScore = scores.length > 0 ? Math.max(...scores) : null
      return { item, score: bestScore }
    })
    .filter((entry) => entry.score !== null)

  // Sort by score (highest first)
  return scoredItems
    .sort((a, b) => (b.score ?? 0) - (a.score ?? 0))
    .map((entry) => entry.item)
}
