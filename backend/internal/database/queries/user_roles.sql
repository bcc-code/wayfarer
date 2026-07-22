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
-- HasRole checks for a GLOBAL (unscoped) role only. The scope columns must all be
-- NULL so a scoped role (e.g. a PROJECT_ADMIN of a single project) can never satisfy
-- a global-role check. Global roles (SUPERADMIN/ADMIN/USER/M2M) always have NULL
-- scope columns (enforced by the CHECK constraint in migration 00005).
SELECT EXISTS (
    SELECT 1
    FROM user_roles
    WHERE user_id = @user_id::char(28)
      AND role = @role
      AND church_id IS NULL
      AND project_id IS NULL
      AND team_id IS NULL
) AS has_role;

-- name: HasAnyProjectAdminRole :one
-- Reports whether the user is a PROJECT_ADMIN of any project (used for coarse,
-- non-target-specific gates such as file uploads). This intentionally matches a
-- scoped row, unlike HasRole.
SELECT EXISTS (
    SELECT 1
    FROM user_roles
    WHERE user_id = @user_id::char(28)
      AND role = 'PROJECT_ADMIN'
      AND project_id IS NOT NULL
) AS has_role;

-- name: CanProjectAdminAccessUser :one
-- Reports whether the actor is a PROJECT_ADMIN of at least one project the target
-- user is a member of. This scopes project-admin user access to their own projects.
SELECT EXISTS (
    SELECT 1
    FROM user_roles ur
    JOIN user_projects up ON up.project_id = ur.project_id
    WHERE ur.user_id = @actor_id::char(28)
      AND ur.role = 'PROJECT_ADMIN'
      AND ur.project_id IS NOT NULL
      AND up.user_id = @target_id::char(28)
) AS can_access;

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
WHERE ur.user_id = ANY(@user_ids::char(28)[]);

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
