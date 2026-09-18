import { describe, it, expect } from 'vitest'
import { readdirSync, statSync, existsSync } from 'node:fs'
import { join, resolve } from 'node:path'

/**
 * Layer invariants that nothing else detects.
 *
 * Each of these fails silently: no build error, no type error, no failing test
 * anywhere else. They are cheap to assert and expensive to discover.
 */

const FE = resolve(__dirname, '../..')
const LAYERS = readdirSync(join(FE, 'layers'), { withFileTypes: true })
  .filter((e) => e.isDirectory())
  .map((e) => e.name)

function walk(dir: string, match: RegExp): string[] {
  if (!existsSync(dir)) return []
  return readdirSync(dir).flatMap((entry) => {
    const path = join(dir, entry)
    return statSync(path).isDirectory()
      ? walk(path, match)
      : match.test(path)
        ? [path]
        : []
  })
}

/** Loads a layer's config with the Nuxt macro stubbed to the identity. */
async function layerConfig(name: string) {
  const g = globalThis as { defineNuxtConfig?: (c: unknown) => unknown }
  g.defineNuxtConfig = (c) => c
  try {
    const mod = await import(`../../layers/${name}/nuxt.config.ts`)
    return mod.default as { components: { dirs: { path: string }[] } }
  } finally {
    delete g.defineNuxtConfig
  }
}

describe('nuxt layers', () => {
  it('finds the layers to check', () => {
    // Guards the whole suite against globbing nothing and passing vacuously.
    expect(LAYERS).toEqual(['admin', 'user'])
  })

  it.each(LAYERS)('the %s layer has a nuxt.config', (name) => {
    // A layer directory without one is dropped before `_layers.push`
    // (@nuxt/kit loadNuxtConfig) — no warning, the routes simply vanish.
    expect(existsSync(join(FE, 'layers', name, 'nuxt.config.ts'))).toBe(true)
  })

  it.each(LAYERS)(
    'the %s layer disables the components path prefix',
    async (name) => {
      // A layer that declares no `components` config falls through to
      // normalizeDirs' default, where pathPrefix is true — and every component in
      // the layer is renamed (AdminUserMenu -> AdminAdminUserMenu).
      //
      // The path must stay the bare relative string: normalizeDirs resolves it
      // against the layer's srcDir, whereas `~/components` would resolve through
      // the *global* alias map and point back at the root app.
      const { dirs } = (await layerConfig(name)).components
      expect(dirs[0]).toEqual({ path: 'components', pathPrefix: false })
    },
  )

  it('registers the user layer global components globally', async () => {
    // Three icons are passed as *strings* (`icon="IconSettings"`, `IconClose`,
    // `IconChevronRight`) and resolve by name alone. Without this entry they
    // render nothing, with no error anywhere.
    const { dirs } = (await layerConfig('user')).components
    expect(dirs).toContainEqual({ path: 'components/global', global: true })
  })

  it('keeps every global middleware in the root layer', () => {
    // Global middleware are gathered per layer, extended layers FIRST
    // (resolveApp iterates layerDirs.toReversed()), so a filename prefix only
    // orders them within one layer. A `*.global.ts` in a layer would run before
    // `01.auth.global.ts` regardless of its name — and the admin guard reads
    // `me`, so a signed-in superadmin would be bounced to '/'.
    expect(existsSync(join(FE, 'app/middleware/01.auth.global.ts'))).toBe(true)
    expect(
      existsSync(join(FE, 'app/middleware/02.admin-permission.global.ts')),
    ).toBe(true)
    expect(walk(join(FE, 'layers'), /\.global\.ts$/)).toEqual([])
  })

  it.each(LAYERS)('the %s layer defines no plugins or middleware', (name) => {
    // Plugins have the identical layer reversal as middleware, so a layer
    // plugin would run before `0.urql` installs the GraphQL client.
    //
    // Checked as directories under the layer's srcDir rather than by matching
    // `/plugins/` anywhere in the path: `pages/plugins/ladder-to-heaven/` is a
    // page route, not a Nuxt plugin dir.
    expect(existsSync(join(FE, 'layers', name, 'app/plugins'))).toBe(false)
    expect(existsSync(join(FE, 'layers', name, 'app/middleware'))).toBe(false)
  })
})
