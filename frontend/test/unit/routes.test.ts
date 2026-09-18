import { describe, it, expect } from 'vitest'
import { buildTree, toVueRouter4 } from 'unrouting'
import { readdirSync, statSync, readFileSync } from 'node:fs'
import { join, resolve, relative } from 'node:path'

/**
 * Route manifest guards.
 *
 * Nuxt 4 generates its routes through `unrouting` (see `buildTree` /
 * `toVueRouter4` in `nuxt/dist/index.mjs`), so building the tree here with the
 * same options reproduces the real manifest without spinning up Nuxt. That
 * makes route structure reviewable in a diff and lets us assert two invariants
 * that nothing else in the toolchain catches:
 *
 *  - a parent route whose children include none at path '' renders an empty
 *    `<NuxtPage />` — a blank page with no 404, no error and no type error;
 *  - a `name:` binding pointing at a route that no longer exists fails only at
 *    runtime, as a vue-router warning and a dead link.
 */

const APP = resolve(__dirname, '../../app')
const PAGES = join(APP, 'pages')

function walk(dir: string, match: RegExp): string[] {
  return readdirSync(dir).flatMap((entry) => {
    const path = join(dir, entry)
    return statSync(path).isDirectory()
      ? walk(path, match)
      : match.test(path)
        ? [path]
        : []
  })
}

interface Route {
  path: string
  name?: string
  file?: string
  children?: Route[]
}

function buildManifest(): Route[] {
  const files = walk(PAGES, /\.vue$/)
    .sort()
    .map((path) => ({ path, priority: 0 }))
  // Mirrors createPagesContext() in nuxt/dist/index.mjs.
  const tree = buildTree(files, { roots: [PAGES], modes: ['client'] })
  return toVueRouter4(tree, { attrs: { mode: ['client'] } }) as Route[]
}

function flatten(
  routes: Route[],
  parent = '',
): Array<{ name?: string; path: string; file?: string; children: number }> {
  return routes.flatMap((route) => {
    const full = route.path.startsWith('/')
      ? route.path
      : `${parent}/${route.path}`.replace(/\/+/g, '/')
    return [
      {
        name: route.name,
        path: full,
        file: route.file ? relative(APP, route.file) : undefined,
        children: route.children?.length ?? 0,
      },
      ...flatten(route.children ?? [], full),
    ]
  })
}

describe('route manifest', () => {
  const manifest = flatten(buildManifest())

  it('matches the committed snapshot', () => {
    const rows = manifest
      .map((r) => `${r.name ?? '(unnamed)'}  ${r.path}  ${r.file ?? '-'}`)
      .sort()
    expect(rows).toMatchSnapshot()
  })

  // Guards the blank-page failure mode. A parent page (e.g. `[projectId].vue`)
  // renders `<NuxtPage />`; if its directory loses its `index.vue`, the route
  // still resolves and the content area silently renders nothing.
  it('every parent route has a child at path ""', () => {
    const parents = buildManifest()
    const offenders: string[] = []

    const check = (routes: Route[]) => {
      for (const route of routes) {
        if (route.children?.length) {
          if (!route.children.some((c) => c.path === '')) {
            offenders.push(`${route.name ?? route.path} (${route.file})`)
          }
          check(route.children)
        }
      }
    }
    check(parents)

    expect(offenders).toEqual([])
  })

  it('every route name referenced in the app resolves to a real route', () => {
    const names = new Set(manifest.flatMap((r) => (r.name ? [r.name] : [])))

    const dangling = new Map<string, string[]>()
    for (const file of walk(APP, /\.(vue|ts)$/)) {
      if (file.endsWith('api/generated.ts')) continue
      const text = readFileSync(file, 'utf8')
      const referenced = [
        // `:to="{ name: 'admin-projects' }"`, `navigateTo({ name: ... })`
        ...text.matchAll(/name:\s*'([a-z][a-z0-9]*(?:-[a-z0-9]+)+)'/g),
        // `useRoute('admin-projects-projectId')`
        ...text.matchAll(/useRoute\(\s*'([^']+)'\s*\)/g),
      ].map((m) => m[1]!)

      for (const name of referenced) {
        if (!names.has(name)) {
          const at = relative(APP, file)
          dangling.set(name, [...(dangling.get(name) ?? []), at])
        }
      }
    }

    expect(Object.fromEntries(dangling)).toEqual({})
  })
})
