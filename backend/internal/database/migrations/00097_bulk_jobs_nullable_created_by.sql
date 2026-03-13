-- +goose Up
-- +goose StatementBegin

-- Allow M2M users to create bulk jobs (they don't exist in users table)
ALTER TABLE bulk_jobs ALTER COLUMN created_by DROP NOT NULL;

-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin

-- Revert to NOT NULL (will fail if there are NULL values)
ALTER TABLE bulk_jobs ALTER COLUMN created_by SET NOT NULL;

-- +goose StatementEnd
