// @vitest-environment nuxt
import { describe, it, expect, beforeEach, vi } from 'vitest'
import { ref } from 'vue'
import { mountSuspended, mockNuxtImport } from '@nuxt/test-utils/runtime'
import ArchivePage from '../../layers/user/app/pages/settings/archive.vue'
import DesignSkeleton from '../../layers/user/app/components/design/DesignSkeleton.vue'
import ErrorState from '../../layers/user/app/components/ErrorState.vue'
import EmptyState from '../../layers/user/app/components/EmptyState.vue'

const { queryMock, currentProjectMock, authReadyMock } = vi.hoisted(() => ({
  queryMock: vi.fn(),
  currentProjectMock: vi.fn(),
  authReadyMock: vi.fn(),
}))
mockNuxtImport('useProjectArchiveQuery', () => queryMock)
mockNuxtImport('useCurrentProjectQuery', () => currentProjectMock)
mockNuxtImport('useAuthReady', () => authReadyMock)

const achievement = (id: string) => ({
  __typename: 'SimpleAchievement' as const,
  id,
  name: `Achievement ${id}`,
  descriptionPending: '',
  descriptionCompleted: '',
  imagePendingObject: { url: 'pending.png' },
  imageCompletedObject: { url: 'completed.png' },
  hidden: false,
  achievedAt: null,
  celebratedAt: null,
  points: 10,
})

const project = (id: string, name: string, startDate: string, count = 1) => ({
  id,
  name,
  startDate,
  archivedAt: true,
  achievements: Array.from({ length: count }, (_, i) =>
    achievement(`${id}-AC${i}`),
  ),
})

function mountWith(
  state: { data?: unknown; error?: unknown; fetching?: boolean } = {},
  currentProjectId?: string,
) {
  authReadyMock.mockReturnValue({ isAuthReady: ref(true) })
  queryMock.mockReturnValue({
    data: ref(state.data ?? null),
    error: ref(state.error ?? null),
    fetching: ref(state.fetching ?? false),
  })
  currentProjectMock.mockReturnValue({
    data: ref(
      currentProjectId ? { myCurrentProject: { id: currentProjectId } } : null,
    ),
    error: ref(null),
    fetching: ref(false),
  })
  return mountSuspended(ArchivePage, {
    // The badge owns a teleporting drawer and its own mutation; the page's job
    // is only to hand it the right achievements, so assert on its props.
    global: { stubs: { AchievementBadge: true } },
  })
}

describe('archive page', () => {
  beforeEach(() => {
    vi.clearAllMocks()
  })

  it('pauses both queries until auth is ready', async () => {
    authReadyMock.mockReturnValue({ isAuthReady: ref(false) })
    queryMock.mockReturnValue({
      data: ref(null),
      error: ref(null),
      fetching: ref(false),
    })
    currentProjectMock.mockReturnValue({
      data: ref(null),
      error: ref(null),
      fetching: ref(false),
    })
    await mountSuspended(ArchivePage)

    const pause = queryMock.mock.calls[0]?.[0]?.pause as { value: boolean }
    expect(pause.value).toBe(true)
  })

  it('shows skeletons while loading', async () => {
    const wrapper = await mountWith({ fetching: true, data: null })

    expect(wrapper.findComponent(DesignSkeleton).exists()).toBe(true)
    expect(wrapper.findComponent(ErrorState).exists()).toBe(false)
  })

  it('shows the error state when the query fails', async () => {
    const wrapper = await mountWith({ error: new Error('boom') })

    expect(wrapper.findComponent(ErrorState).exists()).toBe(true)
  })

  it('shows an empty state when there is no history', async () => {
    const wrapper = await mountWith({
      data: { me: { id: 'US1', projects: [] } },
    })

    expect(wrapper.findComponent(EmptyState).exists()).toBe(true)
  })

  it('renders a project name and its achievement badges', async () => {
    const wrapper = await mountWith({
      data: {
        me: {
          id: 'US1',
          projects: [project('PR1', 'Sommercamp 2026', '2024-01-01', 3)],
        },
      },
    })

    expect(wrapper.text()).toContain('Sommercamp 2026')
    const badges = wrapper.findAllComponents({ name: 'AchievementBadge' })
    expect(badges).toHaveLength(3)
    expect(badges[0]?.props('achievement')).toMatchObject({ id: 'PR1-AC0' })
  })

  it('keeps projects that have no achievements', async () => {
    const wrapper = await mountWith({
      data: {
        me: {
          id: 'US1',
          projects: [
            project('PR1', 'Youth Winter Retreat 2025', '2024-01-01', 0),
          ],
        },
      },
    })

    expect(wrapper.text()).toContain('Youth Winter Retreat 2025')
    expect(wrapper.findComponent({ name: 'AchievementBadge' }).exists()).toBe(
      false,
    )
  })

  it('excludes the current project', async () => {
    const wrapper = await mountWith(
      {
        data: {
          me: {
            id: 'US1',
            projects: [
              project('PR_CURRENT', 'Ladder to Heaven', '2025-01-01'),
              project('PR_OLD', 'Sommercamp 2026', '2024-01-01'),
            ],
          },
        },
      },
      'PR_CURRENT',
    )

    expect(wrapper.text()).not.toContain('Ladder to Heaven')
    expect(wrapper.text()).toContain('Sommercamp 2026')
  })

  it('orders projects newest first regardless of server order', async () => {
    const wrapper = await mountWith({
      data: {
        me: {
          id: 'US1',
          projects: [
            project('PR_A', 'Oldest', '2022-06-15T12:00:00Z'),
            project('PR_C', 'Newest', '2024-06-15T12:00:00Z'),
            project('PR_B', 'Middle', '2023-06-15T12:00:00Z'),
          ],
        },
      },
    })

    const text = wrapper.text()
    expect(text.indexOf('Newest')).toBeLessThan(text.indexOf('Middle'))
    expect(text.indexOf('Middle')).toBeLessThan(text.indexOf('Oldest'))
  })
})
