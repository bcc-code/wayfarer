-- +goose Up
-- +goose StatementBegin

-- Quiz sessions table (prefix: QN for quiz ruN, QS taken by submissions)
-- A session represents a specific run of a quiz for a group of users
CREATE TABLE quiz_sessions (
    id CHAR(28) PRIMARY KEY CHECK (id ~ '^QN[0-9A-Z]{26}$'),
    quiz_id CHAR(28) NOT NULL REFERENCES quizzes(id) ON DELETE CASCADE,
    name VARCHAR(255),

    -- State management (VARCHAR with CHECK, not enum)
    state VARCHAR(20) NOT NULL DEFAULT 'DRAFT' CHECK (state IN ('DRAFT', 'OPEN', 'LOCKED', 'FINISHED')),
    open_at TIMESTAMPTZ,
    lock_at TIMESTAMPTZ,
    finish_at TIMESTAMPTZ,

    -- Audit
    created_by CHAR(28) NOT NULL REFERENCES users(id),
    created_at TIMESTAMPTZ DEFAULT now() NOT NULL,
    updated_at TIMESTAMPTZ DEFAULT now() NOT NULL
);

CREATE INDEX idx_quiz_sessions_quiz ON quiz_sessions(quiz_id);
CREATE INDEX idx_quiz_sessions_state ON quiz_sessions(state);
CREATE INDEX idx_quiz_sessions_pending_open ON quiz_sessions(open_at) WHERE state = 'DRAFT' AND open_at IS NOT NULL;
CREATE INDEX idx_quiz_sessions_pending_lock ON quiz_sessions(lock_at) WHERE state = 'OPEN' AND lock_at IS NOT NULL;
CREATE INDEX idx_quiz_sessions_pending_finish ON quiz_sessions(finish_at) WHERE state = 'LOCKED' AND finish_at IS NOT NULL;

-- Session user access (prefix: QX for quiz access)
-- Tracks which users have access to a session
CREATE TABLE quiz_session_access (
    id CHAR(28) PRIMARY KEY CHECK (id ~ '^QX[0-9A-Z]{26}$'),
    session_id CHAR(28) NOT NULL REFERENCES quiz_sessions(id) ON DELETE CASCADE,
    user_id CHAR(28) NOT NULL REFERENCES users(id) ON DELETE CASCADE,

    granted_by CHAR(28) NOT NULL REFERENCES users(id),
    granted_at TIMESTAMPTZ DEFAULT now() NOT NULL,

    -- Source tracking for audit purposes
    source_type VARCHAR(20) NOT NULL CHECK (source_type IN ('DIRECT', 'TEAM', 'SUPER_TEAM', 'CHURCH', 'ALL')),
    source_id CHAR(28),

    UNIQUE (session_id, user_id)
);

CREATE INDEX idx_quiz_session_access_session ON quiz_session_access(session_id);
CREATE INDEX idx_quiz_session_access_user ON quiz_session_access(user_id);

-- Link submissions to sessions
ALTER TABLE quiz_submissions ADD COLUMN session_id CHAR(28) REFERENCES quiz_sessions(id);
ALTER TABLE quiz_submissions ADD COLUMN auto_submitted BOOLEAN DEFAULT false NOT NULL;
CREATE INDEX idx_quiz_submissions_session ON quiz_submissions(session_id);

-- Add webhook event type for quiz session finished
ALTER TABLE webhooks DROP CONSTRAINT IF EXISTS webhooks_event_type_check;
ALTER TABLE webhooks ADD CONSTRAINT webhooks_event_type_check
    CHECK (event_type IN ('external_content_event', 'points_awarded', 'quiz_session_finished'));

-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin

-- Remove webhook event type constraint and restore original
ALTER TABLE webhooks DROP CONSTRAINT IF EXISTS webhooks_event_type_check;
ALTER TABLE webhooks ADD CONSTRAINT webhooks_event_type_check
    CHECK (event_type IN ('external_content_event', 'points_awarded'));

-- Remove session columns from submissions
DROP INDEX IF EXISTS idx_quiz_submissions_session;
ALTER TABLE quiz_submissions DROP COLUMN IF EXISTS auto_submitted;
ALTER TABLE quiz_submissions DROP COLUMN IF EXISTS session_id;

-- Drop session tables
DROP TABLE IF EXISTS quiz_session_access;
DROP TABLE IF EXISTS quiz_sessions;

-- +goose StatementEnd
