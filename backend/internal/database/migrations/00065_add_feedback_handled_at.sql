-- +goose Up
ALTER TABLE user_feedback ADD COLUMN handled_at TIMESTAMPTZ;

-- +goose Down
ALTER TABLE user_feedback DROP COLUMN IF EXISTS handled_at;
