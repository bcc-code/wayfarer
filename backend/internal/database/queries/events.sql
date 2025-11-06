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
