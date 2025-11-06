-- name: GetSuperTeamsByIDs :many
SELECT id, project_id, name, description, created_at, updated_at
FROM super_teams
WHERE id = ANY(@ids::text[]);
