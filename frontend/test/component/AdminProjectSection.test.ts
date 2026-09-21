// @vitest-environment nuxt
import { describe, it, expect, beforeEach, afterEach, vi } from 'vitest'
import { mountSuspended, mockNuxtImport } from '@nuxt/test-utils/runtime'
import { ref } from 'vue'
import type { usePermissions } from '../../app/composables/usePermissions'
import AdminProjectSection from '../../layers/admin/app/components/admin/project/AdminProjectSection.vue'

type Permissions = ReturnType<typeof usePermissions>

const trendDays = [
  { date: '2026-09-15', points: 0, activeUsers: 0 },
  { date: '2026-09-16', points: 120, activeUsers: 12 },
  { date: '2026-09-17', points: 0, activeUsers: 0 },
]

const counts = {
  project: { id: 'PR01ARZ3NDEKTSV4RRFFQ69G5FAV', activityTrend: trendDays },
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

    const nav = wrapper.find('nav').text()
    expect(nav).toContain('Utfordringer')
    expect(nav).toContain('12')
    expect(nav).toContain('Utmerkelser')
    expect(nav).toContain('8')
    expect(nav).toContain('Lag')
    expect(nav).toContain('24')
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

    // Scoped to the nav on purpose: the page text also contains "Poeng siste
    // 14 dager" from the trend tile, which is not a shortcut.
    const nav = wrapper.find('nav').text()
    expect(nav).toContain('Utfordringer')
    expect(nav).not.toContain('Lag')
    expect(nav).not.toContain('Poeng')
    expect(nav).not.toContain('Innstillinger')
  })

  it('does not repeat the overview as a shortcut', async () => {
    const wrapper = await mountSuspended(AdminProjectSection, {
      props: { project: project() },
    })

    // The title already links to the project root.
    expect(wrapper.find('nav').text()).not.toContain('Oversikt')
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

  it('renders a stat tile per trend measure', async () => {
    const wrapper = await mountSuspended(AdminProjectSection, {
      props: { project: project() },
    })

    const text = wrapper.text()
    expect(text).toContain('Poeng siste 14 dager')
    expect(text).toContain('Aktive deltakere per dag')
  })

  // Points are additive; a distinct-users-per-day count is not. Summing the
  // second would count the same person once per day they appeared.
  it('sums points but averages active users', async () => {
    const wrapper = await mountSuspended(AdminProjectSection, {
      props: { project: project() },
    })

    const tiles = wrapper.findAll('.grid > div')
    const pointsTile = tiles.find((tile) =>
      tile.text().includes('Poeng siste 14 dager'),
    )
    const usersTile = tiles.find((tile) =>
      tile.text().includes('Aktive deltakere per dag'),
    )

    expect(pointsTile?.text()).toContain('120')
    // (0 + 12 + 0) / 3 = 4, not 12.
    expect(usersTile?.text()).toContain('4')
  })

  // Two measures of very different scale must never share one y-axis, so they
  // are two charts. Each sparkline is its own svg.
  it('plots each measure in its own chart', async () => {
    const wrapper = await mountSuspended(AdminProjectSection, {
      props: { project: project() },
    })

    expect(wrapper.findAll('svg').length).toBeGreaterThanOrEqual(2)
  })

  // The tooltip must not be the only way to reach a value.
  it('exposes the daily values as a table, not only on hover', async () => {
    const wrapper = await mountSuspended(AdminProjectSection, {
      props: { project: project() },
    })

    const tables = wrapper.findAll('.sr-only table')
    expect(tables.length).toBeGreaterThanOrEqual(2)
    expect(tables[0]!.text()).toContain('120')
  })

  // An upcoming project's trend is 14 empty days, which reads as a broken
  // chart; the field is skipped for it entirely.
  it('omits the trend when the project has not started', async () => {
    data.value = { ...counts, project: { id: 'PR1', activityTrend: undefined } }

    const wrapper = await mountSuspended(AdminProjectSection, {
      props: {
        project: project({
          startDate: '2026-10-01T10:00:00.000Z',
          endDate: '2026-10-10T10:00:00.000Z',
        }),
      },
    })

    expect(wrapper.text()).not.toContain('Poeng siste 14 dager')
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
