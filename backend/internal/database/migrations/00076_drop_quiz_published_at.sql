-- +goose Up
-- Drop quiz published_at column
-- Visibility is now controlled solely through challenge.published_at
ALTER TABLE quizzes DROP COLUMN IF EXISTS published_at;

-- +goose Down
-- Re-add published_at column
ALTER TABLE quizzes ADD COLUMN published_at TIMESTAMPTZ;
