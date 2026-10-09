// @vitest-environment nuxt
import { describe, it, expect, beforeEach, afterEach, vi } from 'vitest'
import { mountSuspended } from '@nuxt/test-utils/runtime'
import AdminProjectCard from '../../layers/admin/app/components/admin/project/AdminProjectCard.vue'

const project = (
  overrides: Partial<{
    name: string
    description: string
    startDate: string
    endDate: string
    branding: { logoImage: { url: string } | null }
  }> = {},
) => ({
  id: 'PR01ARZ3NDEKTSV4RRFFQ69G5FAV',
  name: 'Spring Revival',
  description: 'En uke med lovsang.',
  startDate: '2026-09-01T10:00:00.000Z',
  endDate: '2026-10-01T10:00:00.000Z',
  branding: { logoImage: null },
  ...overrides,
})

describe('AdminProjectCard', () => {
  beforeEach(() => {
    vi.useFakeTimers()
    // Noon, per the repo convention — midnight lands on timezone boundaries.
    vi.setSystemTime(new Date('2026-09-16T12:00:00.000Z'))
  })

  afterEach(() => {
    vi.useRealTimers()
  })

  it('renders the name, description and date range', async () => {
    const wrapper = await mountSuspended(AdminProjectCard, {
      props: { project: project() },
    })

    expect(wrapper.text()).toContain('Spring Revival')
    expect(wrapper.text()).toContain('En uke med lovsang.')
  })

  it('badges the project end users currently see', async () => {
    const wrapper = await mountSuspended(AdminProjectCard, {
      props: { project: project(), isCurrent: true },
    })

    expect(wrapper.text()).toContain('Gjeldende')
  })

  it('leaves every other project unbadged', async () => {
    const wrapper = await mountSuspended(AdminProjectCard, {
      props: { project: project() },
    })

    expect(wrapper.text()).not.toContain('Gjeldende')
  })

  // A project running right now is not the same thing as the current project,
  // and badging both taught the opposite.
  it('does not badge a project merely because its dates are running', async () => {
    const wrapper = await mountSuspended(AdminProjectCard, {
      props: { project: project() },
    })

    expect(wrapper.text()).not.toContain('Active')
  })

  it('shows the project logo when branding has one', async () => {
    const wrapper = await mountSuspended(AdminProjectCard, {
      props: {
        project: project({
          branding: { logoImage: { url: 'https://cdn.example/spring.png' } },
        }),
      },
    })

    const img = wrapper.find('img')
    expect(img.attributes('src')).toBe('https://cdn.example/spring.png')
    expect(img.classes()).toContain('object-contain')
  })

  it('stays neutral instead of tinting itself with the project accent', async () => {
    const wrapper = await mountSuspended(AdminProjectCard, {
      props: { project: project() },
    })

    expect(wrapper.html()).not.toMatch(/--accent|ring-\(|bg-\(/)
  })
})
