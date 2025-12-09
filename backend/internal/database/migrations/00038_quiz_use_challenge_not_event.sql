-- +goose Up
-- Replace event_id with challenge_id for quizzes
-- Quizzes are always associated with a challenge, never directly with an event

-- Delete existing quizzes (pre-production cleanup)
DELETE FROM quizzes;

-- Drop the old event_id column and constraint
ALTER TABLE quizzes DROP CONSTRAINT IF EXISTS quizzes_event_id_fkey;
ALTER TABLE quizzes DROP COLUMN IF EXISTS event_id;

-- Drop old index if exists
DROP INDEX IF EXISTS idx_quizzes_event;

-- Add challenge_id column (required)
ALTER TABLE quizzes ADD COLUMN challenge_id CHAR(28) NOT NULL REFERENCES challenges(id) ON DELETE CASCADE;

-- Create index for challenge_id
CREATE INDEX idx_quizzes_challenge ON quizzes(challenge_id);

-- +goose Down
-- Revert to event_id

-- Drop challenge_id
DROP INDEX IF EXISTS idx_quizzes_challenge;
ALTER TABLE quizzes DROP CONSTRAINT IF EXISTS quizzes_challenge_id_fkey;
ALTER TABLE quizzes DROP COLUMN IF EXISTS challenge_id;

-- Add event_id back (nullable)
ALTER TABLE quizzes ADD COLUMN event_id CHAR(28) REFERENCES events(id) ON DELETE SET NULL;
CREATE INDEX idx_quizzes_event ON quizzes(event_id);
