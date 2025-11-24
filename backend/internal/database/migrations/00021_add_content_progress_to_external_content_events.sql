-- +goose Up
-- +goose StatementBegin

-- Add content_progress column to external_content_events table
ALTER TABLE external_content_events ADD COLUMN content_progress REAL;

COMMENT ON COLUMN external_content_events.content_progress IS 'Content completion progress (0.01 to 1.1, where 1.0 = 100%)';

-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin

-- Remove content_progress column from external_content_events table
ALTER TABLE external_content_events DROP COLUMN content_progress;

-- +goose StatementEnd
