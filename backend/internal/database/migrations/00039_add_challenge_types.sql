-- +goose Up
-- +goose StatementBegin

-- Add challenge_type column to distinguish between challenge types
ALTER TABLE challenges
    ADD COLUMN challenge_type VARCHAR(50) NOT NULL DEFAULT 'SIMPLE'
    CHECK (challenge_type IN ('SIMPLE', 'QUIZ', 'EXTERNAL'));

-- Add allow_self_completion for SimpleChallenge type
ALTER TABLE challenges
    ADD COLUMN allow_self_completion BOOLEAN DEFAULT true NOT NULL;

-- Add index on challenge_type for filtering queries
CREATE INDEX idx_challenges_type ON challenges(challenge_type);

-- Add constraint to enforce type-specific rules:
-- EXTERNAL challenges MUST have a URL
-- QUIZ challenges MUST NOT have a URL (quiz is linked separately)
-- SIMPLE challenges have no URL requirement
ALTER TABLE challenges
    ADD CONSTRAINT challenge_type_url_constraint CHECK (
        (challenge_type = 'EXTERNAL' AND url IS NOT NULL AND url != '') OR
        (challenge_type = 'QUIZ' AND (url IS NULL OR url = '')) OR
        (challenge_type = 'SIMPLE')
    );

-- Update existing challenges that have a quiz linked to them to be QUIZ type
UPDATE challenges c
SET challenge_type = 'QUIZ', url = NULL
WHERE EXISTS (SELECT 1 FROM quizzes q WHERE q.challenge_id = c.id);

-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin

-- Drop constraint first
ALTER TABLE challenges DROP CONSTRAINT IF EXISTS challenge_type_url_constraint;

-- Drop index
DROP INDEX IF EXISTS idx_challenges_type;

-- Remove columns
ALTER TABLE challenges
    DROP COLUMN IF EXISTS allow_self_completion,
    DROP COLUMN IF EXISTS challenge_type;

-- +goose StatementEnd
