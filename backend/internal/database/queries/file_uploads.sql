-- name: CreateFileUpload :one
INSERT INTO file_uploads (
    id,
    filename,
    stored_filename,
    file_size,
    mime_type,
    public_url,
    uploaded_by,
    width,
    height,
    blurhash
) VALUES (
    @id::text,
    @filename::text,
    @stored_filename::text,
    @file_size::int,
    @mime_type::text,
    @public_url::text,
    @uploaded_by::text,
    sqlc.narg(width)::int,
    sqlc.narg(height)::int,
    sqlc.narg(blurhash)::text
) RETURNING *;

-- name: GetFileUpload :one
SELECT * FROM file_uploads
WHERE id = @id::text;

-- name: ListFileUploads :many
SELECT * FROM file_uploads
ORDER BY created_at DESC
LIMIT @limit_count::int
OFFSET @offset_count::int;

-- name: GetFileUploadsByUser :many
SELECT * FROM file_uploads
WHERE uploaded_by = @uploaded_by::text
ORDER BY created_at DESC
LIMIT @limit_count::int
OFFSET @offset_count::int;

-- name: GetFileUploadByURL :one
SELECT id, public_url, width, height, blurhash FROM file_uploads
WHERE public_url = @url::text;

-- name: GetFileUploadsByURLs :many
SELECT id, public_url, width, height, blurhash FROM file_uploads
WHERE public_url = ANY(@urls::text[]);
