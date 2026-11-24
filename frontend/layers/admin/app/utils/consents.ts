/** Where a consent stands, as `publishedAt` and `archivedAt` encode it. */
export type ConsentPublicationStatus =
  'archived' | 'published' | 'scheduled' | 'draft'

export interface ConsentStatusFields {
  publishedAt?: string | null
  archivedAt?: string | null
}

export function consentPublicationStatus(
  consent: ConsentStatusFields,
  now: Date = new Date(),
): ConsentPublicationStatus {
  if (consent.archivedAt) return 'archived'
  if (!consent.publishedAt) return 'draft'
  return new Date(consent.publishedAt) > now ? 'scheduled' : 'published'
}

export const CONSENT_STATUS_LABELS: Record<ConsentPublicationStatus, string> = {
  archived: 'Arkivert',
  published: 'Publisert',
  scheduled: 'Planlagt',
  draft: 'Utkast',
}

export const CONSENT_STATUS_COLORS: Record<
  ConsentPublicationStatus,
  'success' | 'info' | 'warning' | 'neutral'
> = {
  archived: 'neutral',
  published: 'success',
  scheduled: 'info',
  draft: 'warning',
}

export const CONSENT_MANAGEMENT_TYPE_LABELS: Record<string, string> = {
  LOCAL: 'Lokal',
  REMOTE: 'Ekstern',
}

export function consentManagementTypeLabel(managementType: string): string {
  return CONSENT_MANAGEMENT_TYPE_LABELS[managementType] ?? managementType
}

export function consentManagementTypeColor(
  managementType: string,
): 'primary' | 'neutral' {
  return managementType === 'LOCAL' ? 'primary' : 'neutral'
}

/** The date the status refers to; a draft has neither. */
export function consentStatusDate(consent: ConsentStatusFields): string | null {
  return consent.archivedAt ?? consent.publishedAt ?? null
}

/**
 * The row shape the consent tables render — declared here, not taken from one
 * generated query type, so both listing pages can feed the same table.
 */
export interface AdminConsentRow {
  id: string
  key: string
  version: number
  title: string
  publishedAt?: string | null
  archivedAt?: string | null
  managementType: string
  isRequired: boolean
  project?: { id: string; name: string } | null
}

export interface GroupedConsents {
  drafts: AdminConsentRow[]
  published: AdminConsentRow[]
  archived: AdminConsentRow[]
}

/**
 * Splits a listing into the page's three tables. Scheduled versions count as
 * published: committed, just not live yet, and their badge says so.
 */
export function groupConsentsForAdmin(
  consents: AdminConsentRow[],
  now: Date = new Date(),
): GroupedConsents {
  const groups: GroupedConsents = { drafts: [], published: [], archived: [] }

  for (const consent of consents) {
    const status = consentPublicationStatus(consent, now)
    if (status === 'archived') groups.archived.push(consent)
    else if (status === 'draft') groups.drafts.push(consent)
    else groups.published.push(consent)
  }

  return groups
}

export const CONSENT_ARCHIVE_DESCRIPTION =
  'Samtykket tas ut av bruk, og brukere blir ikke lenger spurt om det. ' +
  'Svarene som allerede er gitt beholdes.'
