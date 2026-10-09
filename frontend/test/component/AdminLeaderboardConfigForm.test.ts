// @vitest-environment nuxt
import { describe, it, expect } from 'vitest'
import { mountSuspended, mockNuxtImport } from '@nuxt/test-utils/runtime'
import { flushPromises } from '@vue/test-utils'
import { ref } from 'vue'
import type { ZodType } from 'zod'
import AdminLeaderboardConfigForm from '../../layers/admin/app/components/admin/leaderboard/AdminLeaderboardConfigForm.vue'
import {
  ChurchCategory,
  Gender,
  LeaderboardEntityType,
  LeaderboardLimitMode,
} from '../../app/api/generated'

const options = ref({
  project: {
    id: 'PR1',
    events: [{ id: 'EV1', name: 'Sommerleir' }],
  },
  teams: { edges: [{ node: { id: 'TM1', name: 'Lag 1' } }] },
  superteams: { edges: [{ node: { id: 'ST1', name: 'Superlag 1' } }] },
  churches: { edges: [{ node: { id: 'CH1', name: 'Oslo' } }] },
})

mockNuxtImport('useAdminLeaderboardConfigFormOptionsQuery', () => () => ({
  data: options,
  fetching: ref(false),
  error: ref(undefined),
}))

mockNuxtImport('useAuthReady', () => () => ({ isAuthReady: ref(true) }))

const initialData = {
  id: 'LC1',
  name: 'Topp 20',
  entityType: LeaderboardEntityType.Persons,
  limitMode: LeaderboardLimitMode.Manual,
  maxEntries: 20,
  sortOrder: 2,
  isActive: false,
  event: null,
  filter: {
    minScore: null,
    maxScore: null,
    churchId: null,
    country: null,
    churchCategory: ChurchCategory.Xl,
    gender: Gender.Female,
    ageRange: { min: 13, max: 18 },
    teamId: null,
    superTeamId: null,
  },
}

// `undefined` from an empty stepper field, `null` from a cleared USelectMenu —
// the two shapes the controls actually produce for "unset".
const emptyFilterState = {
  minScore: undefined,
  maxScore: undefined,
  churchId: null,
  country: '',
  churchCategory: null,
  gender: null,
  ageMin: undefined,
  ageMax: undefined,
  teamId: null,
  superTeamId: null,
}

const mount = (props: Record<string, unknown> = {}) =>
  mountSuspended(AdminLeaderboardConfigForm, {
    props: { projectId: 'PR1', submitLabel: 'Lagre', ...props },
  })

type Wrapper = Awaited<ReturnType<typeof mount>>

const fieldNames = (wrapper: Wrapper) =>
  wrapper
    .findAllComponents({ name: 'UFormField' })
    .map((field) => field.props('name'))

const field = (wrapper: Wrapper, name: string) =>
  wrapper
    .findAllComponents({ name: 'UFormField' })
    .find((f) => f.props('name') === name)

/**
 * The template binds `@submit.prevent`, so the handler is handed a real submit
 * event in the app. `$emit` alone gives it a bare object and the `.prevent`
 * modifier throws before the handler runs.
 */
const submit = (wrapper: Wrapper, data: Record<string, unknown>) =>
  wrapper
    .findComponent({ name: 'UForm' })
    .vm.$emit('submit', { data, preventDefault: () => {} })

describe('AdminLeaderboardConfigForm', () => {
  // A persons board is the one entity type every filter below applies to.
  it('has a control for every filter a persons board applies', async () => {
    const wrapper = await mount()

    expect(fieldNames(wrapper)).toEqual(
      expect.arrayContaining([
        'filter.minScore',
        'filter.maxScore',
        'filter.churchId',
        'filter.ageMin',
        'filter.ageMax',
        'filter.teamId',
        'filter.superTeamId',
      ]),
    )
  })

  // These three reach `buildFilterParamsMap` (the cache key) and no `GetFull*`
  // parameter builder, so setting one narrows nothing. No control until the
  // backend applies them.
  it('offers no control for the filters the backend ignores', async () => {
    const wrapper = await mount()

    for (const name of [
      'filter.gender',
      'filter.country',
      'filter.churchCategory',
    ])
      expect(fieldNames(wrapper)).not.toContain(name)
  })

  // Events are not in use yet, and the scope could only ever be set at
  // creation — `UpdateLeaderboardConfigInput` has no `eventId`. The manual sort
  // position went the same way.
  it('offers neither event scope nor sort position', async () => {
    const creating = await mount()
    const editing = await mount({ initialData, isEditMode: true })

    for (const wrapper of [creating, editing]) {
      expect(fieldNames(wrapper)).not.toContain('eventId')
      expect(fieldNames(wrapper)).not.toContain('sortOrder')
    }
  })

  // The bounds are inclusive and counted off the birth year, which nothing in
  // "Alder fra"/"Alder til" said.
  it('says how the age bounds are counted', async () => {
    const wrapper = await mount()

    const text = wrapper.text()
    expect(text).toContain('Yngste alder')
    expect(text).toContain('Eldste alder')
    expect(text).toContain('Alder regnes etter fødselsår, ikke bursdag')
  })

  it('fills the age bounds from a fixed age group', async () => {
    const wrapper = await mount({ initialData, isEditMode: true })

    const u36 = wrapper
      .findAllComponents({ name: 'UButton' })
      .find((button) => button.text() === 'U36')!
    await u36.trigger('click')
    await wrapper.find('form').trigger('submit.prevent')
    await flushPromises()

    expect(u36.attributes('aria-pressed')).toBe('true')
    expect(wrapper.emitted('submit')?.[0]?.[0]).toMatchObject({
      filter: { ageRange: { min: 18, max: 35 } },
    })
  })

  // Hidden, not dropped: an existing board keeps the position it was given.
  it('saves the sort position it was given', async () => {
    const wrapper = await mount({ initialData, isEditMode: true })

    await wrapper.find('form').trigger('submit.prevent')
    await flushPromises()

    expect(wrapper.emitted('submit')?.[0]?.[0]).toMatchObject({ sortOrder: 2 })
  })

  it('fills the form from an existing config, ageRange flattened', async () => {
    const wrapper = await mount({ initialData, isEditMode: true })

    const valueOf = (name: string, component: string) =>
      field(wrapper, name)
        ?.findComponent({ name: component })
        .props('modelValue')

    expect(valueOf('name', 'UInput')).toBe('Topp 20')
    expect(valueOf('maxEntries', 'UInputNumber')).toBe(20)
    expect(valueOf('isActive', 'USwitch')).toBe(false)
    expect(valueOf('filter.ageMin', 'UInputNumber')).toBe(13)
    expect(valueOf('filter.ageMax', 'UInputNumber')).toBe(18)
  })

  // The update input is full-replace, so a stored value with no control has to
  // survive an edit rather than be dropped by the first save.
  it('resends the uncontrolled filters it loaded', async () => {
    const wrapper = await mount({ initialData, isEditMode: true })

    await wrapper.find('form').trigger('submit.prevent')
    await flushPromises()

    expect(wrapper.emitted('submit')?.[0]?.[0]).toMatchObject({
      filter: { gender: Gender.Female, churchCategory: ChurchCategory.Xl },
    })
  })

  // reka-ui rejects a SelectItem whose value is '', reserving that value for
  // "cleared" — so an optional picker has to be clearable rather than carry an
  // explicit "Alle" item, or it can be set once and never unset.
  it('makes every optional picker clearable, with no empty-valued item', async () => {
    const wrapper = await mount()

    for (const name of [
      'filter.churchId',
      'filter.teamId',
      'filter.superTeamId',
    ]) {
      const menu = field(wrapper, name)?.findComponent({ name: 'USelectMenu' })
      const items = menu?.props('items') as { value: string }[]

      expect(menu?.props('clear'), name).toBeTruthy()
      expect(
        items.every((item) => item.value !== ''),
        name,
      ).toBe(true)
    }
  })

  it('emits no filter at all when nothing is set', async () => {
    const wrapper = await mount()

    await submit(wrapper, {
      name: 'Alle',
      entityType: LeaderboardEntityType.Persons,
      limitMode: LeaderboardLimitMode.Manual,
      eventId: null,
      sortOrder: 0,
      isActive: true,
      filter: { ...emptyFilterState },
    })

    expect(wrapper.emitted('submit')?.[0]?.[0]).toMatchObject({
      eventId: null,
      maxEntries: null,
      filter: null,
    })
  })

  it('saves an entry limit and clears it when the input is emptied', async () => {
    const wrapper = await mount({ initialData, isEditMode: true })
    const form = wrapper.findComponent({ name: 'UForm' })
    const state = form.props('state')
    const input = field(wrapper, 'maxEntries')!.findComponent({
      name: 'UInputNumber',
    })
    await input.vm.$emit('update:modelValue', 5)
    await submit(wrapper, { ...state })
    expect(wrapper.emitted('submit')?.[0]?.[0]).toMatchObject({ maxEntries: 5 })
    // What a stepper field holds once you clear it.
    await input.vm.$emit('update:modelValue', undefined)
    await submit(wrapper, { ...state })
    expect(wrapper.emitted('submit')?.[1]?.[0]).toMatchObject({
      maxEntries: null,
    })
  })

  // A stepper field holds `undefined` when empty, never `''`.
  it('accepts only positive integer limits or an empty field', async () => {
    const wrapper = await mount({ initialData, isEditMode: true })
    const form = wrapper.findComponent({ name: 'UForm' })
    const schema = form.props('schema') as ZodType
    for (const maxEntries of [0, -1, 1.5, 2147483648, '']) {
      expect(
        schema.safeParse({ ...form.props('state'), maxEntries }).success,
      ).toBe(false)
    }
    for (const maxEntries of [undefined, 1, 150]) {
      expect(
        schema.safeParse({ ...form.props('state'), maxEntries }).success,
      ).toBe(true)
    }
  })

  it('rebuilds ageRange from the two age controls', async () => {
    const wrapper = await mount()

    await submit(wrapper, {
      name: 'Ungdom',
      entityType: LeaderboardEntityType.Teams,
      limitMode: LeaderboardLimitMode.Manual,
      eventId: 'EV1',
      sortOrder: 1,
      isActive: true,
      filter: {
        ...emptyFilterState,
        gender: Gender.Male,
        ageMin: 13,
        ageMax: 18,
      },
    })

    expect(wrapper.emitted('submit')?.[0]?.[0]).toEqual({
      name: 'Ungdom',
      entityType: LeaderboardEntityType.Teams,
      limitMode: LeaderboardLimitMode.Manual,
      eventId: 'EV1',
      sortOrder: 1,
      isActive: true,
      maxEntries: null,
      filter: { gender: Gender.Male, ageRange: { min: 13, max: 18 } },
    })
  })

  // A zero bound is a real value; `''` is the only "unset".
  it('keeps a zero score bound', async () => {
    const wrapper = await mount()

    await submit(wrapper, {
      name: 'Fra null',
      entityType: LeaderboardEntityType.Persons,
      limitMode: LeaderboardLimitMode.Manual,
      eventId: null,
      sortOrder: 0,
      isActive: true,
      filter: { ...emptyFilterState, minScore: 0 },
    })

    expect(wrapper.emitted('submit')?.[0]?.[0]).toMatchObject({
      filter: { minScore: 0 },
    })
  })
})

describe('automatic leaderboard limits', () => {
  it('starts a new config on the manual limit', async () => {
    const creating = await mount()

    expect(
      field(creating, 'limitMode')
        ?.findComponent({ name: 'USelect' })
        .props('modelValue'),
    ).toBe(LeaderboardLimitMode.Manual)
    expect(fieldNames(creating)).toContain('maxEntries')
  })

  it('loads automatic mode and preserves the optional cap on submit', async () => {
    const wrapper = await mount({
      initialData: {
        ...initialData,
        limitMode: LeaderboardLimitMode.ChurchSize,
        filter: { churchId: 'CH1' },
      },
    })
    expect(fieldNames(wrapper)).toContain('maxEntries')
    expect(wrapper.text()).toContain('100–199: topp 50')
    expect(wrapper.text()).toContain('200 eller flere: topp 100')
    const form = wrapper.findComponent({ name: 'UForm' })
    await submit(wrapper, { ...form.props('state') })
    expect(wrapper.emitted('submit')?.[0]?.[0]).toMatchObject({
      limitMode: LeaderboardLimitMode.ChurchSize,
      maxEntries: 20,
      filter: { churchId: 'CH1' },
    })
  })

  it('requires a church for automatic limits', async () => {
    const wrapper = await mount()
    const form = wrapper.findComponent({ name: 'UForm' })
    const schema = form.props('schema') as ZodType
    const state = {
      ...form.props('state'),
      name: 'Local',
      limitMode: LeaderboardLimitMode.ChurchSize,
    }
    expect(schema.safeParse(state).success).toBe(false)
    const valid = { ...state, filter: { ...emptyFilterState, churchId: 'CH1' } }
    expect(schema.safeParse(valid).success).toBe(true)
  })

  /*
   * `leaderboardLimitModeToDB` rejects CHURCH_SIZE without a persons entity
   * type and a concrete churchId, and `myChurch` only resolves per viewer.
   * Offering the mode in either case would just fail the save, so the whole
   * picker goes once MANUAL is the only choice left.
   */
  it('withdraws automatic mode when the config can no longer use it', async () => {
    const wrapper = await mount({
      initialData: {
        ...initialData,
        limitMode: LeaderboardLimitMode.ChurchSize,
        filter: { churchId: 'CH1' },
      },
    })
    const state = wrapper.findComponent({ name: 'UForm' }).props('state')
    expect(fieldNames(wrapper)).toContain('limitMode')

    state.entityType = LeaderboardEntityType.Teams
    await flushPromises()
    expect(fieldNames(wrapper)).not.toContain('limitMode')
    expect(state.limitMode).toBe(LeaderboardLimitMode.Manual)

    state.entityType = LeaderboardEntityType.Persons
    state.limitMode = LeaderboardLimitMode.ChurchSize
    await flushPromises()
    state.filter.myChurch = true
    await flushPromises()
    expect(fieldNames(wrapper)).not.toContain('limitMode')
    expect(state.limitMode).toBe(LeaderboardLimitMode.Manual)
  })
})

describe('filters the entity type cannot apply', () => {
  const sectionsFor = async (entityType: LeaderboardEntityType) => {
    const wrapper = await mount()
    wrapper.findComponent({ name: 'UForm' }).props('state').entityType =
      entityType
    await flushPromises()
    return fieldNames(wrapper)
  }

  // The live `GetFull*` queries take churchId on persons and teams boards,
  // and age/team/superteam on persons boards only.
  it('shows only the filters the chosen entity type reaches', async () => {
    expect(await sectionsFor(LeaderboardEntityType.Persons)).toEqual(
      expect.arrayContaining([
        'filter.churchId',
        'filter.ageMin',
        'filter.teamId',
        'filter.superTeamId',
      ]),
    )

    const teams = await sectionsFor(LeaderboardEntityType.Teams)
    expect(teams).toContain('filter.churchId')
    for (const name of ['filter.ageMin', 'filter.teamId', 'filter.superTeamId'])
      expect(teams).not.toContain(name)

    for (const entityType of [
      LeaderboardEntityType.Superteams,
      LeaderboardEntityType.Churches,
    ]) {
      const names = await sectionsFor(entityType)
      expect(names).toEqual(
        expect.arrayContaining(['filter.minScore', 'filter.maxScore']),
      )
      expect(names).not.toContain('filter.churchId')
    }
  })

  /*
   * A board ignores a fixed filter it cannot apply, so keeping one costs
   * nothing and survives a type change made by mistake. The relative flags are
   * different: `ValidateLeaderboardRelativeFilter` rejects them outright, and
   * the error would point at a control the admin can no longer see.
   */
  it('drops relative flags the backend would reject, keeps ignored filters', async () => {
    const wrapper = await mount()
    await submit(wrapper, {
      name: 'Lagtavle',
      entityType: LeaderboardEntityType.Teams,
      limitMode: LeaderboardLimitMode.Manual,
      eventId: null,
      sortOrder: 0,
      isActive: true,
      filter: {
        ...emptyFilterState,
        myChurch: true,
        myTeam: true,
        mySuperTeam: true,
        ageMin: 13,
        ageMax: 18,
        gender: Gender.Female,
      },
    })

    expect(wrapper.emitted('submit')?.[0]?.[0]).toMatchObject({
      filter: {
        myChurch: true,
        ageRange: { min: 13, max: 18 },
        gender: Gender.Female,
      },
    })
    const sent = (
      wrapper.emitted('submit')?.[0]?.[0] as { filter: Record<string, unknown> }
    ).filter
    expect(sent).not.toHaveProperty('myTeam')
    expect(sent).not.toHaveProperty('mySuperTeam')
  })
})

it('round-trips relative filters and hides their fixed counterparts', async () => {
  const filter = {
    myChurch: true,
    myTeam: true,
    mySuperTeam: true,
  }
  const wrapper = await mount({
    initialData: { ...initialData, filter },
    isEditMode: true,
  })
  for (const name of ['filter.churchId', 'filter.teamId', 'filter.superTeamId'])
    expect(fieldNames(wrapper)).not.toContain(name)
  const form = wrapper.findComponent({ name: 'UForm' })
  const schema = form.props('schema') as ZodType
  expect(schema.safeParse(form.props('state')).success).toBe(true)
  await submit(wrapper, form.props('state'))
  expect(wrapper.emitted('submit')?.[0]?.[0]).toMatchObject({ filter })
})
