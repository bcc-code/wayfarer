-- name: GetTeamsByIDs :many
SELECT id, project_id, name, description, join_code, super_team_id, leaderboard_excluded, created_at, updated_at
FROM teams
WHERE id = ANY(@ids::text[]);

-- name: GetTeamsFilteredCursor :many
SELECT t.id, t.project_id, t.name, t.description, t.join_code, t.super_team_id, t.leaderboard_excluded, t.created_at, t.updated_at
FROM teams t
LEFT JOIN (
    SELECT team_id, COUNT(*) as member_count
    FROM team_members
    GROUP BY team_id
) tm ON t.id = tm.team_id
WHERE
    (@ids::text[] IS NULL OR t.id = ANY(@ids::text[]))
    AND (@projectid::text = '' OR t.project_id = @projectid::text)
    AND (@superteamid::text = '' OR t.super_team_id = @superteamid::text)
    AND (@nosuperteam::bool = false OR (@nosuperteam::bool = true AND t.super_team_id IS NULL))
    AND (@minmembers::int <= 0 OR COALESCE(tm.member_count, 0) >= @minmembers::int)
    AND (@maxmembers::int <= 0 OR COALESCE(tm.member_count, 0) <= @maxmembers::int)
    AND (@aftercursor::text = '' OR t.id > @aftercursor::text)
    AND (@beforecursor::text = '' OR t.id < @beforecursor::text)
ORDER BY
    CASE WHEN @isbackward::bool = true THEN t.id END DESC,
    CASE WHEN @isbackward::bool = false OR @isbackward::bool IS NULL THEN t.id END ASC
LIMIT CASE WHEN @querylimit::int IS NULL THEN NULL ELSE @querylimit::int END;

-- name: CountTeamsFiltered :one
SELECT COUNT(DISTINCT t.id)
FROM teams t
LEFT JOIN (
    SELECT team_id, COUNT(*) as member_count
    FROM team_members
    GROUP BY team_id
) tm ON t.id = tm.team_id
WHERE
    (@ids::text[] IS NULL OR t.id = ANY(@ids::text[]))
    AND (@projectid::text = '' OR t.project_id = @projectid::text)
    AND (@superteamid::text = '' OR t.super_team_id = @superteamid::text)
    AND (@nosuperteam::bool = false OR (@nosuperteam::bool = true AND t.super_team_id IS NULL))
    AND (@minmembers::int <= 0 OR COALESCE(tm.member_count, 0) >= @minmembers::int)
    AND (@maxmembers::int <= 0 OR COALESCE(tm.member_count, 0) <= @maxmembers::int);

-- name: GetTeamsByUserIDs :many
SELECT t.id, t.project_id, t.name, t.description, t.join_code, t.super_team_id, t.leaderboard_excluded, t.created_at, t.updated_at, tm.user_id
FROM teams t
INNER JOIN team_members tm ON t.id = tm.team_id
WHERE tm.user_id = ANY(@userids::text[])
ORDER BY t.name ASC;

-- name: GetTeamsBySuperTeamIDs :many
SELECT id, project_id, name, description, join_code, super_team_id, leaderboard_excluded, created_at, updated_at
FROM teams
WHERE super_team_id = ANY(@superteamids::text[])
ORDER BY name ASC;

-- name: GetTeamsByProjectIDs :many
SELECT id, project_id, name, description, join_code, super_team_id, leaderboard_excluded, created_at, updated_at
FROM teams
WHERE project_id = ANY(@project_ids::text[])
ORDER BY project_id, created_at DESC;

-- name: GetUserTeamByProjectID :one
SELECT t.id, t.project_id, t.name, t.description, t.join_code, t.super_team_id, t.leaderboard_excluded, t.created_at, t.updated_at
FROM teams t
INNER JOIN team_members tm ON t.id = tm.team_id
WHERE tm.user_id = @userid::text
  AND t.project_id = @projectid::text
LIMIT 1;

-- name: CreateTeam :one
INSERT INTO teams (id, project_id, name, description, join_code)
VALUES (@id::text, @projectid::text, @name::text, @description::text, @joincode::text)
RETURNING id, project_id, name, description, join_code, super_team_id, leaderboard_excluded, created_at, updated_at;

-- name: UpdateTeam :one
UPDATE teams
SET
    name = COALESCE(@name::text, name),
    description = COALESCE(@description::text, description),
    updated_at = now()
WHERE id = @id::text
RETURNING id, project_id, name, description, join_code, super_team_id, leaderboard_excluded, created_at, updated_at;

-- name: DeleteTeam :exec
DELETE FROM teams
WHERE id = @id::text;

-- name: GetTeamByJoinCode :one
SELECT id, project_id, name, description, join_code, super_team_id, leaderboard_excluded, created_at, updated_at
FROM teams
WHERE join_code = @joincode::text;

-- name: RegenerateJoinCode :one
UPDATE teams
SET join_code = @joincode::text, updated_at = now()
WHERE id = @id::text
RETURNING id, project_id, name, description, join_code, super_team_id, leaderboard_excluded, created_at, updated_at;

-- name: AddTeamMember :exec
INSERT INTO team_members (team_id, user_id)
VALUES (@teamid::text, @userid::text)
ON CONFLICT (team_id, user_id) DO NOTHING;

-- name: RemoveTeamMember :exec
DELETE FROM team_members
WHERE team_id = @teamid::text AND user_id = @userid::text;

-- name: HasTeamMemberFromChurch :one
SELECT EXISTS(
    SELECT 1
    FROM team_members tm
    INNER JOIN users u ON tm.user_id = u.id
    WHERE tm.team_id = @teamid::text
      AND u.church_id = @churchid::text
);

-- name: GetTeamProjectID :one
SELECT project_id
FROM teams
WHERE id = @teamid::text;

-- name: RemoveUserFromTeamsInProject :exec
DELETE FROM team_members
WHERE user_id = @userid::text
  AND team_id IN (
    SELECT id FROM teams WHERE project_id = @projectid::text
  );

-- name: RemoveTeamLeadRole :exec
DELETE FROM user_roles
WHERE user_id = @userid::text
  AND team_id = @teamid::text
  AND role = 'TEAM_LEAD';

-- name: AssignTeamLeadRole :one
INSERT INTO user_roles (id, user_id, role, team_id, assigned_by, assigned_at)
VALUES (@id::text, @userid::text, 'TEAM_LEAD', @teamid::text, @assignedby::text, @assignedat::timestamptz)
RETURNING id, user_id, role, church_id, project_id, team_id, assigned_by, assigned_at;

-- name: GetTeamLeadUserID :one
SELECT user_id
FROM user_roles
WHERE team_id = @teamid::text
  AND role = 'TEAM_LEAD'
LIMIT 1;

-- name: IsUserTeamMember :one
SELECT EXISTS(
    SELECT 1
    FROM team_members
    WHERE team_id = @teamid::text
      AND user_id = @userid::text
);

-- name: JoinCodeExists :one
SELECT EXISTS(
    SELECT 1
    FROM teams
    WHERE join_code = @joincode::text
);

-- name: GetTeamMemberLeaderboard :many
SELECT
    u.id AS user_id,
    u.name AS user_name,
    u.members_id,
    u.gender,
    u.church_id,
    c.name AS church_name,
    u.birthdate,
    u.email,
    u.avatar_url,
    COALESCE(SUM(sj.points), 0)::bigint AS score,
    RANK() OVER (ORDER BY COALESCE(SUM(sj.points), 0)::bigint DESC) AS rank
FROM team_members tm
INNER JOIN users u ON tm.user_id = u.id
INNER JOIN teams t ON tm.team_id = t.id
INNER JOIN churches c ON u.church_id = c.id
LEFT JOIN score_journal sj ON sj.user_id = u.id
    AND sj.project_id = t.project_id
WHERE tm.team_id = @team_id::text
GROUP BY
    u.id,
    u.name,
    u.members_id,
    u.gender,
    u.church_id,
    c.name,
    u.birthdate,
    u.email,
    u.avatar_url
ORDER BY score DESC, u.name ASC;

-- name: IsUserInAnyTeamInProject :one
SELECT EXISTS(
    SELECT 1
    FROM team_members tm
    INNER JOIN teams t ON tm.team_id = t.id
    WHERE tm.user_id = @userid::text
      AND t.project_id = @projectid::text
) AS is_member;

-- name: IsUserInAnySuperTeamInProject :one
SELECT EXISTS(
    SELECT 1
    FROM team_members tm
    INNER JOIN teams t ON tm.team_id = t.id
    INNER JOIN super_teams st ON t.super_team_id = st.id
    WHERE tm.user_id = @userid::text
      AND st.project_id = @projectid::text
) AS is_member;

-- name: GetUserIDsInTeams :many
SELECT DISTINCT user_id
FROM team_members
WHERE team_id = ANY(@teamids::text[]);

-- name: GetUserIDsInSuperTeams :many
SELECT DISTINCT tm.user_id
FROM team_members tm
JOIN teams t ON t.id = tm.team_id
WHERE t.super_team_id = ANY(@superteamids::text[]);

-- name: GetTeamAverageAge :one
SELECT COALESCE(
    AVG(EXTRACT(YEAR FROM CURRENT_DATE) - EXTRACT(YEAR FROM u.birthdate)),
    0
)::float AS average_age
FROM team_members tm
INNER JOIN users u ON tm.user_id = u.id
WHERE tm.team_id = @teamid::text
  AND u.birthdate IS NOT NULL;

-- name: UpdateTeamLeaderboardExcluded :one
UPDATE teams
SET
    leaderboard_excluded = @leaderboardexcluded::bool,
    updated_at = now()
WHERE id = @id::text
RETURNING id, project_id, name, description, join_code, super_team_id, leaderboard_excluded, created_at, updated_at;
