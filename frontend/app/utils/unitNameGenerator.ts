/**
 * Generate unique unit names using the church name as prefix with sequential numbers.
 * E.g., "Østfold 1", "Østfold 2", etc.
 *
 * @param count - Number of names to generate
 * @param churchName - The church name to use as prefix
 * @param existingNames - Names that are already taken (to avoid duplicates)
 */
export function generateUniqueNames(
  count: number,
  churchName: string,
  existingNames: string[] = [],
): string[] {
  const names: string[] = []
  const existingSet = new Set(existingNames.map((n) => n.toLowerCase()))

  let number = 1
  while (names.length < count) {
    const name = `${churchName} ${number}`

    if (!existingSet.has(name.toLowerCase())) {
      existingSet.add(name.toLowerCase())
      names.push(name)
    }

    number++
  }

  return names
}
