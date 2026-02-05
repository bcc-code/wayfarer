-- +goose Up
-- +goose StatementBegin

-- Index for file_uploads URL lookups
-- Covers: WHERE public_url = @url and WHERE public_url = ANY(@urls)
CREATE INDEX IF NOT EXISTS idx_file_uploads_public_url ON file_uploads(public_url);

-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin

DROP INDEX IF EXISTS idx_file_uploads_public_url;

-- +goose StatementEnd
