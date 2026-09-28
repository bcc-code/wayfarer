-- name: CreateAuthSession :one
INSERT INTO auth_sessions (id, user_id, refresh_token_hash, expires_at, user_agent)
VALUES (@id::char(28), @user_id::char(28), @refresh_token_hash::bytea, @expires_at::timestamptz, sqlc.narg('user_agent')::text)
RETURNING *;

-- name: RotateAuthSession :one
-- Atomically swaps the current refresh-token hash for a new one and slides
-- the expiry. Returns no row when the token is unknown, revoked or expired.
UPDATE auth_sessions
SET prev_refresh_token_hash = refresh_token_hash,
    refresh_token_hash = @new_hash::bytea,
    rotated_at = now(),
    last_used_at = now(),
    expires_at = @expires_at::timestamptz,
    user_agent = COALESCE(sqlc.narg('user_agent')::text, user_agent)
WHERE refresh_token_hash = @current_hash::bytea
  AND revoked_at IS NULL
  AND expires_at > now()
RETURNING *;

-- name: GetAuthSessionByPrevHash :one
SELECT * FROM auth_sessions
WHERE prev_refresh_token_hash = @prev_hash::bytea;

-- name: RevokeAuthSession :exec
UPDATE auth_sessions
SET revoked_at = now()
WHERE id = @id::char(28) AND revoked_at IS NULL;

-- name: RevokeAuthSessionByHash :exec
UPDATE auth_sessions
SET revoked_at = now()
WHERE refresh_token_hash = @refresh_token_hash::bytea AND revoked_at IS NULL;

-- name: RevokeUserAuthSessions :execrows
UPDATE auth_sessions
SET revoked_at = now()
WHERE user_id = @user_id::char(28) AND revoked_at IS NULL;

-- name: DeleteStaleAuthSessions :execrows
-- Removes sessions that expired or were revoked before the cutoff.
DELETE FROM auth_sessions
WHERE expires_at < @cutoff::timestamptz
   OR revoked_at < @cutoff::timestamptz;
