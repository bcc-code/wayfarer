-- Phrase async job tracking queries (replaces Redis)

-- name: CreatePhraseAsyncJob :exec
INSERT INTO phrase_async_jobs (async_request_id, job_uid, created_at)
VALUES (@async_request_id::text, @job_uid::text, now());

-- name: GetPhraseAsyncJob :one
SELECT job_uid
FROM phrase_async_jobs
WHERE async_request_id = @async_request_id::text;

-- name: DeletePhraseAsyncJob :exec
DELETE FROM phrase_async_jobs
WHERE async_request_id = @async_request_id::text;

-- name: CleanupOldPhraseAsyncJobs :exec
DELETE FROM phrase_async_jobs
WHERE created_at < now() - interval '1 hour';

-- name: GetTranslationHash :one
SELECT hash
FROM translation_hashes
WHERE collection = @collection::text;

-- name: UpsertTranslationHash :exec
INSERT INTO translation_hashes (collection, hash, updated_at)
VALUES (@collection::text, @hash::bytea, now())
ON CONFLICT (collection)
DO UPDATE SET
    hash = EXCLUDED.hash,
    updated_at = now();
