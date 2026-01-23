-- +goose Up
-- Add PLUGIN challenge type and supporting columns

-- Add new columns for plugin challenges
ALTER TABLE challenges
ADD COLUMN plugin_challenge_id VARCHAR(100),
ADD COLUMN plugin_data JSONB;

-- Create unique index on plugin_challenge_id (partial, only non-null)
CREATE UNIQUE INDEX idx_challenges_plugin_challenge_id
ON challenges (plugin_challenge_id)
WHERE plugin_challenge_id IS NOT NULL;

-- Drop existing constraint and add updated one with PLUGIN type
ALTER TABLE challenges DROP CONSTRAINT challenge_type_url_constraint;
ALTER TABLE challenges DROP CONSTRAINT challenges_challenge_type_check;

-- Add updated challenge_type check constraint
ALTER TABLE challenges ADD CONSTRAINT challenges_challenge_type_check
CHECK (challenge_type IN ('SIMPLE', 'QUIZ', 'EXTERNAL', 'PLUGIN'));

-- Add updated type-specific constraint
ALTER TABLE challenges ADD CONSTRAINT challenge_type_url_constraint CHECK (
    (challenge_type = 'EXTERNAL' AND url IS NOT NULL AND url != '') OR
    (challenge_type = 'QUIZ' AND (url IS NULL OR url = '')) OR
    (challenge_type = 'SIMPLE') OR
    (challenge_type = 'PLUGIN' AND plugin_challenge_id IS NOT NULL AND plugin_challenge_id != '')
);

-- +goose Down
-- Remove PLUGIN challenge type and supporting columns

-- First delete any PLUGIN challenges (if any exist)
DELETE FROM challenges WHERE challenge_type = 'PLUGIN';

-- Drop constraints
ALTER TABLE challenges DROP CONSTRAINT challenge_type_url_constraint;
ALTER TABLE challenges DROP CONSTRAINT challenges_challenge_type_check;

-- Restore original constraints
ALTER TABLE challenges ADD CONSTRAINT challenges_challenge_type_check
CHECK (challenge_type IN ('SIMPLE', 'QUIZ', 'EXTERNAL'));

ALTER TABLE challenges ADD CONSTRAINT challenge_type_url_constraint CHECK (
    (challenge_type = 'EXTERNAL' AND url IS NOT NULL AND url != '') OR
    (challenge_type = 'QUIZ' AND (url IS NULL OR url = '')) OR
    (challenge_type = 'SIMPLE')
);

-- Drop index and columns
DROP INDEX idx_challenges_plugin_challenge_id;

ALTER TABLE challenges
DROP COLUMN plugin_challenge_id,
DROP COLUMN plugin_data;
