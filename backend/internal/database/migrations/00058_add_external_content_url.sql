-- +goose Up
-- +goose StatementBegin

-- Migration: Add url column to external_content
-- Allows linking to external articles directly

ALTER TABLE external_content ADD COLUMN url TEXT;

-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin

ALTER TABLE external_content DROP COLUMN IF EXISTS url;

-- +goose StatementEnd
