// @vitest-environment nuxt
import { describe, it, expect, beforeEach, vi } from 'vitest'
import { ref, nextTick } from 'vue'
import { mountSuspended, mockNuxtImport } from '@nuxt/test-utils/runtime'
import AchievementBadge from '~/components/achievements/AchievementBadge.vue'
import DesignImage from '~/components/design/DesignImage.vue'

const { analyticsMock, sheetMock, celebratedMutationMock } = vi.hoisted(() => ({
  analyticsMock: vi.fn(),
  sheetMock: vi.fn(),
  celebratedMutationMock: vi.fn(),
}))
mockNuxtImport('useAnalytics', () => analyticsMock)
mockNuxtImport('useAchievementSheet', () => sheetMock)
mockNuxtImport(
  'useMarkAchievementCelebratedMutation',
  () => celebratedMutationMock,
)

const track = vi.fn()
const markCelebrated = vi.fn()
const clearOpenAchievementId = vi.fn()

type Achievement = InstanceType<typeof AchievementBadge>['achievement']

function makeAchievement(overrides: Partial<Achievement> = {}): Achievement {
  return {
    id: 'AC01ARZ3NDEKTSV4RRFFQ69G5FAV',
    name: 'First Steps',
    achievedAt: null,
    celebratedAt: null,
    points: 0,
    descriptionCompleted: 'You did it!',
    descriptionPending: 'Not yet unlocked',
    imageCompletedObject: { url: 'https://img/completed.png' },
    imagePendingObject: { url: 'https://img/pending.png' },
    ...overrides,
  } as Achievement
}

// A stub that renders both the default slot (badge trigger) and the #content
// slot (the sheet body), and supports v-model:open so we can drive open/close.
const DesignDrawerStub = {
  props: ['open', 'title'],
  emits: ['update:open'],
  template: '<div><slot /><slot name="content" /></div>',
}

function mountWith(
  achievement: Achievement,
  opts: { openId?: string | null; celebrating?: boolean } = {},
) {
  analyticsMock.mockReturnValue({ track })
  sheetMock.mockReturnValue({
    openAchievementId: ref(opts.openId ?? null),
    clearOpenAchievementId,
    celebrating: ref(opts.celebrating ?? false),
  })
  celebratedMutationMock.mockReturnValue({ executeMutation: markCelebrated })
  return mountSuspended(AchievementBadge, {
    props: { achievement },
    global: {
      // DesignImage renders for real; only DesignDrawer is stubbed because it
      // teleports its content (via @nuxt/ui UDrawer) and we drive open/close.
      stubs: { DesignDrawer: DesignDrawerStub },
    },
  })
}

describe('AchievementBadge', () => {
  beforeEach(() => {
    vi.clearAllMocks()
  })

  describe('currentImage', () => {
    it('uses the completed image when achieved and a completed url exists', async () => {
      const wrapper = await mountWith(
        makeAchievement({ achievedAt: '2026-01-01T00:00:00Z' }),
      )

      const image = wrapper.findComponent(DesignImage)
      expect(image.props('image')).toMatchObject({
        url: 'https://img/completed.png',
      })
    })

    it('uses the pending image when not achieved', async () => {
      const wrapper = await mountWith(makeAchievement({ achievedAt: null }))

      const image = wrapper.findComponent(DesignImage)
      expect(image.props('image')).toMatchObject({
        url: 'https://img/pending.png',
      })
    })

    it('falls back to the pending image when the completed image has no url', async () => {
      const wrapper = await mountWith(
        makeAchievement({
          achievedAt: '2026-01-01T00:00:00Z',
          imageCompletedObject: { url: '' },
        }),
      )

      const image = wrapper.findComponent(DesignImage)
      expect(image.props('image')).toMatchObject({
        url: 'https://img/pending.png',
      })
    })
  })

  describe('description', () => {
    it('shows the completed description when achieved', async () => {
      const wrapper = await mountWith(
        makeAchievement({ achievedAt: '2026-01-01T00:00:00Z' }),
      )

      expect(wrapper.html()).toContain('You did it!')
      expect(wrapper.html()).not.toContain('Not yet unlocked')
    })

    it('shows the pending description when not achieved', async () => {
      const wrapper = await mountWith(makeAchievement({ achievedAt: null }))

      expect(wrapper.html()).toContain('Not yet unlocked')
      expect(wrapper.html()).not.toContain('You did it!')
    })
  })

  describe('points display', () => {
    it('shows earned points when achieved', async () => {
      const wrapper = await mountWith(
        makeAchievement({ achievedAt: '2026-01-01T00:00:00Z', points: 50 }),
      )

      const earned = wrapper.find('.text-accent-contrast')
      expect(earned.exists()).toBe(true)
      expect(earned.text()).toContain('+')
      expect(earned.text()).toContain('50')
    })

    it('shows potential points when not achieved', async () => {
      const wrapper = await mountWith(
        makeAchievement({ achievedAt: null, points: 50 }),
      )

      // Not-yet-earned points use the muted style, not the accent one.
      expect(wrapper.find('.text-text-muted').exists()).toBe(true)
      expect(wrapper.find('.text-accent-contrast').exists()).toBe(false)
    })

    it('shows no points block when the achievement is worth zero points', async () => {
      const wrapper = await mountWith(makeAchievement({ points: 0 }))

      expect(wrapper.find('.text-accent-contrast').exists()).toBe(false)
      expect(wrapper.find('.text-text-muted').exists()).toBe(false)
    })
  })

  describe('interaction', () => {
    it('opens automatically when its id is the active sheet id', async () => {
      const achievement = makeAchievement({
        id: 'AC_ACTIVE',
        achievedAt: '2026-01-01T00:00:00Z',
      })
      await mountWith(achievement, { openId: 'AC_ACTIVE' })
      await nextTick()

      expect(track).toHaveBeenCalledWith(
        'achievement_clicked',
        expect.objectContaining({
          achievement_id: 'AC_ACTIVE',
          is_unlocked: true,
        }),
      )
    })

    it('marks an unlocked-but-uncelebrated achievement celebrated on close', async () => {
      const achievement = makeAchievement({
        id: 'AC_ACTIVE',
        achievedAt: '2026-01-01T00:00:00Z',
        celebratedAt: null,
      })
      const wrapper = await mountWith(achievement, { openId: 'AC_ACTIVE' })
      await nextTick()

      // Simulate the drawer closing.
      await wrapper
        .findComponent(DesignDrawerStub)
        .vm.$emit('update:open', false)
      await nextTick()

      expect(markCelebrated).toHaveBeenCalledWith({
        achievementId: 'AC_ACTIVE',
      })
    })

    it('does not mark celebrated on close when already celebrated', async () => {
      const achievement = makeAchievement({
        id: 'AC_ACTIVE',
        achievedAt: '2026-01-01T00:00:00Z',
        celebratedAt: '2026-01-02T00:00:00Z',
      })
      const wrapper = await mountWith(achievement, { openId: 'AC_ACTIVE' })
      await nextTick()

      await wrapper
        .findComponent(DesignDrawerStub)
        .vm.$emit('update:open', false)
      await nextTick()

      expect(markCelebrated).not.toHaveBeenCalled()
    })
  })
})
