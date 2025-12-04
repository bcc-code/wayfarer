-- +goose Up
-- +goose StatementBegin

-- Migration: Add ssf_content table
-- Stores plan chapter items synced from SSF API

CREATE TABLE ssf_content (
    id CHAR(28) PRIMARY KEY CHECK (id ~ '^SC[0-9A-Z]{26}$'),
    plan_id TEXT NOT NULL,
    task_id TEXT NOT NULL,
    content_id TEXT,
    content_type TEXT NOT NULL,
    published_at TIMESTAMPTZ,
    synced_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    UNIQUE (plan_id, task_id)
);

-- Shadow table for translations (follows existing pattern)
CREATE TABLE ssf_content_translations (
    ssf_content_id CHAR(28) NOT NULL REFERENCES ssf_content(id) ON DELETE CASCADE,
    language_code VARCHAR(10) NOT NULL,
    title TEXT,
    created_at TIMESTAMPTZ DEFAULT now(),
    updated_at TIMESTAMPTZ DEFAULT now(),
    PRIMARY KEY (ssf_content_id, language_code)
);

CREATE INDEX idx_ssf_content_plan ON ssf_content(plan_id);
CREATE INDEX idx_ssf_content_task ON ssf_content(task_id);
CREATE INDEX idx_ssf_content_content ON ssf_content(content_id) WHERE content_id IS NOT NULL;
CREATE INDEX idx_ssf_content_type ON ssf_content(content_type);
CREATE INDEX idx_ssf_content_published ON ssf_content(published_at) WHERE published_at IS NOT NULL;

COMMENT ON TABLE ssf_content IS 'Stores plan chapter items synced from SSF API';
COMMENT ON COLUMN ssf_content.plan_id IS 'SSF Plan identifier';
COMMENT ON COLUMN ssf_content.task_id IS 'SSF PlanChapterItem ID (unique within plan)';
COMMENT ON COLUMN ssf_content.content_id IS 'Nested content ID (MediaEpisode, Song, BookChapter, etc.)';
COMMENT ON COLUMN ssf_content.content_type IS 'Type of content: media_episode, song, book_chapter, periodical_article, bible_chapter, bible_verses';
COMMENT ON COLUMN ssf_content.published_at IS 'Content publication date from SSF';

-- Add updated_at trigger
CREATE TRIGGER ssf_content_updated_at
    BEFORE UPDATE ON ssf_content
    FOR EACH ROW
    EXECUTE FUNCTION update_updated_at_column();

CREATE TRIGGER ssf_content_translations_updated_at
    BEFORE UPDATE ON ssf_content_translations
    FOR EACH ROW
    EXECUTE FUNCTION update_updated_at_column();

-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin

DROP TRIGGER IF EXISTS ssf_content_translations_updated_at ON ssf_content_translations;
DROP TRIGGER IF EXISTS ssf_content_updated_at ON ssf_content;
DROP TABLE IF EXISTS ssf_content_translations;
DROP TABLE IF EXISTS ssf_content;

-- +goose StatementEnd
