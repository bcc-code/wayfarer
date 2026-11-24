import { describe, it, expect } from 'vitest'
import { readdirSync, statSync, readFileSync, existsSync } from 'node:fs'
import { join, resolve, relative } from 'node:path'
import { DISPATCHED_QUERY_DOCUMENTS } from '../../app/utils/firestoreNotifications'

/**
 * A page refetches live by listening for query documents with
 * useFirestoreRefresh, and useFirestoreSync only dispatches the documents its
 * notification maps list. A listened-for document that no map lists never
 * fires, with no error: the Completed challenges tab stayed stale that way.
 */

const FE = resolve(__dirname, '../..')
const SOURCE_ROOTS = [
  join(FE, 'app'),
  ...readdirSync(join(FE, 'layers'), { withFileTypes: true })
    .filter((e) => e.isDirectory())
    .map((e) => join(FE, 'layers', e.name, 'app'))
    .filter((dir) => existsSync(dir)),
]

function walk(dir: string): string[] {
  return readdirSync(dir).flatMap((entry) => {
    const path = join(dir, entry)
    return statSync(path).isDirectory()
      ? walk(path)
      : /\.(vue|ts)$/.test(path)
        ? [path]
        : []
  })
}

// document name -> files that listen for it
function listenedDocuments(): Map<string, string[]> {
  const listened = new Map<string, string[]>()
  for (const file of SOURCE_ROOTS.flatMap(walk)) {
    const src = readFileSync(file, 'utf8')
    for (const call of src.matchAll(/useFirestoreRefresh\(\s*\[([^\]]*)\]/g)) {
      for (const name of call[1]!.matchAll(/['"`](\w+)['"`]/g)) {
        const files = listened.get(name[1]!) ?? []
        files.push(relative(FE, file))
        listened.set(name[1]!, files)
      }
    }
  }
  return listened
}

describe('Firestore notification maps', () => {
  const listened = listenedDocuments()

  it('finds the useFirestoreRefresh listeners', () => {
    // Guards against the scan matching nothing and the suite passing vacuously
    expect(listened.size).toBeGreaterThanOrEqual(5)
  })

  it.each([...listened.keys()])(
    '%s is dispatched by a notification map',
    (name) => {
      expect(
        DISPATCHED_QUERY_DOCUMENTS.has(name),
        `${listened.get(name)!.join(', ')} listen for ${name}, but no map in app/utils/firestoreNotifications.ts dispatches it, so it never refetches live`,
      ).toBe(true)
    },
  )

  it.each([...DISPATCHED_QUERY_DOCUMENTS])('%s has a listener', (name) => {
    expect(
      listened.has(name),
      `${name} is dispatched, but nothing listens for it with useFirestoreRefresh`,
    ).toBe(true)
  })

  it('dispatches CompletedChallengesPage wherever it dispatches ActiveChallengesPage', async () => {
    const maps = await import('../../app/utils/firestoreNotifications')
    for (const map of [
      maps.NOTIFICATION_QUERY_MAP,
      maps.PROJECT_NOTIFICATION_QUERY_MAP,
    ]) {
      for (const [category, docs] of Object.entries(map)) {
        const list = docs as readonly string[]
        expect(
          list.includes('ActiveChallengesPageDocument'),
          `category ${category}`,
        ).toBe(list.includes('CompletedChallengesPageDocument'))
      }
    }
  })
})
