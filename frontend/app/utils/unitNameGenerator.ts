const adjectives = [
  'Swift',
  'Brave',
  'Mighty',
  'Wise',
  'Bold',
  'Noble',
  'Fierce',
  'Clever',
  'Golden',
  'Proud',
  'Wild',
  'Tough',
  'Silent',
  'Sharp',
  'Cool',
  'Blazing',
  'Royal',
  'Iron',
  'Cunning',
  'Valiant',
]

const nouns = [
  'Vikings',
  'Lions',
  'Eagles',
  'Bears',
  'Wolves',
  'Stags',
  'Falcons',
  'Sharks',
  'Tigers',
  'Moose',
  'Ravens',
  'Hawks',
  'Swans',
  'Kings',
  'Knights',
  'Hunters',
  'Heroes',
  'Dragons',
  'Pirates',
  'Panthers',
]

export function generateUniqueNames(count: number): string[] {
  const names: string[] = []
  const usedCombinations = new Set<string>()

  const maxPossible = adjectives.length * nouns.length
  const actualCount = Math.min(count, maxPossible)

  while (names.length < actualCount) {
    const adjective = adjectives[Math.floor(Math.random() * adjectives.length)]
    const noun = nouns[Math.floor(Math.random() * nouns.length)]
    const name = `${adjective} ${noun}`

    if (!usedCombinations.has(name)) {
      usedCombinations.add(name)
      names.push(name)
    }
  }

  return names
}
