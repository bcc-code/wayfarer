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

describe('nuxt layers', () => {
  it('the admin layer has a nuxt.config', () => {
    // A layer directory without one is dropped before `_layers.push`
    // (@nuxt/kit loadNuxtConfig) — no warning, the routes simply vanish.
    expect(existsSync(join(FE, 'layers/admin/nuxt.config.ts'))).toBe(true)
  })

  it('the admin layer disables the components path prefix', async () => {
    // A layer that declares no `components` config falls through to
    // normalizeDirs' default, where pathPrefix is true — and every component in
    // the layer is renamed (AdminUserMenu -> AdminAdminUserMenu).
    //
    // The path must stay the bare relative string: normalizeDirs resolves it
    // against the layer's srcDir, whereas `~/components` would resolve through
    // the *global* alias map and point back at the root app.
    const g = globalThis as { defineNuxtConfig?: (c: unknown) => unknown }
    g.defineNuxtConfig = (c) => c
    try {
      const config = (await import('../../layers/admin/nuxt.config'))
        .default as {
        components: { dirs: unknown[] }
      }
      expect(config.components.dirs).toEqual([
        { path: 'components', pathPrefix: false },
      ])
    } finally {
      delete g.defineNuxtConfig
    }
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

  it('keeps every plugin in the root layer', () => {
    // Plugins have the identical layer reversal as middleware, so a layer
    // plugin would run before `0.urql` installs the GraphQL client.
    expect(walk(join(FE, 'layers'), /\/plugins\//)).toEqual([])
  })
})
