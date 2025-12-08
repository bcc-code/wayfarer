-- +goose Up
-- +goose StatementBegin

-- Add new timestamp columns to challenges table for state management
ALTER TABLE challenges
    ADD COLUMN visible_at TIMESTAMPTZ,
    ADD COLUMN started_at TIMESTAMPTZ,
    ADD COLUMN requires_team_membership BOOLEAN DEFAULT false NOT NULL,
    ADD COLUMN requires_super_team_membership BOOLEAN DEFAULT false NOT NULL;

-- Add index on visible_at for filtering queries
CREATE INDEX idx_challenges_visible ON challenges(visible_at) WHERE visible_at IS NOT NULL;

-- Create enrollment junction table following existing patterns
CREATE TABLE user_challenge_enrollments (
    user_id CHAR(28) NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    challenge_id CHAR(28) NOT NULL REFERENCES challenges(id) ON DELETE CASCADE,
    enrolled_at TIMESTAMPTZ DEFAULT now() NOT NULL,
    PRIMARY KEY (user_id, challenge_id)
);

CREATE INDEX idx_user_challenge_enrollments_user ON user_challenge_enrollments(user_id);
CREATE INDEX idx_user_challenge_enrollments_challenge ON user_challenge_enrollments(challenge_id);
CREATE INDEX idx_user_challenge_enrollments_time ON user_challenge_enrollments(enrolled_at);

-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin

-- Drop enrollment table
DROP TABLE IF EXISTS user_challenge_enrollments;

-- Drop index
DROP INDEX IF EXISTS idx_challenges_visible;

-- Remove columns from challenges
ALTER TABLE challenges
    DROP COLUMN IF EXISTS requires_super_team_membership,
    DROP COLUMN IF EXISTS requires_team_membership,
    DROP COLUMN IF EXISTS started_at,
    DROP COLUMN IF EXISTS visible_at;

-- +goose StatementEnd
