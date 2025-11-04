-- +goose Up
-- +goose StatementBegin

-- Drop the age index
DROP INDEX IF EXISTS idx_users_age;

-- Remove the age column
ALTER TABLE users DROP COLUMN age;

-- Add birthdate column (nullable for existing users)
ALTER TABLE users ADD COLUMN birthdate DATE;

-- Add index on birthdate
CREATE INDEX idx_users_birthdate ON users(birthdate);

-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin

-- Drop birthdate index
DROP INDEX IF EXISTS idx_users_birthdate;

-- Remove birthdate column
ALTER TABLE users DROP COLUMN birthdate;

-- Add age column back
ALTER TABLE users ADD COLUMN age INT NOT NULL DEFAULT 0 CHECK (age >= 0);

-- Recreate age index
CREATE INDEX idx_users_age ON users(age);

-- +goose StatementEnd
