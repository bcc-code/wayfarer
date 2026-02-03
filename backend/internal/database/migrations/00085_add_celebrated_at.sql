-- +goose Up
ALTER TABLE user_achievements ADD COLUMN celebrated_at TIMESTAMPTZ;

-- +goose Down
ALTER TABLE user_achievements DROP COLUMN celebrated_at;
