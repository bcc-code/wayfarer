-- +goose Up
CREATE INDEX IF NOT EXISTS idx_external_content_events_person_task
    ON external_content_events (person_id, task_id);

-- +goose Down
DROP INDEX IF EXISTS idx_external_content_events_person_task;
