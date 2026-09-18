import { describe, it, expect } from 'vitest'
import { readdirSync, statSync, readFileSync, existsSync } from 'node:fs'
import { join, resolve, relative } from 'node:path'

/**
 * Domain boundary: user-facing and shared code must not reach into admin code.
 *
 * The ESLint rule in `eslint.config.mjs` covers explicit `import` statements.
 * It cannot cover the vector that actually matters in a Nuxt app: **auto-imports**.
 * `<AdminUserMenu />` in a template and `useAdminNav()` in a script have no
 * import statement at all — Nuxt merges every layer's `components/`,
 * `composables/` and `utils/` into single global registries
 * (`nuxt/dist/index.mjs:3558` and `:3843`), so nothing stops a user-facing page
 * from using an admin component.
 *
 * This test closes that gap by scanning source text. It stays valid after the
 * planned move to Nuxt layers — layers organise files but do not isolate these
 * registries, so the check is still needed then.
 *
 * Admin -> user is deliberately allowed: the admin panel renders user-facing
 * components to preview the end-user experience.
 */

const FE = resolve(__dirname, '../..')
/**
 * Everything that must not reach into admin: the shared root and the
 * user-facing layer. The admin panel is a directory now, so it is simply not
 * scanned.
 */
const SCANNED = [join(FE, 'app'), join(FE, 'layers/user/app')]
const ADMIN_LAYER = join(FE, 'layers/admin/app')

/** Generated or vendored files that are not hand-written source. */
const EXCLUDED = ['app/api/generated.ts']

/**
 * Admin-layer composables whose names the `useAdmin*` scan below would miss.
 * `useConfirm` is the one that matters: its dialog is mounted by the admin
 * layout, so a user-facing caller gets a promise that never resolves — a hang,
 * not an error.
 */
const ADMIN_ONLY_COMPOSABLES = [
  'useCurrentProject',
  'useConfirm',
  'useGroupedProjects',
  'usePagination',
]

function walk(dir: string): string[] {
  return readdirSync(dir).flatMap((entry) => {
    const path = join(dir, entry)
    if (statSync(path).isDirectory()) return walk(path)
    return /\.(vue|ts)$/.test(path) ? [path] : []
  })
}

describe('admin/user domain boundary', () => {
  const userFacingFiles = SCANNED.flatMap(walk)
    .map((f) => relative(FE, f))
    .filter((f) => !EXCLUDED.includes(f))

  it('has user-facing files to check', () => {
    // Guards against the walk silently matching nothing and the suite passing
    // vacuously. Both roots must contribute: scanning only the shared root
    // would miss every user-facing page.
    expect(SCANNED.every((d) => existsSync(d))).toBe(true)
    expect(
      userFacingFiles.filter((f) => f.startsWith('app/')).length,
    ).toBeGreaterThan(20)
    expect(
      userFacingFiles.filter((f) => f.startsWith('layers/user/')).length,
    ).toBeGreaterThan(50)
  })

  it('the admin layer is where it is meant to be', () => {
    // Without this the scan above could pass vacuously against a renamed tree.
    expect(existsSync(ADMIN_LAYER)).toBe(true)
    expect(walk(ADMIN_LAYER).length).toBeGreaterThan(70)
  })

  it('no user-facing or shared file uses an admin component or composable', () => {
    const violations: string[] = []

    for (const relPath of userFacingFiles) {
      const text = readFileSync(join(FE, relPath), 'utf8')
      const hits = new Set<string>()

      // `<AdminFoo ...>` / `<AdminFoo/>` in a template
      for (const m of text.matchAll(/<(Admin[A-Z][A-Za-z0-9]*)[\s/>]/g)) {
        hits.add(`<${m[1]}>`)
      }
      // `useAdminFoo(` auto-imported composable
      for (const m of text.matchAll(/\b(useAdmin[A-Z][A-Za-z0-9]*)\s*\(/g)) {
        hits.add(`${m[1]}()`)
      }
      // ...and the admin-only ones that do not start with `useAdmin`.
      for (const name of ADMIN_ONLY_COMPOSABLES) {
        if (new RegExp(`\\b${name}\\s*\\(`).test(text)) hits.add(`${name}()`)
      }

      if (hits.size) violations.push(`${relPath}: ${[...hits].join(', ')}`)
    }

    expect(violations).toEqual([])
  })
})

/**
 * The ESLint half of the boundary, tested against the real config.
 *
 * `no-restricted-imports` matches its patterns with gitignore semantics, where
 * an unescaped leading `#` marks a *comment* — so `'#layers/admin/**'` is
 * dropped silently, with no config error and no failing lint, and every
 * alias-form import goes uncaught. That is not hypothetical: it is how the rule
 * shipped, and nothing detected it. These cases pin both spellings.
 */
describe('the eslint import boundary', () => {
  const IMPORTS = [
    '#layers/admin/app/utils/adminNav',
    '../../layers/admin/app/utils/adminNav',
  ]

  /**
   * `eslint.config.mjs` exports a Nuxt FlatConfigComposer, not a plain array.
   * It is thenable, so awaiting it yields the resolved config list. Resolved
   * once and shared — it is not cheap.
   */
  const entry = (async () => {
    const mod = await import('../../eslint.config.mjs')
    const configs = (await mod.default) as {
      name?: string
      files?: string[]
      rules?: Record<string, unknown>
    }[]
    expect(Array.isArray(configs), 'resolved eslint config is a list').toBe(
      true,
    )
    const found = configs.find((c) => c.name === 'interact/domain-boundary')
    expect(found, 'the domain-boundary config entry').toBeDefined()
    return found!
  })()

  async function lint(code: string) {
    const { Linter } = await import('eslint')
    const rule = (await entry).rules!['no-restricted-imports']
    return new Linter().verify(code, {
      rules: { 'no-restricted-imports': rule as never },
    })
  }

  it.each(IMPORTS)('rejects an import of %s', async (spec) => {
    const messages = await lint(
      `import { GLOBAL_NAV } from '${spec}'\nexport const x = GLOBAL_NAV\n`,
    )
    expect(messages.map((m) => m.ruleId)).toEqual(['no-restricted-imports'])
  })

  it('still allows importing shared and user-facing code', async () => {
    // Admin -> user is the sanctioned direction; the rule must not block it.
    const messages = await lint(
      `import { x } from '~/utils/formatters'\nimport { y } from '#layers/user/app/utils/teams'\nexport const z = [x, y]\n`,
    )
    expect(messages).toEqual([])
  })

  it('covers both non-admin roots', async () => {
    // User-facing code moved into a layer; a rule scoped to `app/**` alone
    // would no longer see it.
    expect((await entry).files).toEqual([
      'app/**/*.{vue,ts}',
      'layers/user/**/*.{vue,ts}',
    ])
  })
})
