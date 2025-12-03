-- +goose Up
-- +goose StatementBegin

-- Rename columns in external_content_events table
ALTER TABLE external_content_events RENAME COLUMN content_id TO task_id;
ALTER TABLE external_content_events RENAME COLUMN reading_plan_id TO plan_id;

-- Add consumed_at timestamp column
ALTER TABLE external_content_events ADD COLUMN consumed_at TIMESTAMPTZ;

-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin

-- Remove consumed_at column
ALTER TABLE external_content_events DROP COLUMN consumed_at;

-- Revert column renames
ALTER TABLE external_content_events RENAME COLUMN task_id TO content_id;
ALTER TABLE external_content_events RENAME COLUMN plan_id TO reading_plan_id;

-- +goose StatementEnd
