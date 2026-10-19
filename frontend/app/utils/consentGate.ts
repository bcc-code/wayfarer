import { ConsentManagementType } from '~/api/generated'

/**
 * Whether the consent screen must be shown, decided in one place.
 *
 * Pure and structural rather than a composable, because route middleware runs
 * outside component setup and cannot call one — and because the two callers
 * read different queries, with different field selections, of the same status.
 */

type ConsentLike = {
  id: string
  isRequired: boolean
  managementType: ConsentManagementType
}

type ConsentStatusLike = {
  pendingConsents: ConsentLike[]
  rejectedConsents: { consent: ConsentLike }[]
}

/** Remote consents are accepted in the external system, never in this app. */
function isLocal(consent: ConsentLike) {
  return consent.managementType === ConsentManagementType.Local
}

/**
 * Consents that block the project: required, and not accepted.
 *
 * "Not accepted" rather than "pending" — a rejected required consent blocks
 * just as much as one never answered, which is the case plain `pendingConsents`
 * cannot express.
 *
 * Remote ones are excluded deliberately: they cannot be accepted in the app, so
 * gating on one would strand the user on a screen with no way forward. The API
 * is meant to refuse `isRequired` on a remote consent; until it does, this is
 * the half that keeps a misconfiguration from locking anyone out.
 */
export function blockingConsentIds(
  status: ConsentStatusLike | undefined,
): Set<string> {
  if (!status) return new Set()

  return new Set([
    ...status.pendingConsents
      .filter((c) => c.isRequired && isLocal(c))
      .map((c) => c.id),
    ...status.rejectedConsents
      .filter((entry) => entry.consent.isRequired && isLocal(entry.consent))
      .map((entry) => entry.consent.id),
  ])
}

/** Local consents the user has never answered — the soft onboarding nudge. */
export function unansweredLocalConsents(
  status: ConsentStatusLike | undefined,
): ConsentLike[] {
  if (!status) return []

  return status.pendingConsents.filter(isLocal)
}
