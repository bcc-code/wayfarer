// @vitest-environment nuxt
import { describe, it, expect, beforeEach, vi } from 'vitest'
import { ref, nextTick } from 'vue'
import { flushPromises } from '@vue/test-utils'
import { mountSuspended, mockNuxtImport } from '@nuxt/test-utils/runtime'
import StandingsUnit from '~/components/standings/StandingsUnit.vue'
import DesignButton from '~/components/design/DesignButton.vue'
import { LeaderboardEntryTag } from '~/api/generated'

// --- Auto-import mocks --------------------------------------------------------
// Composables/queries have no real backend, auth, or router in a component
// test, so they must be mocked. Child components render for real (below).
const {
  queryMock,
  authMock,
  authReadyMock,
  updateTeamMock,
  assignLeadMock,
  nowMock,
  routerReplace,
} = vi.hoisted(() => ({
  queryMock: vi.fn(),
  authMock: vi.fn(),
  authReadyMock: vi.fn(),
  updateTeamMock: vi.fn(),
  assignLeadMock: vi.fn(),
  nowMock: vi.fn(),
  routerReplace: vi.fn(),
}))
mockNuxtImport('useStandingsUnitPageQuery', () => queryMock)
mockNuxtImport('useAuth', () => authMock)
mockNuxtImport('useAuthReady', () => authReadyMock)
mockNuxtImport('useUpdateTeamMutation', () => updateTeamMock)
mockNuxtImport('useAssignTeamLeadMutation', () => assignLeadMock)
mockNuxtImport('useNow', () => nowMock)
// v4 (@nuxt/test-utils 4) inits Nuxt in beforeAll, so its router plugin calls
// beforeEach/options on useRouter() during setup. Wrap the *real* router (via
// the original passed to the factory) and only override replace with our spy,
// rather than replacing the whole router with a bare stub.
mockNuxtImport('useRouter', (original) => () => {
  const router = original()
  router.replace = routerReplace
  return router
})

// Spies we assert against.
const updateTeam = vi.fn()
const assignLead = vi.fn()
const refetch = vi.fn()

// --- Stubs -------------------------------------------------------------------
// Only two child components are stubbed, and both for a concrete reason:
//   - DesignDrawer wraps @nuxt/ui's UDrawer, which teleports its content out of
//     the wrapper and only renders it while open. This stub renders both slots
//     inline so the edit form is queryable and open/close is deterministic.
//   - LeaderboardList runs entrance-animation side effects and is irrelevant to
//     this flow.
// DesignButton / DesignInput / DesignPanel / icons all render for real.
const DesignDrawerStub = {
  props: ['open', 'title', 'nested', 'dismissible'],
  emits: ['update:open'],
  template: '<div><slot /><slot name="content" /></div>',
}
const stubs = {
  DesignDrawer: DesignDrawerStub,
  LeaderboardList: true,
}

type Member = { id: string; name: string; rank?: number; tags?: string[] }

function makeData(name: string, members: Member[]) {
  return {
    myCurrentProject: {
      myTeam: {
        id: 'TM01',
        name,
        superTeam: null,
        memberLeaderboard: members,
      },
    },
  }
}

const MEMBERS: Member[] = [
  { id: 'US01', name: 'Alice', rank: 1, tags: [LeaderboardEntryTag.TeamLead] },
  { id: 'US02', name: 'Bob', rank: 2, tags: [] },
]

// Mounts as a team lead with no data yet, then resolves the query so the
// `watch(data, { once: true })` form initialization runs (it has no immediate).
async function mountEdit(opts: { now?: Date } = {}) {
  const dataRef = ref<unknown>(null)
  authMock.mockReturnValue({ isTeamLead: ref(true), me: ref({ id: 'US99' }) })
  authReadyMock.mockReturnValue({ isAuthReady: ref(true) })
  queryMock.mockReturnValue({
    data: dataRef,
    error: ref(null),
    fetching: ref(false),
    executeQuery: refetch,
  })
  updateTeamMock.mockReturnValue({ executeMutation: updateTeam })
  assignLeadMock.mockReturnValue({ executeMutation: assignLead })
  nowMock.mockReturnValue(ref(opts.now ?? new Date('2030-01-01T00:00:00Z')))

  const wrapper = await mountSuspended(StandingsUnit, { global: { stubs } })

  // Simulate the query resolving after mount so form init fires.
  dataRef.value = makeData('Alpha', MEMBERS)
  await nextTick()

  return wrapper
}

// Locate the real DesignButtons by their size prop rather than a test-only hook.
const nameInput = (w: Awaited<ReturnType<typeof mountEdit>>) => w.find('input')
const buttonBySize = (w: Awaited<ReturnType<typeof mountEdit>>, size: string) =>
  w.findAllComponents(DesignButton).find((c) => c.props('size') === size)
const memberButton = (w: Awaited<ReturnType<typeof mountEdit>>, name: string) =>
  w.findAll('button').find((b) => b.text().trim() === name)!

describe('StandingsUnit edit drawer', () => {
  beforeEach(() => {
    vi.clearAllMocks()
  })

  it('initializes the form with the current team name', async () => {
    const wrapper = await mountEdit()

    expect(nameInput(wrapper).element.value).toBe('Alpha')
  })

  it('saves the edited team name via the update mutation', async () => {
    const wrapper = await mountEdit()

    await nameInput(wrapper).setValue('Beta')
    await buttonBySize(wrapper, 'large')!.trigger('click')
    await flushPromises()

    expect(updateTeam).toHaveBeenCalledWith({
      id: 'TM01',
      input: { name: 'Beta' },
    })
  })

  it('refetches and closes the drawer after saving', async () => {
    const wrapper = await mountEdit()

    await buttonBySize(wrapper, 'large')!.trigger('click')
    await flushPromises()

    expect(refetch).toHaveBeenCalled()
    // Closing the drawer syncs showEditDrawer=false to the URL query.
    expect(routerReplace).toHaveBeenCalled()
  })

  it('assigns a new team lead when the selection changed', async () => {
    const wrapper = await mountEdit()

    // Pick Bob (currently Alice is the lead).
    await memberButton(wrapper, 'Bob').trigger('click')
    await buttonBySize(wrapper, 'large')!.trigger('click')
    await flushPromises()

    expect(assignLead).toHaveBeenCalledWith({
      teamId: 'TM01',
      userId: 'US02',
    })
  })

  it('does not assign a team lead when the selection is unchanged', async () => {
    const wrapper = await mountEdit()

    // Save without changing the lead (Alice remains).
    await buttonBySize(wrapper, 'large')!.trigger('click')
    await flushPromises()

    expect(updateTeam).toHaveBeenCalled()
    expect(assignLead).not.toHaveBeenCalled()
  })

  it('updates the leader selector label when a member is picked', async () => {
    const wrapper = await mountEdit()

    // The leader-select button initially reflects the current lead.
    expect(wrapper.text()).toContain('Alice')

    await memberButton(wrapper, 'Bob').trigger('click')
    await nextTick()

    expect(buttonBySize(wrapper, 'small')!.text()).toContain('Bob')
  })

  it('hides the edit trigger button before the release date', async () => {
    const wrapper = await mountEdit({ now: new Date('2020-01-01T00:00:00Z') })

    expect(buttonBySize(wrapper, 'medium')).toBeUndefined()
  })

  it('shows the edit trigger button after the release date', async () => {
    const wrapper = await mountEdit({ now: new Date('2030-01-01T00:00:00Z') })

    expect(buttonBySize(wrapper, 'medium')).toBeTruthy()
  })
})
