-- +goose Up
-- +goose StatementBegin

-- Add CHECK constraint to ensure reasonable birthdates
-- Birthdate must be between 1900-01-01 and today
ALTER TABLE users ADD CONSTRAINT chk_users_birthdate_range
    CHECK (birthdate >= '1900-01-01' AND birthdate <= CURRENT_DATE);

-- Make birthdate column NOT NULL
-- NOTE: Run internal/database/migrations/scripts/backfill_birthdates.sql BEFORE this migration
-- if there are any users with NULL birthdate
ALTER TABLE users ALTER COLUMN birthdate SET NOT NULL;

-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin

-- Make birthdate nullable again
ALTER TABLE users ALTER COLUMN birthdate DROP NOT NULL;

-- Drop CHECK constraint
ALTER TABLE users DROP CONSTRAINT chk_users_birthdate_range;

-- +goose StatementEnd
