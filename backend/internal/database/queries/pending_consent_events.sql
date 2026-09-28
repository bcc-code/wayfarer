-- Pending consent events queries

-- name: CreatePendingConsentEvent :one
INSERT INTO pending_consent_events (id, members_id, consent_key, action, occurred_at, source)
VALUES (@id::text, @members_id::text, @consent_key::text, @action::text, @occurred_at, @source)
RETURNING id, members_id, consent_key, action, occurred_at, source, created_at;

-- name: GetPendingConsentEventsByMembersID :many
SELECT id, members_id, consent_key, action, occurred_at, source, created_at
FROM pending_consent_events
WHERE members_id = @members_id::text
ORDER BY occurred_at ASC;

-- name: DeletePendingConsentEventsByMembersID :exec
DELETE FROM pending_consent_events
WHERE members_id = @members_id::text;

-- name: DeletePendingConsentEvent :exec
DELETE FROM pending_consent_events
WHERE id = @id::char(28);
