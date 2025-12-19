-- +goose Up
-- +goose StatementBegin

-- Add person_uuid column to users table
-- This is the new UUID-based identifier from BCC Members
ALTER TABLE users ADD COLUMN person_uuid UUID;

-- Create unique partial index for person_uuid lookups
-- Partial index allows NULL values while enforcing uniqueness for non-NULL
CREATE UNIQUE INDEX idx_users_person_uuid ON users(person_uuid) WHERE person_uuid IS NOT NULL;

-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin

DROP INDEX IF EXISTS idx_users_person_uuid;
ALTER TABLE users DROP COLUMN person_uuid;

-- +goose StatementEnd
