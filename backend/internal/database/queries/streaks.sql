-- name: GetStreaksByIDs :many
SELECT id, project_id, name, description, created_at, updated_at
FROM streaks
WHERE id = ANY(@ids::text[]);
