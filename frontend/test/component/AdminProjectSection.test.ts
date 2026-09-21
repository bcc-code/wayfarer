// @vitest-environment nuxt
import { describe, it, expect, beforeEach, afterEach, vi } from 'vitest'
import { mountSuspended, mockNuxtImport } from '@nuxt/test-utils/runtime'
import { ref } from 'vue'
import type { usePermissions } from '../../app/composables/usePermissions'
import AdminProjectSection from '../../layers/admin/app/components/admin/project/AdminProjectSection.vue'

type Permissions = ReturnType<typeof usePermissions>

const counts = {
  challenges: { totalCount: 12 },
  achievements: { totalCount: 8 },
  events: { totalCount: 0 },
  superteams: { totalCount: 3 },
  teams: { totalCount: 24 },
  users: { totalCount: 1234 },
}

const fetching = ref(false)
const data = ref<typeof counts | undefined>(counts)

mockNuxtImport('useAdminProjectSectionCountsQuery', () => () => ({
  data,
  fetching,
  error: ref(undefined),
}))

mockNuxtImport('useAuthReady', () => () => ({ isAuthReady: ref(true) }))

// Only the three members PROJECT_NAV's `can` predicates actually read.
const canAccessTeams = ref(true)
const canAccessScores = ref(true)
const canEditProject = ref(true)

mockNuxtImport(
  'usePermissions',
  () => () =>
    ({
      canAccessTeams,
      canAccessScores,
      canEditProject: () => canEditProject.value,
    }) as unknown as Permissions,
)

const project = (
  overrides: Partial<{ startDate: string; endDate: string; name: string }> = {},
) => ({
  id: 'PR01ARZ3NDEKTSV4RRFFQ69G5FAV',
  name: 'Spring Revival',
  startDate: '2026-09-01T10:00:00.000Z',
  endDate: '2026-10-01T10:00:00.000Z',
  branding: {
    logoImage: null,
    colors: { light: { accent: '#00ff00' }, dark: { accent: '#00aa00' } },
  },
  ...overrides,
})

describe('AdminProjectSection', () => {
  beforeEach(() => {
    data.value = counts
    fetching.value = false
    canAccessTeams.value = true
    canAccessScores.value = true
    canEditProject.value = true
    vi.useFakeTimers()
    // Noon, per the repo convention — midnight lands on timezone boundaries.
    vi.setSystemTime(new Date('2026-09-16T12:00:00.000Z'))
  })

  afterEach(() => {
    vi.useRealTimers()
  })

  it('renders the project name and countdown', async () => {
    const wrapper = await mountSuspended(AdminProjectSection, {
      props: { project: project() },
    })

    expect(wrapper.text()).toContain('Spring Revival')
    expect(wrapper.text()).toContain('dager igjen')
  })

  it('shows the participant count', async () => {
    const wrapper = await mountSuspended(AdminProjectSection, {
      props: { project: project() },
    })

    expect(wrapper.text()).toContain('deltakere')
  })

  // The shortcuts come from PROJECT_NAV so they cannot drift from the sidebar.
  it('renders a shortcut per project nav entry, with its count', async () => {
    const wrapper = await mountSuspended(AdminProjectSection, {
      props: { project: project() },
    })

    const text = wrapper.text()
    expect(text).toContain('Utfordringer')
    expect(text).toContain('12')
    expect(text).toContain('Utmerkelser')
    expect(text).toContain('8')
    expect(text).toContain('Lag')
    expect(text).toContain('24')
  })

  // Gating is PROJECT_NAV's, shared with the sidebar — so a role that cannot
  // reach a section must not be offered a shortcut to it either.
  it('omits shortcuts the user has no permission for', async () => {
    canAccessTeams.value = false
    canAccessScores.value = false
    canEditProject.value = false

    const wrapper = await mountSuspended(AdminProjectSection, {
      props: { project: project() },
    })

    const text = wrapper.text()
    expect(text).toContain('Utfordringer')
    expect(text).not.toContain('Lag')
    expect(text).not.toContain('Poeng')
    expect(text).not.toContain('Innstillinger')
  })

  it('does not repeat the overview as a shortcut', async () => {
    const wrapper = await mountSuspended(AdminProjectSection, {
      props: { project: project() },
    })

    // The title already links to the project root.
    expect(wrapper.text()).not.toContain('Oversikt')
  })

  it('points every shortcut at this project', async () => {
    const wrapper = await mountSuspended(AdminProjectSection, {
      props: { project: project() },
    })

    const hrefs = wrapper
      .findAll('nav a')
      .map((link) => link.attributes('href'))
    expect(hrefs.length).toBeGreaterThan(0)
    for (const href of hrefs) {
      expect(href).toContain('PR01ARZ3NDEKTSV4RRFFQ69G5FAV')
    }
  })

  // A zero count is ordinary — challenges and events are episodic — so it must
  // render as a number, not be hidden or flagged.
  it('renders a zero count plainly', async () => {
    const wrapper = await mountSuspended(AdminProjectSection, {
      props: { project: project() },
    })

    const events = wrapper
      .findAll('nav a')
      .find((link) => link.text().includes('Arrangement'))
    expect(events?.text()).toContain('0')
    expect(events?.html()).not.toContain('text-warning')
  })

  it('holds the layout with skeletons while counts load', async () => {
    data.value = undefined
    fetching.value = true

    const wrapper = await mountSuspended(AdminProjectSection, {
      props: { project: project() },
    })

    expect(wrapper.text()).toContain('Utfordringer')
    expect(wrapper.find('nav').html()).toContain('animate-pulse')
  })
})
