-- +goose Up
ALTER TABLE quiz_questions ADD COLUMN IF NOT EXISTS image_url VARCHAR(500);

-- +goose Down
ALTER TABLE quiz_questions DROP COLUMN IF EXISTS image_url;
