export type PendingConsent =
  ConsentsPageQuery['me']['consentStatus']['pendingConsents'][number]
export type AcceptedConsent =
  ConsentsPageQuery['me']['consentStatus']['acceptedConsents'][number]
export type RejectedConsent =
  ConsentsPageQuery['me']['consentStatus']['rejectedConsents'][number]
