-- Consent queries

-- name: GetConsentByID :one
SELECT id, key, version, title, body, published_at, created_at, updated_at
FROM consents WHERE id = @id::text;

-- name: GetConsentsByIDs :many
SELECT id, key, version, title, body, published_at, created_at, updated_at
FROM consents WHERE id = ANY(@ids::text[]);

-- name: GetLatestPublishedConsentByKey :one
SELECT id, key, version, title, body, published_at, created_at, updated_at
FROM consents
WHERE key = @key::text AND published_at IS NOT NULL AND published_at <= now()
ORDER BY version DESC LIMIT 1;

-- name: GetAllLatestPublishedConsents :many
SELECT DISTINCT ON (key) id, key, version, title, body, published_at, created_at, updated_at
FROM consents
WHERE published_at IS NOT NULL AND published_at <= now()
ORDER BY key, version DESC;

-- name: CreateConsent :one
INSERT INTO consents (id, key, version, title, body, published_at)
VALUES (@id::text, @key::text, @version::int, @title::text, @body::text, @published_at)
RETURNING id, key, version, title, body, published_at, created_at, updated_at;

-- name: UpdateConsent :one
UPDATE consents SET
    title = CASE WHEN @title::text = '' THEN title ELSE @title::text END,
    body = CASE WHEN @body::text = '' THEN body ELSE @body::text END,
    published_at = @published_at,
    updated_at = now()
WHERE id = @id::text
RETURNING id, key, version, title, body, published_at, created_at, updated_at;

-- name: GetNextVersionForConsentKey :one
SELECT COALESCE(MAX(version), 0) + 1 as next_version FROM consents WHERE key = @key::text;

-- User consent queries

-- name: GetUserConsentsByUserID :many
SELECT uc.id, uc.user_id, uc.consent_id, uc.accepted_at, uc.created_at,
       c.key as consent_key, c.version as consent_version
FROM user_consents uc
INNER JOIN consents c ON uc.consent_id = c.id
WHERE uc.user_id = @user_id::text;

-- name: GetUserConsentsByUserIDs :many
SELECT uc.id, uc.user_id, uc.consent_id, uc.accepted_at, uc.created_at,
       c.key as consent_key, c.version as consent_version
FROM user_consents uc
INNER JOIN consents c ON uc.consent_id = c.id
WHERE uc.user_id = ANY(@user_ids::text[]);

-- name: GetUserConsentByUserAndConsent :one
SELECT id, user_id, consent_id, accepted_at, created_at
FROM user_consents WHERE user_id = @user_id::text AND consent_id = @consent_id::text;

-- name: CreateUserConsent :one
INSERT INTO user_consents (id, user_id, consent_id, accepted_at)
VALUES (@id::text, @user_id::text, @consent_id::text, @accepted_at)
RETURNING id, user_id, consent_id, accepted_at, created_at;

-- name: GetMissingConsentsForUser :many
SELECT c.id, c.key, c.version, c.title, c.body, c.published_at, c.created_at, c.updated_at
FROM (
    SELECT DISTINCT ON (key) id, key, version, title, body, published_at, created_at, updated_at
    FROM consents
    WHERE published_at IS NOT NULL AND published_at <= now()
    ORDER BY key, version DESC
) c
WHERE NOT EXISTS (
    SELECT 1 FROM user_consents uc WHERE uc.user_id = @user_id::text AND uc.consent_id = c.id
);

-- Translation queries

-- name: GetConsentTranslationsByIDs :many
SELECT consent_id, language_code, title, body
FROM consent_translations
WHERE consent_id = ANY(@entity_ids::text[])
  AND language_code = @language_code::text;

-- name: DeleteConsentTranslations :exec
DELETE FROM consent_translations WHERE consent_id = @consent_id::text;

-- name: UpsertConsentTranslation :one
INSERT INTO consent_translations (consent_id, language_code, title, body)
VALUES (@consent_id::text, @language_code::text, @title, @body)
ON CONFLICT (consent_id, language_code) DO UPDATE SET
    title = EXCLUDED.title,
    body = EXCLUDED.body,
    updated_at = now()
RETURNING consent_id, language_code, title, body, created_at, updated_at;
