// @vitest-environment nuxt
import { describe, it, expect, beforeEach, vi } from 'vitest'
import { ref, nextTick } from 'vue'
import { mountSuspended, mockNuxtImport } from '@nuxt/test-utils/runtime'
import ProfileProjectArchive from '../../layers/user/app/components/profile/ProfileProjectArchive.vue'
import DesignSkeleton from '../../layers/user/app/components/design/DesignSkeleton.vue'
import ErrorState from '../../layers/user/app/components/ErrorState.vue'

const { queryMock, authReadyMock } = vi.hoisted(() => ({
  queryMock: vi.fn(),
  authReadyMock: vi.fn(),
}))
mockNuxtImport('useProjectArchiveQuery', () => queryMock)
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

// Kept around so a test can push data in after mount, the way the real query
// does once the section is expanded and the request resolves.
let dataRef = ref<unknown>(null)

function mountWith(
  state: { data?: unknown; error?: unknown; fetching?: boolean } = {},
  props: { currentProjectId?: string } = {},
) {
  authReadyMock.mockReturnValue({ isAuthReady: ref(true) })
  dataRef = ref(state.data ?? null)
  queryMock.mockReturnValue({
    data: dataRef,
    error: ref(state.error ?? null),
    fetching: ref(state.fetching ?? false),
  })
  return mountSuspended(ProfileProjectArchive, {
    props,
    // The badge owns a teleporting drawer and its own mutation; the archive's
    // job is only to hand it the right achievements, so assert on its props.
    global: { stubs: { AchievementBadge: true } },
  })
}

// The query option is a ComputedRef, so read it after toggling the section.
function pauseState() {
  return queryMock.mock.calls[0]?.[0]?.pause as { value: boolean }
}

describe('profile project archive', () => {
  beforeEach(() => {
    vi.clearAllMocks()
  })

  it('starts collapsed, with the query paused', async () => {
    const wrapper = await mountWith({
      data: {
        me: { id: 'US1', projects: [project('PR1', 'Old', '2024-01-01')] },
      },
    })

    expect(wrapper.find('button').attributes('aria-expanded')).toBe('false')
    expect(pauseState().value).toBe(true)
    expect(wrapper.findComponent({ name: 'AchievementBadge' }).exists()).toBe(
      false,
    )
  })

  it('unpauses the query once expanded', async () => {
    const wrapper = await mountWith({
      data: {
        me: { id: 'US1', projects: [project('PR1', 'Old', '2024-01-01')] },
      },
    })

    await wrapper.find('button').trigger('click')

    expect(wrapper.find('button').attributes('aria-expanded')).toBe('true')
    expect(pauseState().value).toBe(false)
  })

  it('keeps the query paused while auth is not ready', async () => {
    authReadyMock.mockReturnValue({ isAuthReady: ref(false) })
    queryMock.mockReturnValue({
      data: ref(null),
      error: ref(null),
      fetching: ref(false),
    })
    const wrapper = await mountSuspended(ProfileProjectArchive)

    await wrapper.find('button').trigger('click')

    expect(pauseState().value).toBe(true)
  })

  it('shows skeletons while the archive loads', async () => {
    const wrapper = await mountWith({ fetching: true, data: null })

    await wrapper.find('button').trigger('click')

    expect(wrapper.findComponent(DesignSkeleton).exists()).toBe(true)
    expect(wrapper.findComponent(ErrorState).exists()).toBe(false)
  })

  it('shows the error state when the query fails', async () => {
    const wrapper = await mountWith({ error: new Error('boom') })

    await wrapper.find('button').trigger('click')

    expect(wrapper.findComponent(ErrorState).exists()).toBe(true)
  })

  it('shows an empty message when there are no earlier projects', async () => {
    const wrapper = await mountWith({
      data: { me: { id: 'US1', projects: [] } },
    })

    await wrapper.find('button').trigger('click')

    expect(wrapper.findComponent({ name: 'AchievementBadge' }).exists()).toBe(
      false,
    )
    expect(wrapper.text()).toContain('no earlier projects')
  })

  it('renders a project name and its achievement badges', async () => {
    const wrapper = await mountWith({
      data: {
        me: {
          id: 'US1',
          projects: [project('PR1', 'Sermon on the Mount', '2024-01-01', 3)],
        },
      },
    })

    await wrapper.find('button').trigger('click')

    expect(wrapper.text()).toContain('Sermon on the Mount')
    const badges = wrapper.findAllComponents({ name: 'AchievementBadge' })
    expect(badges).toHaveLength(3)
    expect(badges[0]?.props('achievement')).toMatchObject({ id: 'PR1-AC0' })
  })

  it('excludes the current project', async () => {
    const wrapper = await mountWith(
      {
        data: {
          me: {
            id: 'US1',
            projects: [
              project('PR_CURRENT', 'Ladder to Heaven', '2025-01-01'),
              project('PR_OLD', 'Sermon on the Mount', '2024-01-01'),
            ],
          },
        },
      },
      { currentProjectId: 'PR_CURRENT' },
    )

    await wrapper.find('button').trigger('click')

    expect(wrapper.text()).not.toContain('Ladder to Heaven')
    expect(wrapper.text()).toContain('Sermon on the Mount')
  })

  // The real sequence: expand with nothing cached, see skeletons, then the
  // response lands. This is also what triggers the entrance animation.
  it('swaps skeletons for projects when the response arrives', async () => {
    const wrapper = await mountWith({ data: null, fetching: true })

    await wrapper.find('button').trigger('click')
    expect(wrapper.findComponent(DesignSkeleton).exists()).toBe(true)

    dataRef.value = {
      me: {
        id: 'US1',
        projects: [project('PR1', 'Sommercamp 2026', '2024-01-01', 2)],
      },
    }
    await nextTick()
    await nextTick()

    expect(wrapper.findComponent(DesignSkeleton).exists()).toBe(false)
    expect(wrapper.text()).toContain('Sommercamp 2026')
    expect(
      wrapper.findAllComponents({ name: 'AchievementBadge' }),
    ).toHaveLength(2)
    // The stagger animation targets these.
    expect(wrapper.findAll('.archive-project')).toHaveLength(1)
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

    await wrapper.find('button').trigger('click')

    const text = wrapper.text()
    expect(text.indexOf('Newest')).toBeLessThan(text.indexOf('Middle'))
    expect(text.indexOf('Middle')).toBeLessThan(text.indexOf('Oldest'))
  })
})
