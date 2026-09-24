// @vitest-environment nuxt
import { describe, it, expect, vi } from 'vitest'
import { mountSuspended, mockNuxtImport } from '@nuxt/test-utils/runtime'
import { ref } from 'vue'
import { VueDraggable } from 'vue-draggable-plus'
import LeaderboardsPage from '../../layers/admin/app/pages/admin/projects/[projectId]/leaderboards/index.vue'
import { LeaderboardEntityType } from '../../app/api/generated'

const { updateConfig } = vi.hoisted(() => ({
  updateConfig: vi.fn().mockResolvedValue({}),
}))
const configs = [
  {
    id: 'LC1',
    name: 'Top five',
    entityType: LeaderboardEntityType.Persons,
    maxEntries: 5,
    sortOrder: 0,
    isActive: true,
    filter: null,
  },
  {
    id: 'LC2',
    name: 'All teams',
    entityType: LeaderboardEntityType.Teams,
    maxEntries: null,
    sortOrder: 1,
    isActive: true,
    filter: null,
  },
]
mockNuxtImport('useRoute', () => () => ({ params: { projectId: 'PR1' } }))
mockNuxtImport('usePermissions', () => () => ({ canEditProject: () => true }))
mockNuxtImport('useAuthReady', () => () => ({ isAuthReady: ref(true) }))
mockNuxtImport('useAdminProjectLeaderboardsQuery', () => () => ({
  data: ref({ project: { leaderboards: configs } }),
  error: ref(null),
  fetching: ref(false),
  executeQuery: vi.fn(),
}))
mockNuxtImport('useUpdateLeaderboardConfigMutation', () => () => ({
  executeMutation: updateConfig,
}))
mockNuxtImport('useToast', () => () => ({ add: vi.fn() }))

describe('admin leaderboards page', () => {
  it('preserves configured and unset entry limits when reordering', async () => {
    const wrapper = await mountSuspended(LeaderboardsPage)
    expect(wrapper.text()).toContain('Topp 5')
    const draggable = wrapper.findComponent(VueDraggable)
    await draggable.vm.$emit('update:modelValue', [...configs].reverse())
    await draggable.vm.$emit('end')
    expect(updateConfig).toHaveBeenCalledWith({
      id: 'LC1',
      input: expect.objectContaining({ maxEntries: 5, sortOrder: 1 }),
    })
    expect(updateConfig).toHaveBeenCalledWith({
      id: 'LC2',
      input: expect.objectContaining({ maxEntries: null, sortOrder: 0 }),
    })
  })
})
