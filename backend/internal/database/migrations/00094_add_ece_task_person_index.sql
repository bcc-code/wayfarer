-- +goose Up
CREATE INDEX idx_ece_task_person ON public.external_content_events USING btree (task_id, person_id);

-- +goose Down
DROP INDEX IF EXISTS idx_ece_task_person;
