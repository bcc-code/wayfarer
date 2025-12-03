-- +goose Up
-- Add the new columns first
ALTER TABLE users
ADD COLUMN first_name VARCHAR(255),
ADD COLUMN last_name VARCHAR(255),
ADD COLUMN middle_name VARCHAR(255);

-- Populate first_name and last_name by splitting the name field
-- Assumes "FirstName LastName" format (space-separated)
UPDATE users
SET
    first_name = SPLIT_PART(name, ' ', 1),
    last_name = CASE
        WHEN SPLIT_PART(name, ' ', 2) != ''
        THEN SPLIT_PART(name, ' ', 2)
        ELSE NULL
    END
WHERE name IS NOT NULL AND name != '';

-- Now add the generated column
ALTER TABLE users
ADD COLUMN display_name VARCHAR(255) GENERATED ALWAYS AS (
    CASE
        WHEN first_name IS NOT NULL AND first_name != ''
             AND last_name IS NOT NULL AND last_name != ''
        THEN first_name || ' ' || LEFT(last_name, 1) || '.'
        ELSE name
    END
) STORED;

-- Add indexes for common queries
CREATE INDEX idx_users_first_name ON users(first_name);
CREATE INDEX idx_users_last_name ON users(last_name);
CREATE INDEX idx_users_display_name ON users(display_name);

-- +goose Down
DROP INDEX IF EXISTS idx_users_display_name;
DROP INDEX IF EXISTS idx_users_last_name;
DROP INDEX IF EXISTS idx_users_first_name;

ALTER TABLE users
DROP COLUMN display_name,
DROP COLUMN middle_name,
DROP COLUMN last_name,
DROP COLUMN first_name;
