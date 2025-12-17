-- Consent queries

-- name: GetConsentByID :one
SELECT id, key, version, title, short_text, body, url, published_at, is_remote, managed_by, created_at, updated_at
FROM consents WHERE id = @id::text;

-- name: GetConsentsByIDs :many
SELECT id, key, version, title, short_text, body, url, published_at, is_remote, managed_by, created_at, updated_at
FROM consents WHERE id = ANY(@ids::text[]);

-- name: GetLatestPublishedConsentByKey :one
SELECT id, key, version, title, short_text, body, url, published_at, is_remote, managed_by, created_at, updated_at
FROM consents
WHERE key = @key::text AND published_at IS NOT NULL AND published_at <= now()
ORDER BY version DESC LIMIT 1;

-- name: GetAllLatestPublishedConsents :many
SELECT DISTINCT ON (key) id, key, version, title, short_text, body, url, published_at, is_remote, managed_by, created_at, updated_at
FROM consents
WHERE published_at IS NOT NULL AND published_at <= now()
ORDER BY key, version DESC;

-- name: GetLatestConsentByKey :one
SELECT id, key, version, title, short_text, body, url, published_at, is_remote, managed_by, created_at, updated_at
FROM consents
WHERE key = @key::text
ORDER BY version DESC
LIMIT 1;

-- name: CreateConsent :one
INSERT INTO consents (id, key, version, title, short_text, body, url, published_at, is_remote, managed_by)
VALUES (@id::text, @key::text, @version::int, @title::text, @short_text::text, @body::text, @url, @published_at, @is_remote::bool, @managed_by)
RETURNING id, key, version, title, short_text, body, url, published_at, is_remote, managed_by, created_at, updated_at;

-- name: UpdateConsent :one
UPDATE consents SET
    title = CASE WHEN @title::text = '' THEN title ELSE @title::text END,
    short_text = CASE WHEN @short_text::text = '' THEN short_text ELSE @short_text::text END,
    body = CASE WHEN @body::text = '' THEN body ELSE @body::text END,
    url = CASE WHEN @url::text = '' THEN url ELSE @url::text END,
    managed_by = CASE WHEN @managed_by::text = '' THEN managed_by ELSE @managed_by END,
    published_at = @published_at,
    updated_at = now()
WHERE id = @id::text
RETURNING id, key, version, title, short_text, body, url, published_at, is_remote, managed_by, created_at, updated_at;

-- name: GetNextVersionForConsentKey :one
SELECT COALESCE(MAX(version), 0) + 1 as next_version FROM consents WHERE key = @key::text;

-- User consent history queries

-- name: CreateUserConsentHistory :one
INSERT INTO user_consent_history (
    id, user_id, consent_id, consent_key, action, occurred_at,
    source, external_consent_id, external_timestamp
)
VALUES (
    @id::text, @user_id::text, @consent_id::text, @consent_key::text,
    @action::text, @occurred_at, @source, @external_consent_id, @external_timestamp
)
RETURNING id, user_id, consent_id, consent_key, action, occurred_at,
          source, external_consent_id, external_timestamp;

-- name: GetLatestUserConsentActionByKey :one
SELECT id, user_id, consent_id, consent_key, action, occurred_at,
       source, external_consent_id, external_timestamp
FROM user_consent_history
WHERE user_id = @user_id::text AND consent_key = @consent_key::text
ORDER BY occurred_at DESC
LIMIT 1;

-- name: GetCurrentUserConsentStatusesByUsers :many
-- Gets the latest action for each consent key for multiple users
SELECT DISTINCT ON (user_id, consent_key)
    id, user_id, consent_id, consent_key, action, occurred_at, source
FROM user_consent_history
WHERE user_id = ANY(@user_ids::text[])
ORDER BY user_id, consent_key, occurred_at DESC;

-- name: GetUserConsentHistoryByUserAndKey :many
SELECT id, user_id, consent_id, consent_key, action, occurred_at,
       source, external_consent_id, external_timestamp
FROM user_consent_history
WHERE user_id = @user_id::text AND consent_key = @consent_key::text
ORDER BY occurred_at DESC;

-- name: GetUserConsentHistoryByUser :many
SELECT id, user_id, consent_id, consent_key, action, occurred_at,
       source, external_consent_id, external_timestamp
FROM user_consent_history
WHERE user_id = @user_id::text
ORDER BY occurred_at DESC;

-- name: GetUserConsentHistoryByUsers :many
SELECT id, user_id, consent_id, consent_key, action, occurred_at,
       source, external_consent_id, external_timestamp
FROM user_consent_history
WHERE user_id = ANY(@user_ids::text[])
ORDER BY user_id, occurred_at DESC;

-- name: GetMissingConsentsForUserWithRejections :many
-- Gets consents that user has never acted upon (no history)
SELECT c.id, c.key, c.version, c.title, c.short_text, c.body, c.url, c.published_at,
       c.is_remote, c.managed_by, c.created_at, c.updated_at
FROM (
    SELECT DISTINCT ON (key) id, key, version, title, short_text, body, url, published_at,
           is_remote, managed_by, created_at, updated_at
    FROM consents
    WHERE published_at IS NOT NULL AND published_at <= now()
    ORDER BY key, version DESC
) c
WHERE NOT EXISTS (
    SELECT 1 FROM user_consent_history uch
    WHERE uch.user_id = @user_id::text
    AND uch.consent_key = c.key
);

-- Translation queries

-- name: GetConsentTranslationsByIDs :many
SELECT consent_id, language_code, title, short_text, body
FROM consent_translations
WHERE consent_id = ANY(@entity_ids::text[])
  AND language_code = @language_code::text;

-- name: DeleteConsentTranslations :exec
DELETE FROM consent_translations WHERE consent_id = @consent_id::text;

-- name: UpsertConsentTranslation :one
INSERT INTO consent_translations (consent_id, language_code, title, short_text, body)
VALUES (@consent_id::text, @language_code::text, @title, @short_text, @body)
ON CONFLICT (consent_id, language_code) DO UPDATE SET
    title = EXCLUDED.title,
    short_text = EXCLUDED.short_text,
    body = EXCLUDED.body,
    updated_at = now()
RETURNING consent_id, language_code, title, short_text, body, created_at, updated_at;
