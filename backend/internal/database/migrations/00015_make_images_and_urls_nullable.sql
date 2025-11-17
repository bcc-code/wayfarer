-- +goose Up
-- Make image and URL fields nullable across multiple tables
ALTER TABLE challenges ALTER COLUMN image_url DROP NOT NULL;
ALTER TABLE challenges ALTER COLUMN url DROP NOT NULL;
ALTER TABLE achievements ALTER COLUMN image_url DROP NOT NULL;
ALTER TABLE reading_achievement_articles ALTER COLUMN url DROP NOT NULL;

-- +goose Down
-- Revert changes (note: this will fail if NULL values exist)
ALTER TABLE challenges ALTER COLUMN image_url SET NOT NULL;
ALTER TABLE challenges ALTER COLUMN url SET NOT NULL;
ALTER TABLE achievements ALTER COLUMN image_url SET NOT NULL;
ALTER TABLE reading_achievement_articles ALTER COLUMN url SET NOT NULL;
