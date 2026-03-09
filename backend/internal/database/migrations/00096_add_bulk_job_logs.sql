-- +goose Up
-- +goose StatementBegin

-- Add logs column for capturing slog output during job processing
ALTER TABLE bulk_jobs ADD COLUMN logs JSONB;

-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin

ALTER TABLE bulk_jobs DROP COLUMN IF EXISTS logs;

-- +goose StatementEnd
