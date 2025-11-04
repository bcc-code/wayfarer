-- +goose Up
-- +goose StatementBegin

-- Add rounding column to projects table for border-radius styling
ALTER TABLE projects ADD COLUMN rounding INT NOT NULL DEFAULT 0;

-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin

-- Remove rounding column from projects table
ALTER TABLE projects DROP COLUMN rounding;

-- +goose StatementEnd
