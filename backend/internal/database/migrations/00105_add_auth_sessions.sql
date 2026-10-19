-- +goose Up
-- +goose StatementBegin

-- Wayfarer-owned login sessions. One row per signed-in device. The refresh
-- token embeds the session id and is never stored, only its SHA-256 hash.
-- Each refresh rotates the hash and slides expires_at forward.
CREATE TABLE auth_sessions (
    id CHAR(28) PRIMARY KEY CHECK (id ~ '^AS[0-9A-Z]{26}$'),
    user_id CHAR(28) NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    refresh_token_hash BYTEA NOT NULL,
    prev_refresh_token_hash BYTEA,
    rotated_at TIMESTAMPTZ,
    expires_at TIMESTAMPTZ NOT NULL,
    last_used_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    revoked_at TIMESTAMPTZ,
    user_agent TEXT,
    created_at TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE INDEX idx_auth_sessions_user ON auth_sessions(user_id);
CREATE INDEX idx_auth_sessions_expires ON auth_sessions(expires_at);

-- Hashes of refresh tokens a session has rotated out. Presenting one of
-- these proves the token was really issued (it is not just a guessed
-- session id), so the session is revoked as a token-theft signal.
CREATE TABLE auth_session_retired_tokens (
    token_hash BYTEA PRIMARY KEY,
    session_id CHAR(28) NOT NULL REFERENCES auth_sessions(id) ON DELETE CASCADE,
    retired_at TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE INDEX idx_auth_session_retired_tokens_session ON auth_session_retired_tokens(session_id);

-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin

DROP TABLE IF EXISTS auth_session_retired_tokens;
DROP TABLE IF EXISTS auth_sessions;

-- +goose StatementEnd
