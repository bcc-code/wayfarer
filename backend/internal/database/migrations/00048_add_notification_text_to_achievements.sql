-- +goose Up
ALTER TABLE achievements ADD COLUMN notification_text TEXT NOT NULL DEFAULT '';
ALTER TABLE achievement_translations ADD COLUMN notification_text TEXT;

-- +goose Down
ALTER TABLE achievements DROP COLUMN notification_text;
ALTER TABLE achievement_translations DROP COLUMN notification_text;
