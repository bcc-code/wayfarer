-- +goose Up
-- +goose StatementBegin

-- Migration: Add image metadata columns to file_uploads table
-- Stores width, height, and blurhash for uploaded images

ALTER TABLE file_uploads
ADD COLUMN width INT,
ADD COLUMN height INT,
ADD COLUMN blurhash VARCHAR(100);

COMMENT ON COLUMN file_uploads.width IS 'Image width in pixels (null for non-images)';
COMMENT ON COLUMN file_uploads.height IS 'Image height in pixels (null for non-images)';
COMMENT ON COLUMN file_uploads.blurhash IS 'Blurhash placeholder string (null for non-images)';

-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin

ALTER TABLE file_uploads
DROP COLUMN IF EXISTS width,
DROP COLUMN IF EXISTS height,
DROP COLUMN IF EXISTS blurhash;

-- +goose StatementEnd
