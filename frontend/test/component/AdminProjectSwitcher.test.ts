// @vitest-environment nuxt
import { describe, it, expect, beforeEach, vi } from 'vitest'
import { mountSuspended, mockNuxtImport } from '@nuxt/test-utils/runtime'
import { reactive, ref } from 'vue'
import type { DropdownMenuItem } from '@nuxt/ui'
import type { usePermissions } from '../../app/composables/usePermissions'
import AdminProjectSwitcher from '../../layers/admin/app/components/admin/AdminProjectSwitcher.vue'

type Permissions = ReturnType<typeof usePermissions>
type ProjectItem = DropdownMenuItem & { projectId?: string }

const BRANDED = 'PR01ARZ3NDEKTSV4RRFFQ69G5FAV'
const PLAIN = 'PR02ARZ3NDEKTSV4RRFFQ69G5FAV'

const node = (
  id: string,
  name: string,
  logoUrl: string | null,
  startDate: string,
  endDate: string,
) => ({
  node: {
    id,
    name,
    startDate,
    endDate,
    branding: { logoImage: logoUrl ? { url: logoUrl } : null },
  },
})

const data = ref({
  projects: {
    edges: [
      node(
        BRANDED,
        'Spring Revival',
        'https://cdn.example/spring.png',
        '2026-09-01T10:00:00.000Z',
        '2026-10-01T10:00:00.000Z',
      ),
      node(
        PLAIN,
        'Winter Camp',
        null,
        '2026-12-01T10:00:00.000Z',
        '2026-12-20T10:00:00.000Z',
      ),
    ],
  },
})

mockNuxtImport('useAdminProjectSwitcherQuery', () => () => ({
  data,
  fetching: ref(false),
  error: ref(undefined),
}))

mockNuxtImport('useAuthReady', () => () => ({ isAuthReady: ref(true) }))

mockNuxtImport(
  'usePermissions',
  () => () => ({ canViewProject: () => true }) as unknown as Permissions,
)

const route = reactive({
  name: 'admin-projects-projectId',
  params: {} as Record<string, string>,
})

mockNuxtImport('useRoute', () => () => route)

/** Flattens the grouped menu down to the rows that stand for a project. */
function projectItems(wrapper: { vm: unknown }): ProjectItem[] {
  const items = (wrapper.vm as { items: ProjectItem[][] }).items
  return items.flat().filter((item) => item.projectId)
}

describe('AdminProjectSwitcher', () => {
  beforeEach(() => {
    route.params = { projectId: BRANDED }
    vi.useFakeTimers()
    // Noon, per the repo convention — midnight lands on timezone boundaries.
    vi.setSystemTime(new Date('2026-09-16T12:00:00.000Z'))
  })

  it('gives each project its logo as the leading avatar', async () => {
    const wrapper = await mountSuspended(AdminProjectSwitcher)

    const branded = projectItems(wrapper).find((i) => i.projectId === BRANDED)
    expect(branded?.avatar).toMatchObject({
      src: 'https://cdn.example/spring.png',
      alt: 'Spring Revival',
    })
  })

  it('falls back to the layers icon when a project has no logo', async () => {
    const wrapper = await mountSuspended(AdminProjectSwitcher)

    const plain = projectItems(wrapper).find((i) => i.projectId === PLAIN)
    expect(plain?.avatar).toMatchObject({
      src: undefined,
      icon: 'lucide:layers',
    })
  })

  it('keeps the logo on the rows rather than replacing it with a check', async () => {
    const wrapper = await mountSuspended(AdminProjectSwitcher)

    // The active row used to carry `icon: 'lucide:check'`, which wins over
    // `avatar` in Nuxt UI's item template — the check now trails instead.
    for (const item of projectItems(wrapper)) {
      expect(item.icon).toBeUndefined()
    }
  })

  it("shows the active project's logo on the trigger", async () => {
    const wrapper = await mountSuspended(AdminProjectSwitcher)

    const button = wrapper.findComponent({ name: 'UButton' })
    expect(button.props('avatar')).toMatchObject({
      src: 'https://cdn.example/spring.png',
    })
    expect(button.props('icon')).toBeUndefined()
  })

  it('falls back to the layers icon on the trigger with no project selected', async () => {
    route.params = {}
    const wrapper = await mountSuspended(AdminProjectSwitcher)

    const button = wrapper.findComponent({ name: 'UButton' })
    expect(button.props('icon')).toBe('lucide:layers')
    expect(button.props('avatar')).toBeUndefined()
  })
})
