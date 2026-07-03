// @vitest-environment nuxt
import { describe, it, expect, vi, beforeEach } from 'vitest'
import { mountSuspended, mockNuxtImport } from '@nuxt/test-utils/runtime'
import ChallengeCard from '~/components/challenges/ChallengeCard.vue'

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

// Stub heavy/global child components so we test ChallengeCard in isolation.
const global = {
  stubs: {
    // `true` = auto-stub: renders a placeholder but preserves props and the
    // component name, so findComponent(...).props('variant') still works.
    DesignImage: true,
    DesignButton: true,
    // Custom stub renders a real <a> so we can trigger a click event.
    NuxtLink: { template: '<a><slot /></a>' },
  },
}

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

    const button = wrapper.findComponent({ name: 'DesignButton' })
    expect(button.props('variant')).toBe('primary')
  })

  it('shows the secondary button variant when completed', async () => {
    const wrapper = await mountSuspended(ChallengeCard, {
      props: {
        challenge: makeChallenge({ userCompletedAt: '2026-01-01T00:00:00Z' }),
      },
      global,
    })

    const button = wrapper.findComponent({ name: 'DesignButton' })
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
