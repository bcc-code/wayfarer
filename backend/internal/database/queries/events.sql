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
