-- +goose Up
ALTER TABLE user_feedback ADD COLUMN tags TEXT[] DEFAULT '{}';

-- +goose Down
ALTER TABLE user_feedback DROP COLUMN tags;
