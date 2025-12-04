-- +goose Up
-- +goose StatementBegin

-- Migration: Rename ssf_content to external_content and add source column
-- Generalizes the SSF-specific tables to support multiple external content sources

-- Rename tables
ALTER TABLE ssf_content RENAME TO external_content;
ALTER TABLE ssf_content_translations RENAME TO external_content_translations;

-- Add source column with default 'ssf'
ALTER TABLE external_content ADD COLUMN source TEXT NOT NULL DEFAULT 'ssf';

-- Update ID check constraint
ALTER TABLE external_content DROP CONSTRAINT ssf_content_id_check;
ALTER TABLE external_content ADD CONSTRAINT external_content_id_check CHECK (id ~ '^EC[0-9A-Z]{26}$');

-- Update foreign key constraint name
ALTER TABLE external_content_translations
    DROP CONSTRAINT ssf_content_translations_ssf_content_id_fkey;
ALTER TABLE external_content_translations
    ADD CONSTRAINT external_content_translations_external_content_id_fkey
    FOREIGN KEY (ssf_content_id) REFERENCES external_content(id) ON DELETE CASCADE;

-- Rename column in translations table
ALTER TABLE external_content_translations
    RENAME COLUMN ssf_content_id TO external_content_id;

-- Drop old indexes
DROP INDEX IF EXISTS idx_ssf_content_plan;
DROP INDEX IF EXISTS idx_ssf_content_task;
DROP INDEX IF EXISTS idx_ssf_content_content;
DROP INDEX IF EXISTS idx_ssf_content_type;
DROP INDEX IF EXISTS idx_ssf_content_published;

-- Create new indexes
CREATE INDEX idx_external_content_plan ON external_content(plan_id);
CREATE INDEX idx_external_content_task ON external_content(task_id);
CREATE INDEX idx_external_content_content ON external_content(content_id) WHERE content_id IS NOT NULL;
CREATE INDEX idx_external_content_type ON external_content(content_type);
CREATE INDEX idx_external_content_published ON external_content(published_at) WHERE published_at IS NOT NULL;
CREATE INDEX idx_external_content_source ON external_content(source);

-- Drop old triggers
DROP TRIGGER IF EXISTS ssf_content_updated_at ON external_content;
DROP TRIGGER IF EXISTS ssf_content_translations_updated_at ON external_content_translations;

-- Create new triggers
CREATE TRIGGER external_content_updated_at
    BEFORE UPDATE ON external_content
    FOR EACH ROW
    EXECUTE FUNCTION update_updated_at_column();

CREATE TRIGGER external_content_translations_updated_at
    BEFORE UPDATE ON external_content_translations
    FOR EACH ROW
    EXECUTE FUNCTION update_updated_at_column();

-- Update comments
COMMENT ON TABLE external_content IS 'Stores content items synced from external sources (SSF, etc.)';
COMMENT ON COLUMN external_content.source IS 'Content source identifier (e.g., ssf)';
COMMENT ON COLUMN external_content.plan_id IS 'External plan identifier';
COMMENT ON COLUMN external_content.task_id IS 'External item ID (unique within plan)';
COMMENT ON COLUMN external_content.content_id IS 'Nested content ID (MediaEpisode, Song, BookChapter, etc.)';
COMMENT ON COLUMN external_content.content_type IS 'Type of content: media_episode, song, book_chapter, periodical_article, bible_chapter, bible_verses';
COMMENT ON COLUMN external_content.published_at IS 'Content publication date';

-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin

-- Drop new triggers
DROP TRIGGER IF EXISTS external_content_updated_at ON external_content;
DROP TRIGGER IF EXISTS external_content_translations_updated_at ON external_content_translations;

-- Recreate old triggers
CREATE TRIGGER ssf_content_updated_at
    BEFORE UPDATE ON external_content
    FOR EACH ROW
    EXECUTE FUNCTION update_updated_at_column();

CREATE TRIGGER ssf_content_translations_updated_at
    BEFORE UPDATE ON external_content_translations
    FOR EACH ROW
    EXECUTE FUNCTION update_updated_at_column();

-- Drop new indexes
DROP INDEX IF EXISTS idx_external_content_plan;
DROP INDEX IF EXISTS idx_external_content_task;
DROP INDEX IF EXISTS idx_external_content_content;
DROP INDEX IF EXISTS idx_external_content_type;
DROP INDEX IF EXISTS idx_external_content_published;
DROP INDEX IF EXISTS idx_external_content_source;

-- Recreate old indexes
CREATE INDEX idx_ssf_content_plan ON external_content(plan_id);
CREATE INDEX idx_ssf_content_task ON external_content(task_id);
CREATE INDEX idx_ssf_content_content ON external_content(content_id) WHERE content_id IS NOT NULL;
CREATE INDEX idx_ssf_content_type ON external_content(content_type);
CREATE INDEX idx_ssf_content_published ON external_content(published_at) WHERE published_at IS NOT NULL;

-- Rename column back in translations table
ALTER TABLE external_content_translations
    RENAME COLUMN external_content_id TO ssf_content_id;

-- Update foreign key constraint name back
ALTER TABLE external_content_translations
    DROP CONSTRAINT external_content_translations_external_content_id_fkey;
ALTER TABLE external_content_translations
    ADD CONSTRAINT ssf_content_translations_ssf_content_id_fkey
    FOREIGN KEY (ssf_content_id) REFERENCES external_content(id) ON DELETE CASCADE;

-- Update ID check constraint back
ALTER TABLE external_content DROP CONSTRAINT external_content_id_check;
ALTER TABLE external_content ADD CONSTRAINT ssf_content_id_check CHECK (id ~ '^SC[0-9A-Z]{26}$');

-- Drop source column
ALTER TABLE external_content DROP COLUMN source;

-- Rename tables back
ALTER TABLE external_content RENAME TO ssf_content;
ALTER TABLE external_content_translations RENAME TO ssf_content_translations;

-- +goose StatementEnd
