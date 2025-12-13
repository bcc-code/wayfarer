-- +goose Up
-- +goose StatementBegin

-- Migration: Add file_uploads table
-- Stores metadata for files uploaded to S3 by admin users

CREATE TABLE file_uploads (
    id CHAR(28) PRIMARY KEY CHECK (id ~ '^FL[0-9A-Z]{26}$'),
    filename VARCHAR(255) NOT NULL,
    stored_filename VARCHAR(255) NOT NULL,
    file_size INT NOT NULL,
    mime_type VARCHAR(100) NOT NULL,
    public_url VARCHAR(1000) NOT NULL,
    uploaded_by CHAR(28) NOT NULL REFERENCES users(id) ON DELETE RESTRICT,
    created_at TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE INDEX idx_file_uploads_uploaded_by ON file_uploads(uploaded_by);
CREATE INDEX idx_file_uploads_created_at ON file_uploads(created_at DESC);

COMMENT ON TABLE file_uploads IS 'Stores metadata for files uploaded to S3 via /api/upload endpoint';
COMMENT ON COLUMN file_uploads.filename IS 'Original filename as uploaded by user';
COMMENT ON COLUMN file_uploads.stored_filename IS 'ULID-based filename stored in S3';
COMMENT ON COLUMN file_uploads.file_size IS 'File size in bytes';
COMMENT ON COLUMN file_uploads.mime_type IS 'MIME type detected from file header';
COMMENT ON COLUMN file_uploads.public_url IS 'Full public URL to access the file';
COMMENT ON COLUMN file_uploads.uploaded_by IS 'User ID of admin who uploaded the file';

-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin

DROP TABLE IF EXISTS file_uploads;

-- +goose StatementEnd
