-- name: GetProjectsForSeed :many
SELECT id, name, start_date, end_date
FROM projects
ORDER BY created_at DESC
LIMIT 50;

-- name: GetUsersNotInTeamForProject :many
SELECT u.id
FROM users u
WHERE NOT EXISTS (
    SELECT 1 FROM team_members tm
    JOIN teams t ON tm.team_id = t.id
    WHERE tm.user_id = u.id AND t.project_id = @project_id::text
)
ORDER BY u.id;

-- name: GetUserCountNotInTeamForProject :one
SELECT COUNT(*)::int
FROM users u
WHERE NOT EXISTS (
    SELECT 1 FROM team_members tm
    JOIN teams t ON tm.team_id = t.id
    WHERE tm.user_id = u.id AND t.project_id = @project_id::text
);

-- name: GetRandomUserIDs :many
SELECT id FROM users ORDER BY random() LIMIT @limit_count::int;

-- name: GetRandomChurchIDs :many
SELECT id FROM churches ORDER BY random() LIMIT @limit_count::int;

-- name: GetTotalUserCount :one
SELECT COUNT(*)::int FROM users;

-- name: GetTotalChurchCount :one
SELECT COUNT(*)::int FROM churches;

-- name: GetTeamsWithoutLeadForProject :many
SELECT t.id
FROM teams t
WHERE t.project_id = @project_id::text
  AND NOT EXISTS (
    SELECT 1 FROM user_roles ur
    WHERE ur.team_id = t.id AND ur.role = 'TEAM_LEAD'
  )
ORDER BY t.id;

-- name: GetRandomMemberForTeams :many
SELECT DISTINCT ON (tm.team_id) tm.team_id, tm.user_id
FROM team_members tm
WHERE tm.team_id = ANY(@team_ids::text[])
ORDER BY tm.team_id, random();

-- name: DeleteTeamLeadsForProject :exec
DELETE FROM user_roles
WHERE role = 'TEAM_LEAD'
  AND team_id IN (
    SELECT id FROM teams WHERE project_id = @project_id::text
  );

-- name: GetTeamIDsForProject :many
SELECT id FROM teams WHERE project_id = @project_id::text ORDER BY id;
