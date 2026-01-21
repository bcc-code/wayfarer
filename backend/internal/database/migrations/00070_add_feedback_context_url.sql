-- +goose Up
ALTER TABLE user_feedback ADD COLUMN context_url TEXT;

-- +goose Down
ALTER TABLE user_feedback DROP COLUMN context_url;
