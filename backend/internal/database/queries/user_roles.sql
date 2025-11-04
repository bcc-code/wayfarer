-- name: GetUserRoles :many
SELECT id, user_id, role, church_id, project_id, team_id, assigned_by, assigned_at
FROM user_roles
WHERE user_id = @user_id;

-- name: AssignRole :one
INSERT INTO user_roles (id, user_id, role, church_id, project_id, team_id, assigned_by, assigned_at)
VALUES (@id, @user_id, @role, @church_id, @project_id, @team_id, @assigned_by, @assigned_at)
RETURNING id, user_id, role, church_id, project_id, team_id, assigned_by, assigned_at;

-- name: RevokeRole :exec
DELETE FROM user_roles
WHERE user_id = @user_id
  AND role = @role
  AND (church_id = @church_id OR (church_id IS NULL AND @church_id::text IS NULL))
  AND (project_id = @project_id OR (project_id IS NULL AND @project_id::text IS NULL))
  AND (team_id = @team_id OR (team_id IS NULL AND @team_id::text IS NULL));

-- name: HasRole :one
SELECT EXISTS (
    SELECT 1
    FROM user_roles
    WHERE user_id = @user_id
      AND role = @role
) AS has_role;

-- name: HasRoleInChurch :one
SELECT EXISTS (
    SELECT 1
    FROM user_roles
    WHERE user_id = @user_id
      AND role = @role
      AND church_id = @church_id
) AS has_role;

-- name: HasRoleInProject :one
SELECT EXISTS (
    SELECT 1
    FROM user_roles
    WHERE user_id = @user_id
      AND role = @role
      AND project_id = @project_id
) AS has_role;

-- name: HasRoleInTeam :one
SELECT EXISTS (
    SELECT 1
    FROM user_roles
    WHERE user_id = @user_id
      AND role = @role
      AND team_id = @team_id
) AS has_role;

-- name: GetUsersWithRole :many
SELECT DISTINCT u.id, u.members_id, u.gender, u.church_id, u.birthdate, u.email, u.name, u.avatar_url
FROM users u
INNER JOIN user_roles ur ON u.id = ur.user_id
WHERE ur.role = @role
  AND ur.church_id IS NULL
  AND ur.project_id IS NULL
  AND ur.team_id IS NULL;

-- name: GetUsersWithRoleInChurch :many
SELECT DISTINCT u.id, u.members_id, u.gender, u.church_id, u.birthdate, u.email, u.name, u.avatar_url
FROM users u
INNER JOIN user_roles ur ON u.id = ur.user_id
WHERE ur.role = @role
  AND ur.church_id = @church_id;

-- name: GetUsersWithRoleInProject :many
SELECT DISTINCT u.id, u.members_id, u.gender, u.church_id, u.birthdate, u.email, u.name, u.avatar_url
FROM users u
INNER JOIN user_roles ur ON u.id = ur.user_id
WHERE ur.role = @role
  AND ur.project_id = @project_id;

-- name: GetUsersWithRoleInTeam :many
SELECT DISTINCT u.id, u.members_id, u.gender, u.church_id, u.birthdate, u.email, u.name, u.avatar_url
FROM users u
INNER JOIN user_roles ur ON u.id = ur.user_id
WHERE ur.role = @role
  AND ur.team_id = @team_id;

-- name: GetAllRolesForUsers :many
SELECT ur.user_id, ur.id, ur.role, ur.church_id, ur.project_id, ur.team_id, ur.assigned_by, ur.assigned_at
FROM user_roles ur
WHERE ur.user_id = ANY(@user_ids::text[]);

-- name: GetRoleByID :one
SELECT id, user_id, role, church_id, project_id, team_id, assigned_by, assigned_at
FROM user_roles
WHERE id = @id;

-- name: GetUserRolesInChurch :many
SELECT id, user_id, role, church_id, project_id, team_id, assigned_by, assigned_at
FROM user_roles
WHERE user_id = @user_id
  AND church_id = @church_id;

-- name: GetUserRolesInProject :many
SELECT id, user_id, role, church_id, project_id, team_id, assigned_by, assigned_at
FROM user_roles
WHERE user_id = @user_id
  AND project_id = @project_id;

-- name: GetUserRolesInTeam :many
SELECT id, user_id, role, church_id, project_id, team_id, assigned_by, assigned_at
FROM user_roles
WHERE user_id = @user_id
  AND team_id = @team_id;
