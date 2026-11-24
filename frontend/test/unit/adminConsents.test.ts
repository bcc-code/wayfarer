import { describe, it, expect } from 'vitest'
import {
  consentManagementTypeLabel,
  consentPublicationStatus,
  consentStatusDate,
  groupConsentsForAdmin,
  type AdminConsentRow,
} from '../../layers/admin/app/utils/consents'

const now = new Date('2026-06-15T12:00:00Z')

function row(overrides: Partial<AdminConsentRow>): AdminConsentRow {
  return {
    id: 'CN1',
    key: 'key',
    version: 1,
    title: 'Tittel',
    publishedAt: null,
    archivedAt: null,
    managementType: 'LOCAL',
    isRequired: false,
    ...overrides,
  }
}

describe('consentPublicationStatus', () => {
  it('treats a missing date as a draft', () => {
    expect(consentPublicationStatus({ publishedAt: null }, now)).toBe('draft')
    expect(consentPublicationStatus({}, now)).toBe('draft')
  })

  it('separates a scheduled consent from a live one', () => {
    expect(
      consentPublicationStatus({ publishedAt: '2026-06-16T12:00:00Z' }, now),
    ).toBe('scheduled')
    expect(
      consentPublicationStatus({ publishedAt: '2026-06-14T12:00:00Z' }, now),
    ).toBe('published')
  })

  it('counts a date that has just passed as published', () => {
    expect(
      consentPublicationStatus({ publishedAt: '2026-06-15T11:59:59Z' }, now),
    ).toBe('published')
  })

  it('lets archived win over every publication state', () => {
    expect(
      consentPublicationStatus(
        { publishedAt: '2026-06-14T12:00:00Z', archivedAt: '2026-06-15' },
        now,
      ),
    ).toBe('archived')
    expect(
      consentPublicationStatus(
        { publishedAt: null, archivedAt: '2026-06-15' },
        now,
      ),
    ).toBe('archived')
  })
})

describe('consentStatusDate', () => {
  it('prefers the archive date over the publication date', () => {
    expect(
      consentStatusDate({
        publishedAt: '2026-01-01T00:00:00Z',
        archivedAt: '2026-06-01T00:00:00Z',
      }),
    ).toBe('2026-06-01T00:00:00Z')
  })

  it('has no date for a draft', () => {
    expect(consentStatusDate({})).toBeNull()
  })
})

describe('groupConsentsForAdmin', () => {
  it('puts drafts, published and archived in their own tables', () => {
    const draft = row({ id: 'CN_DRAFT' })
    const live = row({ id: 'CN_LIVE', publishedAt: '2026-06-14T12:00:00Z' })
    const scheduled = row({
      id: 'CN_SCHEDULED',
      publishedAt: '2026-06-16T12:00:00Z',
    })
    const archived = row({
      id: 'CN_ARCHIVED',
      publishedAt: '2026-01-01T00:00:00Z',
      archivedAt: '2026-06-01T00:00:00Z',
    })
    const archivedDraft = row({
      id: 'CN_ARCHIVED_DRAFT',
      archivedAt: '2026-06-01T00:00:00Z',
    })

    const groups = groupConsentsForAdmin(
      [draft, live, scheduled, archived, archivedDraft],
      now,
    )

    expect(groups.drafts.map((c) => c.id)).toEqual(['CN_DRAFT'])
    expect(groups.published.map((c) => c.id)).toEqual([
      'CN_LIVE',
      'CN_SCHEDULED',
    ])
    expect(groups.archived.map((c) => c.id)).toEqual([
      'CN_ARCHIVED',
      'CN_ARCHIVED_DRAFT',
    ])
  })

  it('returns empty groups for an empty listing', () => {
    expect(groupConsentsForAdmin([], now)).toEqual({
      drafts: [],
      published: [],
      archived: [],
    })
  })
})

describe('consentManagementTypeLabel', () => {
  it('translates the known types', () => {
    expect(consentManagementTypeLabel('LOCAL')).toBe('Lokal')
    expect(consentManagementTypeLabel('REMOTE')).toBe('Ekstern')
  })

  it('falls back to the raw value for an unknown type', () => {
    expect(consentManagementTypeLabel('SOMETHING_NEW')).toBe('SOMETHING_NEW')
  })
})
