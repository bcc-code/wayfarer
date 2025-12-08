export type PendingConsent =
  SettingsPageQuery['me']['consentStatus']['pendingConsents'][number]
export type AcceptedConsent =
  SettingsPageQuery['me']['consentStatus']['acceptedConsents'][number]
export type RejectedConsent =
  SettingsPageQuery['me']['consentStatus']['rejectedConsents'][number]
