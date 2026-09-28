import { describe, it, expect } from 'vitest'
import { readdirSync, statSync, readFileSync, existsSync } from 'node:fs'
import { join, resolve, relative } from 'node:path'

/**
 * Root `app/` is the SHARED base layer. A module used by only one domain does
 * not belong in it.
 *
 * This is drift nothing else catches. Auto-imports mean a module keeps working
 * wherever it sits, so a util can quietly serve one layer alone for months —
 * which is exactly what happened: `dates`, `fuzzySearch`, `languageMapping`,
 * `unitNameGenerator`, `pagination` and `usePagination` were all admin-only and
 * all still sitting in the shared root after the admin layer was extracted.
 *
 * Only `utils/` and `composables/` are checked. They are the auto-import
 * registries, and they are the only root directories where a domain module can
 * hide — `pages/`, `plugins/` and `middleware/` are structural.
 */

const FE = resolve(__dirname, '../..')
const ROOTS = {
  shared: join(FE, 'app'),
  user: join(FE, 'layers/user/app'),
  admin: join(FE, 'layers/admin/app'),
}

/**
 * Modules that serve one domain but must stay in root anyway, with the reason.
 * Each is load-bearing for a ROOT file, so moving it would make the shared base
 * import from a layer — the one direction the boundary forbids.
 */
const PINNED_TO_ROOT: Record<string, string> = {
  'utils/adminPermissions.ts':
    'the permission boundary itself; read by 02.admin-permission.global.ts',
  'utils/permissions.ts':
    'role helpers behind usePermissions, which the root guard calls',
  'composables/usePermissions.ts': 'called by 02.admin-permission.global.ts',
}

function walk(dir: string): string[] {
  if (!existsSync(dir)) return []
  return readdirSync(dir).flatMap((e) => {
    const p = join(dir, e)
    return statSync(p).isDirectory()
      ? walk(p)
      : /\.(ts|vue)$/.test(p) && !p.endsWith('.d.ts')
        ? [p]
        : []
  })
}

/** Top-level exported names — what an auto-import consumer would reference. */
function exportedNames(file: string): string[] {
  const src = readFileSync(file, 'utf8')
  const names = new Set<string>()
  for (const m of src.matchAll(
    /^export\s+(?:async\s+)?(?:const|function|class|type|interface|enum)\s+([A-Za-z_$][\w$]*)/gm,
  )) {
    names.add(m[1]!)
  }
  for (const m of src.matchAll(/^export\s*\{([^}]*)\}/gm)) {
    for (const part of m[1]!.split(',')) {
      const name = part
        .trim()
        .split(/\s+as\s+/)
        .pop()
        ?.trim()
      if (name) names.add(name)
    }
  }
  return [...names]
}

const CANDIDATES = [
  join(ROOTS.shared, 'utils'),
  join(ROOTS.shared, 'composables'),
]
  .flatMap(walk)
  .map((f) => relative(ROOTS.shared, f))

describe('root app/ holds only shared code', () => {
  it('has candidates to check', () => {
    // Guards against the walk matching nothing and the suite passing vacuously.
    expect(CANDIDATES.length).toBeGreaterThan(10)
  })

  it.each(CANDIDATES)('%s is used by more than one domain', (rel) => {
    const file = join(ROOTS.shared, rel)
    const names = exportedNames(file)
    // A module with no detectable exports tells us nothing; skip rather than
    // assert something false about it.
    if (!names.length) return

    const word = new RegExp(
      `\\b(${names.map((n) => n.replace(/\$/g, '\\$')).join('|')})\\b`,
    )
    const used = (root: string) =>
      walk(root).some((f) => f !== file && word.test(readFileSync(f, 'utf8')))

    const domains = (['user', 'admin'] as const).filter((d) => used(ROOTS[d]))
    // A reference from a root file is what makes a module load-bearing for the
    // shared base, whichever domain also uses it.
    if (used(ROOTS.shared)) return
    if (rel in PINNED_TO_ROOT) return

    expect(
      domains.length === 0 ? 'unreferenced' : domains.join('+'),
      `app/${rel} is reachable only from layers/${domains.join('+')}. Move it into that layer, or add it to PINNED_TO_ROOT with the reason it must stay.`,
    ).toBe('user+admin')
  })

  it('every PINNED_TO_ROOT entry still exists', () => {
    // Otherwise the exception list silently outlives what it excuses.
    for (const rel of Object.keys(PINNED_TO_ROOT)) {
      expect(
        existsSync(join(ROOTS.shared, rel)),
        `${rel} is pinned but missing`,
      ).toBe(true)
    }
  })
})
