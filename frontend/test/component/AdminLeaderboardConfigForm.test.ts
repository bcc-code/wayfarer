// @vitest-environment nuxt
import { describe, it, expect } from 'vitest'
import { mountSuspended, mockNuxtImport } from '@nuxt/test-utils/runtime'
import { ref } from 'vue'
import type { ZodType } from 'zod'
import AdminLeaderboardConfigForm from '../../layers/admin/app/components/admin/leaderboard/AdminLeaderboardConfigForm.vue'
import {
  ChurchCategory,
  Gender,
  LeaderboardEntityType,
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

// `''` from an emptied UInput, `null` from a cleared USelectMenu — the two
// shapes the controls actually produce for "unset".
const emptyFilterState = {
  minScore: '',
  maxScore: '',
  churchId: null,
  country: '',
  churchCategory: null,
  gender: null,
  ageMin: '',
  ageMax: '',
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
  // `UpdateLeaderboardConfigInput` is full-replace, so a filter field with no
  // control would be wiped the first time an existing config is saved.
  it('has a control for every LeaderboardFilter field', async () => {
    const wrapper = await mount()

    expect(fieldNames(wrapper)).toEqual(
      expect.arrayContaining([
        'filter.minScore',
        'filter.maxScore',
        'filter.churchId',
        'filter.country',
        'filter.churchCategory',
        'filter.gender',
        'filter.ageMin',
        'filter.ageMax',
        'filter.teamId',
        'filter.superTeamId',
      ]),
    )
  })

  // Event scope is fixed at creation — `UpdateLeaderboardConfigInput` has no
  // `eventId` — so offering the picker in edit mode invites a change that
  // cannot be saved.
  it('offers the event picker only when creating', async () => {
    const creating = await mount()
    expect(fieldNames(creating)).toContain('eventId')

    const editing = await mount({ initialData, isEditMode: true })
    expect(fieldNames(editing)).not.toContain('eventId')
  })

  it('fills the form from an existing config, ageRange flattened', async () => {
    const wrapper = await mount({ initialData, isEditMode: true })

    const valueOf = (name: string, component: string) =>
      field(wrapper, name)
        ?.findComponent({ name: component })
        .props('modelValue')

    expect(valueOf('name', 'UInput')).toBe('Topp 20')
    expect(valueOf('sortOrder', 'UInput')).toBe(2)
    expect(valueOf('maxEntries', 'UInput')).toBe(20)
    expect(valueOf('isActive', 'USwitch')).toBe(false)
    expect(valueOf('filter.ageMin', 'UInput')).toBe(13)
    expect(valueOf('filter.ageMax', 'UInput')).toBe(18)
    expect(valueOf('filter.gender', 'USelectMenu')).toBe(Gender.Female)
    expect(valueOf('filter.churchCategory', 'USelectMenu')).toBe(
      ChurchCategory.Xl,
    )
    // Unset on the server, so unset in the form — not the string "null".
    expect(valueOf('filter.country', 'UInput')).toBe('')
  })

  // reka-ui rejects a SelectItem whose value is '', reserving that value for
  // "cleared" — so an optional picker has to be clearable rather than carry an
  // explicit "Alle" item, or it can be set once and never unset.
  it('makes every optional picker clearable, with no empty-valued item', async () => {
    const wrapper = await mount()

    for (const name of [
      'eventId',
      'filter.gender',
      'filter.churchCategory',
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
      name: 'UInput',
    })
    await input.vm.$emit('update:modelValue', 5)
    await submit(wrapper, { ...state })
    expect(wrapper.emitted('submit')?.[0]?.[0]).toMatchObject({ maxEntries: 5 })
    await input.vm.$emit('update:modelValue', '')
    await submit(wrapper, { ...state })
    expect(wrapper.emitted('submit')?.[1]?.[0]).toMatchObject({
      maxEntries: null,
    })
  })

  it('accepts only positive integer limits or a blank field', async () => {
    const wrapper = await mount({ initialData, isEditMode: true })
    const form = wrapper.findComponent({ name: 'UForm' })
    const schema = form.props('schema') as ZodType
    for (const maxEntries of [0, -1, 1.5, 2147483648]) {
      expect(
        schema.safeParse({ ...form.props('state'), maxEntries }).success,
      ).toBe(false)
    }
    for (const maxEntries of ['', 1, 150]) {
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
