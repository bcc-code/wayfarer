-- +goose Up
-- Make logo_url nullable in projects table
ALTER TABLE projects ALTER COLUMN logo_url DROP NOT NULL;

-- +goose Down
-- Revert logo_url to NOT NULL (note: this will fail if there are NULL values)
ALTER TABLE projects ALTER COLUMN logo_url SET NOT NULL;
