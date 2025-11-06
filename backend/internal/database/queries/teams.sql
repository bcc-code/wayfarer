-- name: GetTeamsByIDs :many
SELECT id, project_id, name, description, join_code, super_team_id, created_at, updated_at
FROM teams
WHERE id = ANY(@ids::text[]);
