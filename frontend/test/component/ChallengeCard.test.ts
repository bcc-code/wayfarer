// @vitest-environment nuxt
import { describe, it, expect, vi, beforeEach } from 'vitest'
import { mountSuspended, mockNuxtImport } from '@nuxt/test-utils/runtime'
import ChallengeCard from '~/components/challenges/ChallengeCard.vue'
import DesignButton from '~/components/design/DesignButton.vue'

// useAnalytics is a Nuxt auto-import; mockNuxtImport replaces it globally.
const track = vi.fn()
mockNuxtImport('useAnalytics', () => () => ({
  track,
  identify: vi.fn(),
  page: vi.fn(),
  reset: vi.fn(),
}))

// Minimal factory for the challenge prop the card expects. Only the fields the
// component actually reads are populated; cast keeps us free of the full type.
type Challenge = InstanceType<typeof ChallengeCard>['challenge']
function makeChallenge(overrides: Partial<Challenge> = {}): Challenge {
  return {
    __typename: 'SimpleChallenge',
    id: 'CL01ARZ3NDEKTSV4RRFFQ69G5FAV',
    name: 'Read the intro',
    description: '<p>Do the thing</p>',
    imageObject: null,
    buttonText: 'Start',
    userCompletedAt: null,
    ...overrides,
  } as Challenge
}

// All children render for real (DesignImage, DesignButton, NuxtLink). Only the
// useAnalytics composable is mocked, above.
const global = {}

describe('ChallengeCard', () => {
  beforeEach(() => {
    vi.clearAllMocks()
  })

  it('renders the challenge name and description', async () => {
    const wrapper = await mountSuspended(ChallengeCard, {
      props: { challenge: makeChallenge() },
      global,
    })

    expect(wrapper.text()).toContain('Read the intro')
    expect(wrapper.html()).toContain('Do the thing')
  })

  it('shows the primary button variant when not completed', async () => {
    const wrapper = await mountSuspended(ChallengeCard, {
      props: { challenge: makeChallenge({ userCompletedAt: null }) },
      global,
    })

    const button = wrapper.findComponent(DesignButton)
    expect(button.props('variant')).toBe('primary')
  })

  it('shows the secondary button variant when completed', async () => {
    const wrapper = await mountSuspended(ChallengeCard, {
      props: {
        challenge: makeChallenge({ userCompletedAt: '2026-01-01T00:00:00Z' }),
      },
      global,
    })

    const button = wrapper.findComponent(DesignButton)
    expect(button.props('variant')).toBe('secondary')
  })

  it('tracks an analytics event when the challenge link is clicked', async () => {
    const wrapper = await mountSuspended(ChallengeCard, {
      props: { challenge: makeChallenge({ id: 'CL01TESTID' }) },
      global,
    })

    await wrapper.find('a').trigger('click')

    expect(track).toHaveBeenCalledWith(
      'challenge_link_clicked',
      expect.objectContaining({
        challenge_id: 'CL01TESTID',
        is_external: false,
      }),
    )
  })
})
