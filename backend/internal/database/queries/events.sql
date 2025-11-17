-- name: GetEventByID :one
SELECT id, project_id, name, description, start_date, end_date, created_at, updated_at
FROM events
WHERE id = @id;

-- name: GetEventsByIDs :many
SELECT id, project_id, name, description, start_date, end_date, created_at, updated_at
FROM events
WHERE id = ANY(@ids::text[]);

-- name: GetEventsByProjectID :many
SELECT id, project_id, name, description, start_date, end_date, created_at, updated_at
FROM events
WHERE project_id = @project_id
ORDER BY start_date DESC;

-- name: GetEventsByProjectIDs :many
SELECT id, project_id, name, description, start_date, end_date, created_at, updated_at
FROM events
WHERE project_id = ANY(@project_ids::text[])
ORDER BY project_id, start_date DESC;

-- name: GetEventsFilteredCursor :many
SELECT id, project_id, name, description, start_date, end_date, created_at, updated_at
FROM events
WHERE
    (@ids::text[] IS NULL OR id = ANY(@ids::text[]))
    AND (@projectid::text = '' OR project_id = @projectid::text)
    AND (@startdateafter::timestamptz IS NULL OR start_date >= @startdateafter::timestamptz)
    AND (@startdatebefore::timestamptz IS NULL OR start_date <= @startdatebefore::timestamptz)
    AND (@enddateafter::timestamptz IS NULL OR end_date >= @enddateafter::timestamptz)
    AND (@enddatebefore::timestamptz IS NULL OR end_date <= @enddatebefore::timestamptz)
    AND (@aftercursor::text = '' OR id > @aftercursor::text)
    AND (@beforecursor::text = '' OR id < @beforecursor::text)
ORDER BY
    CASE WHEN @isbackward::bool = true THEN id END DESC,
    CASE WHEN @isbackward::bool = false OR @isbackward::bool IS NULL THEN id END ASC
LIMIT CASE WHEN @querylimit::int IS NULL THEN NULL ELSE @querylimit::int END;

-- name: CountEventsFiltered :one
SELECT COUNT(id)
FROM events
WHERE
    (@ids::text[] IS NULL OR id = ANY(@ids::text[]))
    AND (@projectid::text = '' OR project_id = @projectid::text)
    AND (@startdateafter::timestamptz IS NULL OR start_date >= @startdateafter::timestamptz)
    AND (@startdatebefore::timestamptz IS NULL OR start_date <= @startdatebefore::timestamptz)
    AND (@enddateafter::timestamptz IS NULL OR end_date >= @enddateafter::timestamptz)
    AND (@enddatebefore::timestamptz IS NULL OR end_date <= @enddatebefore::timestamptz);

-- name: GetEventsByUserIDs :many
SELECT e.id, e.project_id, e.name, e.description, e.start_date, e.end_date, e.created_at, e.updated_at, ue.user_id
FROM events e
INNER JOIN user_events ue ON e.id = ue.event_id
WHERE ue.user_id = ANY(@userids::text[])
ORDER BY e.start_date DESC;

-- name: CreateEvent :one
INSERT INTO events (
    id,
    project_id,
    name,
    description,
    start_date,
    end_date
)
VALUES (
    @id::text,
    @projectid::text,
    @name::text,
    @description::text,
    @startdate::timestamptz,
    @enddate::timestamptz
)
RETURNING id, project_id, name, description, start_date, end_date, created_at, updated_at;

-- name: UpdateEvent :one
UPDATE events
SET
    name = COALESCE(sqlc.narg('name')::text, name),
    description = COALESCE(sqlc.narg('description')::text, description),
    start_date = COALESCE(sqlc.narg('startdate')::timestamptz, start_date),
    end_date = COALESCE(sqlc.narg('enddate')::timestamptz, end_date),
    updated_at = now()
WHERE id = @id::text
RETURNING id, project_id, name, description, start_date, end_date, created_at, updated_at;

-- name: DeleteEvent :exec
DELETE FROM events
WHERE id = @id::text;

-- name: MoveEvent :one
UPDATE events
SET
    project_id = @newprojectid::text,
    updated_at = now()
WHERE id = @id::text
RETURNING id, project_id, name, description, start_date, end_date, created_at, updated_at;
