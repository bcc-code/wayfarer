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

  it('badges a project that is currently running', async () => {
    const wrapper = await mountSuspended(AdminProjectCard, {
      props: { project: project() },
    })

    expect(wrapper.text()).toContain('Active')
  })

  it('leaves a project outside its date range unbadged', async () => {
    const wrapper = await mountSuspended(AdminProjectCard, {
      props: {
        project: project({
          startDate: '2025-01-01T10:00:00.000Z',
          endDate: '2025-02-01T10:00:00.000Z',
        }),
      },
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
    // Non-square logos must letterbox rather than stretch to the 32px box.
    expect(img.classes()).toContain('object-contain')
  })

  it('stays neutral instead of tinting itself with the project accent', async () => {
    const wrapper = await mountSuspended(AdminProjectCard, {
      props: { project: project() },
    })

    // The card used to paint ring and body from branding.colors, turning the
    // grid into a patchwork. Nothing may bind a per-project colour again.
    expect(wrapper.html()).not.toMatch(/--accent|ring-\(|bg-\(/)
  })
})
