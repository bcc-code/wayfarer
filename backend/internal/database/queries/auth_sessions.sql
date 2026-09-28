-- name: CreateAuthSession :one
INSERT INTO auth_sessions (id, user_id, refresh_token_hash, expires_at, user_agent)
VALUES (@id::char(28), @user_id::char(28), @refresh_token_hash::bytea, @expires_at::timestamptz, sqlc.narg('user_agent')::text)
RETURNING *;

-- name: GetAuthSessionByID :one
SELECT * FROM auth_sessions
WHERE id = @id::char(28);

-- name: RotateAuthSession :one
-- Compare-and-swap: replaces the refresh-token hash only if it is still the
-- one the caller validated, slides the expiry, and records the replaced hash
-- as retired, all in one statement. Returns no row when a concurrent
-- request rotated or revoked the session first.
WITH rotated AS (
    UPDATE auth_sessions
    SET prev_refresh_token_hash = refresh_token_hash,
        refresh_token_hash = @new_hash::bytea,
        rotated_at = now(),
        last_used_at = now(),
        expires_at = @expires_at::timestamptz,
        user_agent = COALESCE(sqlc.narg('user_agent')::text, user_agent)
    WHERE auth_sessions.id = @id::char(28)
      AND auth_sessions.refresh_token_hash = @current_hash::bytea
      AND auth_sessions.revoked_at IS NULL
    RETURNING auth_sessions.id, auth_sessions.prev_refresh_token_hash
)
INSERT INTO auth_session_retired_tokens (token_hash, session_id)
SELECT rotated.prev_refresh_token_hash, rotated.id FROM rotated
RETURNING session_id;

-- name: IsRetiredAuthSessionToken :one
-- Whether a hash was once a refresh token of this session.
SELECT EXISTS (
    SELECT 1 FROM auth_session_retired_tokens
    WHERE token_hash = @token_hash::bytea
      AND session_id = @session_id::char(28)
);

-- name: RevokeAuthSession :exec
UPDATE auth_sessions
SET revoked_at = now()
WHERE id = @id::char(28) AND revoked_at IS NULL;

-- name: RevokeUserAuthSessions :execrows
UPDATE auth_sessions
SET revoked_at = now()
WHERE user_id = @user_id::char(28) AND revoked_at IS NULL;

-- name: DeleteStaleAuthSessions :execrows
-- Removes sessions that expired or were revoked before the cutoff.
DELETE FROM auth_sessions
WHERE expires_at < @cutoff::timestamptz
   OR revoked_at < @cutoff::timestamptz;
