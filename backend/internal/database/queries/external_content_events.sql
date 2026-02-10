-- name: CreateExternalContentEvent :one
INSERT INTO external_content_events (id, person_id, task_id, plan_id, source, received_at, content_progress, consumed_at)
VALUES (@id, @personid::uuid, @taskid::text, @planid::text, @source::text, @receivedat::timestamptz, @contentprogress, @consumedat::timestamptz)
RETURNING id, person_id, task_id, plan_id, source, received_at, content_progress, consumed_at;

-- name: GetExternalContentEventByID :one
SELECT id, person_id, task_id, plan_id, source, received_at, content_progress, consumed_at
FROM external_content_events
WHERE id = @id;

-- name: GetExternalContentEventsByPersonID :many
SELECT id, person_id, task_id, plan_id, source, received_at, content_progress, consumed_at
FROM external_content_events
WHERE person_id = @personid::uuid
ORDER BY received_at DESC
LIMIT CASE WHEN @querylimit::int IS NULL THEN NULL ELSE @querylimit::int END
OFFSET CASE WHEN @queryoffset::int IS NULL THEN 0 ELSE @queryoffset::int END;

-- name: GetExternalContentEventsBySource :many
SELECT id, person_id, task_id, plan_id, source, received_at, content_progress, consumed_at
FROM external_content_events
WHERE source = @source::text
ORDER BY received_at DESC
LIMIT CASE WHEN @querylimit::int IS NULL THEN NULL ELSE @querylimit::int END
OFFSET CASE WHEN @queryoffset::int IS NULL THEN 0 ELSE @queryoffset::int END;

-- name: ExternalContentEventExists :one
-- Single-event check used by the webhook handler.
SELECT EXISTS(
    SELECT 1 FROM external_content_events
    WHERE person_id = @personid::uuid AND task_id = @taskid::text
) AS exists;

-- name: GetExistingExternalContentEventTaskIDs :many
-- Batch check: given a person and a list of task_ids, return those that already have events.
-- Used by syncContentEvents to filter a whole page in one round-trip.
SELECT DISTINCT task_id
FROM external_content_events
WHERE person_id = @personid::uuid AND task_id = ANY(@taskids::text[]);

-- name: GetExternalContentEventsForProcessing :many
-- Returns one event per task_id (most recent) for a person_id, used when processing
-- pending events for newly registered users.
SELECT DISTINCT ON (task_id) id, person_id, task_id, plan_id, content_progress, consumed_at
FROM external_content_events
WHERE person_id = @personid::uuid
ORDER BY task_id, consumed_at DESC;
