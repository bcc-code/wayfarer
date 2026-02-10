-- +goose Up
ALTER TABLE challenges ADD COLUMN notification_text TEXT NOT NULL DEFAULT '';
ALTER TABLE challenge_translations ADD COLUMN notification_text TEXT;

-- +goose Down
ALTER TABLE challenges DROP COLUMN notification_text;
ALTER TABLE challenge_translations DROP COLUMN notification_text;
