-- +goose Up
-- +goose StatementBegin

-- Migration: Add complete_by column to external_content
-- Used for SSF sync to track when content should be completed (based on completion_mode)

ALTER TABLE external_content ADD COLUMN complete_by TIMESTAMPTZ;

CREATE INDEX idx_external_content_complete_by ON external_content(complete_by) WHERE complete_by IS NOT NULL;

COMMENT ON COLUMN external_content.complete_by IS 'Deadline for completing the content (calculated from completion_mode during SSF sync)';

-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin

DROP INDEX IF EXISTS idx_external_content_complete_by;
ALTER TABLE external_content DROP COLUMN IF EXISTS complete_by;

-- +goose StatementEnd
