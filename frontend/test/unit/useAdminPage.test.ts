import { describe, it, expect, vi, beforeEach, afterEach } from 'vitest'
import { computed, ref, watchEffect, toRef, onScopeDispose, type Ref } from 'vue'

/**
 * Breadcrumb derivation.
 *
 * The structure comes from the route plus the nav model, so it is worth pinning
 * exactly: these are the shapes that ~25 hand-written breadcrumb blocks used to
 * express by copy-paste, and which drifted from each other.
 */

let routeName: Ref<string>
let projectId: Ref<string | undefined>
let projectName: Ref<string | undefined>

function stubGlobals() {
  vi.stubGlobal('ref', ref)
  vi.stubGlobal('computed', computed)
  vi.stubGlobal('toRef', toRef)
  vi.stubGlobal('watchEffect', watchEffect)
  vi.stubGlobal('onScopeDispose', onScopeDispose)
  vi.stubGlobal('useRoute', () => ({
    get name() {
      return routeName.value
    },
  }))
  vi.stubGlobal('useCurrentProject', () => ({
    projectId,
    project: computed(() =>
      projectId.value ? { id: projectId.value, name: projectName.value } : null,
    ),
  }))
}

async function load() {
  stubGlobals()
  vi.resetModules()
  const mod = await import('../../app/composables/useAdminPage')
  return mod.useAdminPage
}

/** `label → to` pairs, so both the text and the linking are asserted. */
function crumbs(items: { label?: string; to?: unknown }[]) {
  return items.map((i) => [i.label, i.to ? 'link' : 'current'])
}

beforeEach(() => {
  routeName = ref('admin')
  projectId = ref(undefined)
  projectName = ref(undefined)
})

afterEach(() => {
  vi.unstubAllGlobals()
})

describe('admin breadcrumb', () => {
  it('is empty on a top-level destination', async () => {
    const useAdminPage = await load()
    routeName.value = 'admin'

    // "Hjem" is an exact-match entry: a destination, not an ancestor.
    expect(useAdminPage().breadcrumb.value).toEqual([])
  })

  it('names the section on a global list', async () => {
    const useAdminPage = await load()
    routeName.value = 'admin-users'

    expect(crumbs(useAdminPage().breadcrumb.value)).toEqual([
      ['Brukere', 'current'],
    ])
  })

  it('builds project → section for a project subsection', async () => {
    const useAdminPage = await load()
    routeName.value = 'admin-projects-projectId-challenges'
    projectId.value = 'PR01'
    projectName.value = 'Sommerleir'

    expect(crumbs(useAdminPage().breadcrumb.value)).toEqual([
      ['Prosjekter', 'link'],
      ['Sommerleir', 'link'],
      ['Utfordringer', 'current'],
    ])
  })

  it('appends the page label a page supplies', async () => {
    const useAdminPage = await load()
    routeName.value = 'admin-projects-projectId-challenges-challengeId'
    projectId.value = 'PR01'
    projectName.value = 'Sommerleir'

    useAdminPage(() => 'Bibelquiz')

    expect(crumbs(useAdminPage().breadcrumb.value)).toEqual([
      ['Prosjekter', 'link'],
      ['Sommerleir', 'link'],
      ['Utfordringer', 'link'],
      ['Bibelquiz', 'current'],
    ])
  })

  // The overview *is* the project crumb; listing "Oversikt" after it would say
  // the same thing twice.
  it('does not repeat the project as its own overview section', async () => {
    const useAdminPage = await load()
    routeName.value = 'admin-projects-projectId'
    projectId.value = 'PR01'
    projectName.value = 'Sommerleir'

    expect(crumbs(useAdminPage().breadcrumb.value)).toEqual([
      ['Prosjekter', 'link'],
      ['Sommerleir', 'current'],
    ])
  })

  it('falls back to the id before the project name loads', async () => {
    const useAdminPage = await load()
    routeName.value = 'admin-projects-projectId'
    projectId.value = 'PR01'
    projectName.value = undefined

    expect(useAdminPage().breadcrumb.value[1]?.label).toBe('PR01')
  })

  it('never links the last crumb', async () => {
    const useAdminPage = await load()
    routeName.value = 'admin-projects-projectId-teams'
    projectId.value = 'PR01'

    const items = useAdminPage().breadcrumb.value
    expect(items.at(-1)?.to).toBeUndefined()
    expect(items.at(0)?.to).toBeDefined()
  })
})

describe('admin navbar title', () => {
  it('uses the page label when one is set', async () => {
    const useAdminPage = await load()
    routeName.value = 'admin-projects-projectId-challenges-challengeId'
    projectId.value = 'PR01'

    useAdminPage(() => 'Bibelquiz')

    expect(useAdminPage().title.value).toBe('Bibelquiz')
  })

  it('falls back to the section label', async () => {
    const useAdminPage = await load()
    routeName.value = 'admin-projects-projectId-challenges'
    projectId.value = 'PR01'

    expect(useAdminPage().title.value).toBe('Utfordringer')
  })

  it('prefers the project section over the global one', async () => {
    const useAdminPage = await load()
    routeName.value = 'admin-projects-projectId-teams'
    projectId.value = 'PR01'

    // Both "Prosjekter" (global) and "Lag" (project) match this route.
    expect(useAdminPage().title.value).toBe('Lag')
  })
})
