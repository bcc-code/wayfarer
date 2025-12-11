-- +goose Up
-- +goose StatementBegin

-- Add new columns to achievements
ALTER TABLE achievements ADD COLUMN description_pending TEXT;
ALTER TABLE achievements ADD COLUMN description_completed TEXT;
ALTER TABLE achievements ADD COLUMN image_pending VARCHAR(500);
ALTER TABLE achievements ADD COLUMN image_completed VARCHAR(500);

-- Migrate existing data
UPDATE achievements SET
    description_pending = description,
    description_completed = description,
    image_pending = COALESCE(image_url, ''),
    image_completed = COALESCE(image_url, '');

-- Make columns NOT NULL after data migration
ALTER TABLE achievements ALTER COLUMN description_pending SET NOT NULL;
ALTER TABLE achievements ALTER COLUMN description_completed SET NOT NULL;
ALTER TABLE achievements ALTER COLUMN image_pending SET NOT NULL;
ALTER TABLE achievements ALTER COLUMN image_completed SET NOT NULL;

-- Drop old columns
ALTER TABLE achievements DROP COLUMN description;
ALTER TABLE achievements DROP COLUMN image_url;

-- Update achievement_translations table
ALTER TABLE achievement_translations ADD COLUMN description_pending TEXT;
ALTER TABLE achievement_translations ADD COLUMN description_completed TEXT;

UPDATE achievement_translations SET
    description_pending = description,
    description_completed = description;

ALTER TABLE achievement_translations DROP COLUMN description;

-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin

-- Restore achievement_translations
ALTER TABLE achievement_translations ADD COLUMN description TEXT;
UPDATE achievement_translations SET description = description_pending;
ALTER TABLE achievement_translations DROP COLUMN description_pending;
ALTER TABLE achievement_translations DROP COLUMN description_completed;

-- Restore achievements
ALTER TABLE achievements ADD COLUMN description TEXT;
ALTER TABLE achievements ADD COLUMN image_url VARCHAR(500);

UPDATE achievements SET
    description = description_pending,
    image_url = NULLIF(image_pending, '');

ALTER TABLE achievements ALTER COLUMN description SET NOT NULL;
ALTER TABLE achievements DROP COLUMN description_pending;
ALTER TABLE achievements DROP COLUMN description_completed;
ALTER TABLE achievements DROP COLUMN image_pending;
ALTER TABLE achievements DROP COLUMN image_completed;

-- +goose StatementEnd
