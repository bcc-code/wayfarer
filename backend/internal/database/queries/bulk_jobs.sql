-- Bulk Jobs

-- name: CreateBulkJob :one
INSERT INTO bulk_jobs (
    id,
    operation_type,
    status,
    created_by,
    project_id,
    input_params,
    total_count,
    message_id
) VALUES (
    @id::text,
    @operationtype::text,
    @status::text,
    sqlc.narg('createdby')::text,
    sqlc.narg('projectid')::text,
    @inputparams::jsonb,
    @totalcount::int,
    sqlc.narg('messageid')::text
) RETURNING *;

-- name: GetBulkJobByID :one
SELECT * FROM bulk_jobs
WHERE id = @id::text;

-- name: GetBulkJobsByCreator :many
SELECT * FROM bulk_jobs
WHERE created_by = @createdby::text
ORDER BY created_at DESC
LIMIT @limitcount::int;

-- name: GetBulkJobsByProject :many
SELECT * FROM bulk_jobs
WHERE project_id = @projectid::text
ORDER BY created_at DESC
LIMIT @limitcount::int;

-- name: GetPendingBulkJobs :many
SELECT * FROM bulk_jobs
WHERE status = 'PENDING'
ORDER BY created_at ASC
LIMIT @limitcount::int;

-- name: UpdateBulkJobStatus :one
UPDATE bulk_jobs
SET
    status = @status::text,
    started_at = CASE WHEN @status::text = 'PROCESSING' AND started_at IS NULL THEN now() ELSE started_at END,
    completed_at = CASE WHEN @status::text IN ('COMPLETED', 'FAILED') THEN now() ELSE completed_at END,
    error_message = sqlc.narg('errormessage')::text,
    error_details = sqlc.narg('errordetails')::jsonb
WHERE id = @id::text
RETURNING *;

-- name: UpdateBulkJobProgress :one
UPDATE bulk_jobs
SET
    processed_count = @processedcount::int,
    success_count = @successcount::int,
    failure_count = @failurecount::int
WHERE id = @id::text
RETURNING *;

-- name: UpdateBulkJobMessageID :exec
UPDATE bulk_jobs
SET message_id = @messageid::text
WHERE id = @id::text;

-- name: MarkBulkJobProcessing :one
UPDATE bulk_jobs
SET
    status = 'PROCESSING',
    started_at = now()
WHERE id = @id::text
RETURNING *;

-- name: MarkBulkJobCompleted :one
UPDATE bulk_jobs
SET
    status = 'COMPLETED',
    completed_at = now(),
    processed_count = @processedcount::int,
    success_count = @successcount::int,
    failure_count = @failurecount::int
WHERE id = @id::text
RETURNING *;

-- name: MarkBulkJobFailed :one
UPDATE bulk_jobs
SET
    status = 'FAILED',
    completed_at = now(),
    error_message = @errormessage::text,
    error_details = sqlc.narg('errordetails')::jsonb
WHERE id = @id::text
RETURNING *;

-- name: GetRecentBulkJobsByStatus :many
SELECT * FROM bulk_jobs
WHERE status = @status::text
  AND created_at > @since::timestamptz
ORDER BY created_at DESC
LIMIT @limitcount::int;

-- name: CleanupOldBulkJobs :exec
DELETE FROM bulk_jobs
WHERE status IN ('COMPLETED', 'FAILED')
  AND completed_at < @before::timestamptz;

-- name: ListBulkJobsForward :many
-- Paginated list of bulk jobs with optional filtering, forward pagination (first/after)
SELECT * FROM bulk_jobs
WHERE (sqlc.narg('status')::text IS NULL OR status = sqlc.narg('status')::text)
  AND (sqlc.narg('operationtype')::text IS NULL OR operation_type = sqlc.narg('operationtype')::text)
  AND (sqlc.narg('projectid')::text IS NULL OR project_id = sqlc.narg('projectid')::text)
  AND (sqlc.narg('createdby')::text IS NULL OR created_by = sqlc.narg('createdby')::text)
  AND (sqlc.narg('afterid')::text IS NULL OR id < sqlc.narg('afterid')::text)
ORDER BY created_at DESC, id DESC
LIMIT @limitcount::int;

-- name: ListBulkJobsBackward :many
-- Paginated list of bulk jobs with optional filtering, backward pagination (last/before)
SELECT * FROM (
    SELECT * FROM bulk_jobs
    WHERE (sqlc.narg('status')::text IS NULL OR status = sqlc.narg('status')::text)
      AND (sqlc.narg('operationtype')::text IS NULL OR operation_type = sqlc.narg('operationtype')::text)
      AND (sqlc.narg('projectid')::text IS NULL OR project_id = sqlc.narg('projectid')::text)
      AND (sqlc.narg('createdby')::text IS NULL OR created_by = sqlc.narg('createdby')::text)
      AND (sqlc.narg('beforeid')::text IS NULL OR id > sqlc.narg('beforeid')::text)
    ORDER BY created_at ASC, id ASC
    LIMIT @limitcount::int
) AS subquery
ORDER BY created_at DESC, id DESC;

-- name: CountBulkJobs :one
-- Count total bulk jobs matching the filter criteria
SELECT COUNT(*) FROM bulk_jobs
WHERE (sqlc.narg('status')::text IS NULL OR status = sqlc.narg('status')::text)
  AND (sqlc.narg('operationtype')::text IS NULL OR operation_type = sqlc.narg('operationtype')::text)
  AND (sqlc.narg('projectid')::text IS NULL OR project_id = sqlc.narg('projectid')::text)
  AND (sqlc.narg('createdby')::text IS NULL OR created_by = sqlc.narg('createdby')::text);

-- name: AppendBulkJobLogs :exec
-- Append new log entries to existing logs (or initialize if null)
UPDATE bulk_jobs
SET logs = COALESCE(logs, '[]'::jsonb) || @newlogs::jsonb
WHERE id = @id::text;
