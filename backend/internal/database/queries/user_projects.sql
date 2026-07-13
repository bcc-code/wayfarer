-- name: JoinProject :exec
INSERT INTO user_projects (user_id, project_id)
VALUES (@userid::text, @projectid::text)
ON CONFLICT (user_id, project_id) DO NOTHING;

-- name: LeaveProject :exec
DELETE FROM user_projects
WHERE user_id = @userid::char(28) AND project_id = @projectid::char(28);

-- name: GetUserProjects :many
SELECT project_id, joined_at
FROM user_projects
WHERE user_id = @userid::char(28)
ORDER BY joined_at DESC;

-- name: IsUserInProject :one
SELECT EXISTS(
    SELECT 1
    FROM user_projects
    WHERE user_id = @userid::char(28) AND project_id = @projectid::char(28)
) AS is_member;

-- name: GetUserIDsInChurchAndProject :many
SELECT DISTINCT up.user_id
FROM user_projects up
JOIN users u ON u.id = up.user_id
WHERE u.church_id = @churchid::char(28)
  AND up.project_id = @projectid::char(28);

-- name: GetUserIDsInProject :many
SELECT DISTINCT user_id
FROM user_projects
WHERE project_id = @projectid::char(28);
