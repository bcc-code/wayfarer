import { describe, it, expect } from 'vitest'
import { readdirSync, statSync, readFileSync } from 'node:fs'
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

const APP = resolve(__dirname, '../../app')

/** Paths that make up the admin domain; kept in step with eslint.config.mjs. */
const ADMIN_DOMAIN = [
  'pages/admin/',
  'components/admin/',
  'layouts/admin.vue',
  'layouts/church-admin.vue',
  'composables/useAdminNav.ts',
  'utils/adminNav.ts',
]

/** Generated or vendored files that are not hand-written source. */
const EXCLUDED = ['api/generated.ts']

function walk(dir: string): string[] {
  return readdirSync(dir).flatMap((entry) => {
    const path = join(dir, entry)
    if (statSync(path).isDirectory()) return walk(path)
    return /\.(vue|ts)$/.test(path) ? [path] : []
  })
}

function isAdminDomain(relPath: string): boolean {
  return ADMIN_DOMAIN.some((p) =>
    p.endsWith('/') ? relPath.startsWith(p) : relPath === p,
  )
}

describe('admin/user domain boundary', () => {
  const userFacingFiles = walk(APP)
    .map((f) => relative(APP, f))
    .filter((f) => !isAdminDomain(f) && !EXCLUDED.includes(f))

  it('has user-facing files to check', () => {
    // Guards against the walk silently matching nothing and the suite passing
    // vacuously.
    expect(userFacingFiles.length).toBeGreaterThan(50)
  })

  it('no user-facing or shared file uses an admin component or composable', () => {
    const violations: string[] = []

    for (const relPath of userFacingFiles) {
      const text = readFileSync(join(APP, relPath), 'utf8')
      const hits = new Set<string>()

      // `<AdminFoo ...>` / `<AdminFoo/>` in a template
      for (const m of text.matchAll(/<(Admin[A-Z][A-Za-z0-9]*)[\s/>]/g)) {
        hits.add(`<${m[1]}>`)
      }
      // `useAdminFoo(` auto-imported composable
      for (const m of text.matchAll(/\b(useAdmin[A-Z][A-Za-z0-9]*)\s*\(/g)) {
        hits.add(`${m[1]}()`)
      }

      if (hits.size) violations.push(`${relPath}: ${[...hits].join(', ')}`)
    }

    expect(violations).toEqual([])
  })
})
