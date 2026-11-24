// @vitest-environment nuxt
import { describe, it, expect, vi, beforeEach } from 'vitest'
import { mountSuspended, mockNuxtImport } from '@nuxt/test-utils/runtime'
import AdminConsentTable from '../../layers/admin/app/components/admin/consent/AdminConsentTable.vue'
import type { AdminConsentRow } from '../../layers/admin/app/utils/consents'

const { archiveConsent, unarchiveConsent, confirm } = vi.hoisted(() => ({
  archiveConsent: vi.fn().mockResolvedValue({}),
  unarchiveConsent: vi.fn().mockResolvedValue({}),
  confirm: vi.fn().mockResolvedValue(true),
}))

mockNuxtImport('useArchiveConsentMutation', () => () => ({
  executeMutation: archiveConsent,
}))
mockNuxtImport('useUnarchiveConsentMutation', () => () => ({
  executeMutation: unarchiveConsent,
}))
mockNuxtImport('useConfirm', () => () => ({ confirm }))
mockNuxtImport('useToast', () => () => ({ add: vi.fn() }))

const live: AdminConsentRow = {
  id: 'CN_LIVE',
  key: 'privacy_policy',
  version: 1,
  title: 'Personvern',
  publishedAt: '2026-01-10T12:00:00Z',
  archivedAt: null,
  managementType: 'LOCAL',
  isRequired: true,
  project: { id: 'PR1', name: 'Sommerleir' },
}

const archived: AdminConsentRow = {
  id: 'CN_ARCHIVED',
  key: 'old_terms',
  version: 1,
  title: 'Gamle vilkår',
  publishedAt: '2025-01-10T12:00:00Z',
  archivedAt: '2026-03-01T12:00:00Z',
  managementType: 'LOCAL',
  isRequired: false,
  project: null,
}

function mountTable() {
  return mountSuspended(AdminConsentTable, {
    props: { consents: [live, archived], canManage: true, showProject: true },
  })
}

describe('AdminConsentTable', () => {
  beforeEach(() => {
    vi.clearAllMocks()
  })

  it('labels a live consent and a retired one differently', async () => {
    const wrapper = await mountTable()

    expect(wrapper.text()).toContain('Publisert')
    expect(wrapper.text()).toContain('Arkivert')
  })

  it('archives a consent after confirming, keeping the answers', async () => {
    const wrapper = await mountTable()

    await wrapper.find('[aria-label="Arkiver Personvern"]').trigger('click')
    await Promise.resolve()

    expect(confirm).toHaveBeenCalledWith(
      expect.objectContaining({
        description: expect.stringContaining('beholdes'),
      }),
    )
    expect(archiveConsent).toHaveBeenCalledWith({ id: 'CN_LIVE' })
    expect(wrapper.emitted('refresh')).toBeTruthy()
  })

  it('does not archive when the confirmation is dismissed', async () => {
    confirm.mockResolvedValueOnce(false)
    const wrapper = await mountTable()

    await wrapper.find('[aria-label="Arkiver Personvern"]').trigger('click')
    await Promise.resolve()

    expect(archiveConsent).not.toHaveBeenCalled()
  })

  it('offers restore instead of archive for an archived consent', async () => {
    const wrapper = await mountTable()

    expect(wrapper.find('[aria-label="Arkiver Gamle vilkår"]').exists()).toBe(
      false,
    )

    await wrapper
      .find('[aria-label="Gjenopprett Gamle vilkår"]')
      .trigger('click')
    await Promise.resolve()

    expect(unarchiveConsent).toHaveBeenCalledWith({ id: 'CN_ARCHIVED' })
    expect(wrapper.emitted('refresh')).toBeTruthy()
  })

  it('hides both actions without manage permission', async () => {
    const wrapper = await mountSuspended(AdminConsentTable, {
      props: { consents: [live], canManage: false },
    })

    expect(wrapper.find('[aria-label="Arkiver Personvern"]').exists()).toBe(
      false,
    )
  })
})
