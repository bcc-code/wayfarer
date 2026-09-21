// @vitest-environment nuxt
import { describe, it, expect, vi, beforeEach } from 'vitest'
import { mountSuspended, mockNuxtImport } from '@nuxt/test-utils/runtime'
import { defineComponent, ref, computed } from 'vue'
import { RoleType, ScopeType } from '../../app/api/generated'
import AdminUserRoles from '../../layers/admin/app/components/admin/user/AdminUserRoles.vue'

const assignRole = vi.fn(() => Promise.resolve({ error: undefined }))
const revokeRole = vi.fn(() => Promise.resolve({ error: undefined }))

mockNuxtImport('useAssignRoleMutation', () => () => ({
  executeMutation: assignRole,
}))
mockNuxtImport('useRevokeRoleMutation', () => () => ({
  executeMutation: revokeRole,
}))
mockNuxtImport('useAuthReady', () => () => ({ isAuthReady: ref(true) }))

mockNuxtImport('useAdminUserRoleScopeOptionsQuery', () => () => ({
  data: computed(() => ({
    churches: { edges: [{ node: { id: 'CH1', name: 'Østfold' } }] },
    projects: { edges: [{ node: { id: 'PR1', name: 'Sommerleir' } }] },
  })),
  fetching: ref(false),
  error: ref(undefined),
}))

mockNuxtImport('useAdminUserRoleTeamOptionsQuery', () => () => ({
  data: computed(() => ({
    teams: { edges: [{ node: { id: 'TM1', name: 'Sigvus' } }] },
  })),
  fetching: ref(false),
  error: ref(undefined),
}))

/**
 * `UModal` teleports its content and only renders it while open, so its slots
 * are invisible to the wrapper. Stubbed to render them inline — the repo's
 * testing notes call this out as the one case where stubbing a child is
 * required rather than preferred.
 */
const UModalStub = defineComponent({
  props: { open: Boolean },
  emits: ['update:open'],
  template:
    '<div data-modal><slot name="header" /><slot name="body" /><slot name="footer" /></div>',
})

const mount = (roles: unknown[] = []) =>
  mountSuspended(AdminUserRoles, {
    props: { userId: 'US1', roles, canManage: true },
    global: { stubs: { UModal: UModalStub } },
  })

const submitButton = (wrapper: Awaited<ReturnType<typeof mount>>) =>
  wrapper.findAll('button').find((b) => b.text() === 'Tildel rolle')

describe('AdminUserRoles', () => {
  beforeEach(() => {
    assignRole.mockClear()
    revokeRole.mockClear()
  })

  it('names a role scope instead of printing its id', async () => {
    const wrapper = await mount([
      {
        id: 'UR1',
        role: RoleType.ChurchAdmin,
        scope: {
          id: 'CH01K9VZ865699692N7FVTXYR4AQ',
          type: ScopeType.Church,
          church: { id: 'CH01K9VZ865699692N7FVTXYR4AQ', name: 'Østfold' },
        },
      },
    ])

    expect(wrapper.text()).toContain('Østfold')
    expect(wrapper.text()).not.toContain('CH01K9VZ865699692N7FVTXYR4AQ')
  })

  // A scope pointing at something deleted still has to render, and then the raw
  // id is the only honest thing left to show.
  it('falls back to the id when nothing resolved', async () => {
    const wrapper = await mount([
      {
        id: 'UR1',
        role: RoleType.ChurchAdmin,
        scope: { id: 'CH_GONE', type: ScopeType.Church },
      },
    ])

    expect(wrapper.text()).toContain('CH_GONE')
  })

  it('revokes with the role and its scope', async () => {
    const wrapper = await mount([
      {
        id: 'UR1',
        role: RoleType.TeamLead,
        scope: {
          id: 'TM1',
          type: ScopeType.Team,
          team: { id: 'TM1', name: 'Sigvus' },
        },
      },
    ])

    const remove = wrapper
      .findAll('button')
      .find((b) => b.attributes('aria-label')?.startsWith('Fjern rollen'))
    await remove!.trigger('click')

    expect(revokeRole).toHaveBeenCalledWith({
      input: {
        userId: 'US1',
        role: RoleType.TeamLead,
        scopeType: ScopeType.Team,
        scopeId: 'TM1',
      },
    })
  })

  // A global role needs no scope, so the form must not demand one.
  it('submits a global role with no scope', async () => {
    const wrapper = await mount()

    const submit = submitButton(wrapper)
    expect(submit?.attributes('disabled')).toBeUndefined()

    await submit!.trigger('click')

    expect(assignRole).toHaveBeenCalledWith({
      input: {
        userId: 'US1',
        role: RoleType.User,
        scopeType: null,
        scopeId: undefined,
      },
    })
  })

  // The rule this dialog now enforces: a scoped role submitted without its
  // scope would be assigned *globally* — a far larger grant than was asked
  // for. The old form allowed exactly that.
  it('blocks a scoped role until its scope is chosen', async () => {
    const wrapper = await mount()
    const select = wrapper.findComponent({ name: 'USelect' })
    await select.setValue(RoleType.ChurchAdmin)

    expect(submitButton(wrapper)?.attributes('disabled')).toBeDefined()

    await submitButton(wrapper)!.trigger('click')
    expect(assignRole).not.toHaveBeenCalled()
  })

  // The scope type is no longer a question: the role determines it.
  it('derives the scope type from the role', async () => {
    const wrapper = await mount()
    await wrapper.findComponent({ name: 'USelect' }).setValue(RoleType.TeamLead)

    expect(wrapper.text()).toContain('Prosjekt')
    expect(wrapper.text()).toContain('Lag')
    expect(wrapper.text()).not.toContain('Omfangstype')
  })

  it('says so when a role needs no scope at all', async () => {
    const wrapper = await mount()
    await wrapper.findComponent({ name: 'USelect' }).setValue(RoleType.Admin)

    expect(wrapper.text()).toContain('gjelder globalt')
  })
})
