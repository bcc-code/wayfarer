-- +goose Up
-- +goose StatementBegin

CREATE INDEX IF NOT EXISTS idx_score_journal_user_project_event
ON score_journal(user_id, project_id, event_id) INCLUDE (points);

-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin

DROP INDEX IF EXISTS idx_score_journal_user_project_event;

-- +goose StatementEnd
