-- +goose Up
-- +goose StatementBegin

-- Add external_id column to churches table for mapping to external church IDs
ALTER TABLE churches ADD COLUMN external_id INT UNIQUE;

-- Update users age constraint to allow 0 as placeholder
ALTER TABLE users DROP CONSTRAINT users_age_check;
ALTER TABLE users ADD CONSTRAINT users_age_check CHECK (age >= 0);

-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin

-- Revert users age constraint
ALTER TABLE users DROP CONSTRAINT users_age_check;
ALTER TABLE users ADD CONSTRAINT users_age_check CHECK (age > 0);

-- Remove external_id column from churches
ALTER TABLE churches DROP COLUMN external_id;

-- +goose StatementEnd
