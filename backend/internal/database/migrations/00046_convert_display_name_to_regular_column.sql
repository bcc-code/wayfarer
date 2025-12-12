-- +goose Up
-- Drop the index first
DROP INDEX IF EXISTS idx_users_display_name;

-- Temporarily store the computed values
ALTER TABLE users ADD COLUMN display_name_temp VARCHAR(255);

UPDATE users SET display_name_temp =
    CASE
        WHEN first_name IS NOT NULL AND first_name != ''
             AND last_name IS NOT NULL AND last_name != ''
        THEN first_name || ' ' || LEFT(last_name, 1) || '.'
        ELSE name
    END;

-- Drop the generated column
ALTER TABLE users DROP COLUMN display_name;

-- Rename temp to display_name
ALTER TABLE users RENAME COLUMN display_name_temp TO display_name;

-- Recreate the index
CREATE INDEX idx_users_display_name ON users(display_name);

-- +goose Down
DROP INDEX IF EXISTS idx_users_display_name;
ALTER TABLE users DROP COLUMN display_name;
ALTER TABLE users ADD COLUMN display_name VARCHAR(255) GENERATED ALWAYS AS (
    CASE
        WHEN first_name IS NOT NULL AND first_name != ''
             AND last_name IS NOT NULL AND last_name != ''
        THEN first_name || ' ' || LEFT(last_name, 1) || '.'
        ELSE name
    END
) STORED;
CREATE INDEX idx_users_display_name ON users(display_name);
