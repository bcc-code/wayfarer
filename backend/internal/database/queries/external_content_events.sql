-- name: CreateExternalContentEvent :one
INSERT INTO external_content_events (id, person_id, content_id, reading_plan_id, source, received_at, content_progress)
VALUES (@id, @personid::uuid, @contentid::text, @readingplanid::text, @source::text, @receivedat::timestamptz, @contentprogress)
RETURNING id, person_id, content_id, reading_plan_id, source, received_at, content_progress;

-- name: GetExternalContentEventByID :one
SELECT id, person_id, content_id, reading_plan_id, source, received_at, content_progress
FROM external_content_events
WHERE id = @id;

-- name: GetExternalContentEventsByPersonID :many
SELECT id, person_id, content_id, reading_plan_id, source, received_at, content_progress
FROM external_content_events
WHERE person_id = @personid::uuid
ORDER BY received_at DESC
LIMIT CASE WHEN @querylimit::int IS NULL THEN NULL ELSE @querylimit::int END
OFFSET CASE WHEN @queryoffset::int IS NULL THEN 0 ELSE @queryoffset::int END;

-- name: GetExternalContentEventsBySource :many
SELECT id, person_id, content_id, reading_plan_id, source, received_at, content_progress
FROM external_content_events
WHERE source = @source::text
ORDER BY received_at DESC
LIMIT CASE WHEN @querylimit::int IS NULL THEN NULL ELSE @querylimit::int END
OFFSET CASE WHEN @queryoffset::int IS NULL THEN 0 ELSE @queryoffset::int END;
