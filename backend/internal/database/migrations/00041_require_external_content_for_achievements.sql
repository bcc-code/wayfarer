-- +goose Up
-- +goose StatementBegin

-- Migration: Simplify reading/listening achievement content to always use external_content FK
-- Remove local metadata fields and require external_content_id

-- Remove CHECK constraints that allowed local metadata
ALTER TABLE reading_achievement_articles DROP CONSTRAINT IF EXISTS reading_article_source_check;
ALTER TABLE listening_achievement_tracks DROP CONSTRAINT IF EXISTS listening_track_source_check;

-- Make external_content_id NOT NULL
ALTER TABLE reading_achievement_articles ALTER COLUMN external_content_id SET NOT NULL;
ALTER TABLE listening_achievement_tracks ALTER COLUMN external_content_id SET NOT NULL;

-- Drop local metadata columns from reading_achievement_articles
ALTER TABLE reading_achievement_articles DROP COLUMN IF EXISTS article_id;
ALTER TABLE reading_achievement_articles DROP COLUMN IF EXISTS title;
ALTER TABLE reading_achievement_articles DROP COLUMN IF EXISTS author;
ALTER TABLE reading_achievement_articles DROP COLUMN IF EXISTS url;

-- Drop local metadata columns from listening_achievement_tracks
ALTER TABLE listening_achievement_tracks DROP COLUMN IF EXISTS track_id;
ALTER TABLE listening_achievement_tracks DROP COLUMN IF EXISTS name;
ALTER TABLE listening_achievement_tracks DROP COLUMN IF EXISTS description;
ALTER TABLE listening_achievement_tracks DROP COLUMN IF EXISTS image_url;

-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin

-- Re-add local metadata columns to reading_achievement_articles
ALTER TABLE reading_achievement_articles ADD COLUMN article_id VARCHAR(255);
ALTER TABLE reading_achievement_articles ADD COLUMN title VARCHAR(500);
ALTER TABLE reading_achievement_articles ADD COLUMN author VARCHAR(255);
ALTER TABLE reading_achievement_articles ADD COLUMN url VARCHAR(500);

-- Re-add local metadata columns to listening_achievement_tracks
ALTER TABLE listening_achievement_tracks ADD COLUMN track_id VARCHAR(255);
ALTER TABLE listening_achievement_tracks ADD COLUMN name VARCHAR(500);
ALTER TABLE listening_achievement_tracks ADD COLUMN description TEXT;
ALTER TABLE listening_achievement_tracks ADD COLUMN image_url VARCHAR(500);

-- Make external_content_id nullable again
ALTER TABLE reading_achievement_articles ALTER COLUMN external_content_id DROP NOT NULL;
ALTER TABLE listening_achievement_tracks ALTER COLUMN external_content_id DROP NOT NULL;

-- Re-add CHECK constraints
ALTER TABLE reading_achievement_articles
    ADD CONSTRAINT reading_article_source_check
    CHECK (
        (external_content_id IS NOT NULL) OR
        (article_id IS NOT NULL AND title IS NOT NULL AND author IS NOT NULL)
    );

ALTER TABLE listening_achievement_tracks
    ADD CONSTRAINT listening_track_source_check
    CHECK (
        (external_content_id IS NOT NULL) OR
        (track_id IS NOT NULL AND name IS NOT NULL)
    );

-- +goose StatementEnd
