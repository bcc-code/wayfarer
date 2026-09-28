import { describe, it, expect, vi, beforeEach, afterEach } from 'vitest'
import { computed, ref, type Ref } from 'vue'

/**
 * Route *locations*, not just visibility — `adminNav.test.ts` covers the model
 * itself. What is worth pinning here is which entries carry `projectId`:
 * handing the param to a global route that does not declare it makes
 * vue-router discard it and log "Discarded invalid param(s)" for every link on
 * every render, which is how the sidebar filled the console inside a project.
 */

let routeName: Ref<string>
let params: Ref<Record<string, string>>

function stubGlobals() {
  vi.stubGlobal('ref', ref)
  vi.stubGlobal('computed', computed)
  vi.stubGlobal('useRoute', () => ({
    get name() {
      return routeName.value
    },
    get params() {
      return params.value
    },
  }))
  // Everything permitted: gating is adminNav.test.ts's subject, not this file's.
  vi.stubGlobal('usePermissions', () => {
    const allowed = computed(() => true)
    return new Proxy(
      {},
      {
        get: (_target, key) =>
          key === 'canEditProject' ? () => true : allowed,
      },
    )
  })
}

async function load() {
  stubGlobals()
  vi.resetModules()
  const mod = await import('../../layers/admin/app/composables/useAdminNav')
  return mod.useAdminNav
}

beforeEach(() => {
  routeName = ref('admin')
  params = ref({})
})

afterEach(() => {
  vi.unstubAllGlobals()
})

describe('useAdminNav', () => {
  it('links global entries without params outside a project', async () => {
    const useAdminNav = await load()
    const { globalNav, projectNav } = useAdminNav()

    expect(globalNav.value.map((i) => i.to)).toEqual([
      { name: 'admin' },
      { name: 'admin-projects' },
      { name: 'admin-users' },
      { name: 'admin-consents' },
      { name: 'admin-feedback' },
      { name: 'admin-maintenance' },
    ])
    expect(projectNav.value).toEqual([])
  })

  it('keeps global entries param-free while inside a project', async () => {
    routeName = ref('admin-projects-projectId-teams')
    params = ref({ projectId: 'PR01' })

    const useAdminNav = await load()
    const { globalNav } = useAdminNav()

    expect(
      globalNav.value.every((item) => !('params' in (item.to as object))),
    ).toBe(true)
  })

  it('scopes project entries to the current project', async () => {
    routeName = ref('admin-projects-projectId-teams')
    params = ref({ projectId: 'PR01' })

    const useAdminNav = await load()
    const { projectNav } = useAdminNav()

    expect(projectNav.value.length).toBeGreaterThan(0)
    expect(projectNav.value.every((item) => item.to)).toBe(true)
    for (const item of projectNav.value) {
      expect(item.to).toMatchObject({ params: { projectId: 'PR01' } })
    }
  })

  it('marks the entry matching the current route active', async () => {
    routeName = ref('admin-projects-projectId-teams')
    params = ref({ projectId: 'PR01' })

    const useAdminNav = await load()
    const { globalNav, projectNav } = useAdminNav()

    expect(
      projectNav.value.filter((i) => i.active).map((i) => i.label),
    ).toEqual(['Lag'])
    // 'Prosjekter' is a prefix entry, so it stays lit for anything nested
    // under /admin/projects — including this project-scoped route.
    expect(globalNav.value.filter((i) => i.active).map((i) => i.label)).toEqual(
      ['Prosjekter'],
    )
  })

  it('offers the same targets in the command palette as in the sidebar', async () => {
    routeName = ref('admin-projects-projectId')
    params = ref({ projectId: 'PR01' })

    const useAdminNav = await load()
    const { globalNav, projectNav, searchGroups } = useAdminNav()

    expect(searchGroups.value.map((g) => g.id)).toEqual(['global', 'project'])
    expect(searchGroups.value[0]!.items!.map((i) => i.to)).toEqual(
      globalNav.value.map((i) => i.to),
    )
    expect(searchGroups.value[1]!.items!.map((i) => i.to)).toEqual(
      projectNav.value.map((i) => i.to),
    )
  })
})
