-- name: GetTeamsByIDs :many
SELECT id, project_id, name, description, join_code, super_team_id, leaderboard_excluded, created_at, updated_at
FROM teams
WHERE id = ANY(@ids::char(28)[]);

-- name: GetTeamsFilteredCursor :many
SELECT t.id, t.project_id, t.name, t.description, t.join_code, t.super_team_id, t.leaderboard_excluded, t.created_at, t.updated_at
FROM teams t
LEFT JOIN (
    SELECT team_id, COUNT(*) as member_count
    FROM team_members
    GROUP BY team_id
) tm ON t.id = tm.team_id
LEFT JOIN user_roles ur ON ur.team_id = t.id AND ur.role = 'TEAM_LEAD'
LEFT JOIN users lead_user ON ur.user_id = lead_user.id
WHERE
    (@ids::char(28)[] IS NULL OR t.id = ANY(@ids::char(28)[]))
    AND (@projectid::char(28) = '' OR t.project_id = @projectid::char(28))
    AND (@superteamid::char(28) = '' OR t.super_team_id = @superteamid::char(28))
    AND (@nosuperteam::bool = false OR (@nosuperteam::bool = true AND t.super_team_id IS NULL))
    AND (@minmembers::int <= 0 OR COALESCE(tm.member_count, 0) >= @minmembers::int)
    AND (@maxmembers::int <= 0 OR COALESCE(tm.member_count, 0) <= @maxmembers::int)
    AND (@churchid::char(28) = '' OR lead_user.church_id = @churchid::char(28))
    AND (@aftercursor::char(28) = '' OR t.id > @aftercursor::char(28))
    AND (@beforecursor::char(28) = '' OR t.id < @beforecursor::char(28))
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
LEFT JOIN user_roles ur ON ur.team_id = t.id AND ur.role = 'TEAM_LEAD'
LEFT JOIN users lead_user ON ur.user_id = lead_user.id
WHERE
    (@ids::char(28)[] IS NULL OR t.id = ANY(@ids::char(28)[]))
    AND (@projectid::char(28) = '' OR t.project_id = @projectid::char(28))
    AND (@superteamid::char(28) = '' OR t.super_team_id = @superteamid::char(28))
    AND (@nosuperteam::bool = false OR (@nosuperteam::bool = true AND t.super_team_id IS NULL))
    AND (@minmembers::int <= 0 OR COALESCE(tm.member_count, 0) >= @minmembers::int)
    AND (@maxmembers::int <= 0 OR COALESCE(tm.member_count, 0) <= @maxmembers::int)
    AND (@churchid::char(28) = '' OR lead_user.church_id = @churchid::char(28));

-- name: GetTeamsByUserIDs :many
SELECT t.id, t.project_id, t.name, t.description, t.join_code, t.super_team_id, t.leaderboard_excluded, t.created_at, t.updated_at, tm.user_id
FROM teams t
INNER JOIN team_members tm ON t.id = tm.team_id
WHERE tm.user_id = ANY(@userids::char(28)[])
ORDER BY t.name ASC;

-- name: GetTeamsBySuperTeamIDs :many
SELECT id, project_id, name, description, join_code, super_team_id, leaderboard_excluded, created_at, updated_at
FROM teams
WHERE super_team_id = ANY(@superteamids::char(28)[])
ORDER BY name ASC;

-- name: GetTeamsByProjectIDs :many
SELECT id, project_id, name, description, join_code, super_team_id, leaderboard_excluded, created_at, updated_at
FROM teams
WHERE project_id = ANY(@project_ids::char(28)[])
ORDER BY project_id, created_at DESC;

-- name: GetTeamsByProjectIDAndChurchID :many
SELECT DISTINCT t.id, t.project_id, t.name, t.description, t.join_code, t.super_team_id, t.leaderboard_excluded, t.created_at, t.updated_at
FROM teams t
LEFT JOIN team_members tm ON t.id = tm.team_id
LEFT JOIN users member_user ON tm.user_id = member_user.id
LEFT JOIN users creator_user ON t.created_by_user_id = creator_user.id
WHERE t.project_id = @projectid::char(28)
  AND (
    -- If team has members: match by member's church
    member_user.church_id = @churchid::char(28)
    OR
    -- If team is empty: fall back to creator's church
    (NOT EXISTS (SELECT 1 FROM team_members WHERE team_id = t.id) AND creator_user.church_id = @churchid::char(28))
  )
ORDER BY t.created_at DESC;

-- name: GetUserTeamByProjectID :one
SELECT t.id, t.project_id, t.name, t.description, t.join_code, t.super_team_id, t.leaderboard_excluded, t.created_at, t.updated_at
FROM teams t
INNER JOIN team_members tm ON t.id = tm.team_id
WHERE tm.user_id = @userid::char(28)
  AND t.project_id = @projectid::char(28)
LIMIT 1;

-- name: GetUserTeamsByProjectIDBulk :many
-- One (arbitrary) team per user in the project, bulk variant of
-- GetUserTeamByProjectID
SELECT DISTINCT ON (tm.user_id) tm.user_id, t.id AS team_id
FROM teams t
INNER JOIN team_members tm ON t.id = tm.team_id
WHERE tm.user_id = ANY(@userids::char(28)[])
  AND t.project_id = @projectid::char(28)
ORDER BY tm.user_id;

-- name: CreateTeam :one
INSERT INTO teams (id, project_id, name, description, join_code, created_by_user_id)
VALUES (@id::text, @projectid::text, @name::text, @description::text, @joincode::text, @createdbyuserid::text)
RETURNING id, project_id, name, description, join_code, super_team_id, leaderboard_excluded, created_at, updated_at;

-- name: UpdateTeam :one
UPDATE teams
SET
    name = COALESCE(@name::text, name),
    description = COALESCE(@description::text, description),
    updated_at = now()
WHERE id = @id::char(28)
RETURNING id, project_id, name, description, join_code, super_team_id, leaderboard_excluded, created_at, updated_at;

-- name: DeleteTeam :exec
DELETE FROM teams
WHERE id = @id::char(28);

-- name: GetTeamByJoinCode :one
SELECT id, project_id, name, description, join_code, super_team_id, leaderboard_excluded, created_at, updated_at
FROM teams
WHERE join_code = @joincode::text;

-- name: GetTeamByJoinCodeAndProject :one
SELECT id, project_id, name, description, join_code, super_team_id, leaderboard_excluded, created_at, updated_at
FROM teams
WHERE join_code = @joincode::text
  AND project_id = @projectid::char(28);

-- name: RegenerateJoinCode :one
UPDATE teams
SET join_code = @joincode::text, updated_at = now()
WHERE id = @id::char(28)
RETURNING id, project_id, name, description, join_code, super_team_id, leaderboard_excluded, created_at, updated_at;

-- name: AddTeamMember :exec
INSERT INTO team_members (team_id, user_id)
VALUES (@teamid::text, @userid::text)
ON CONFLICT (team_id, user_id) DO NOTHING;

-- name: RemoveTeamMember :exec
DELETE FROM team_members
WHERE team_id = @teamid::char(28) AND user_id = @userid::char(28);

-- name: HasTeamMemberFromChurch :one
SELECT EXISTS(
    SELECT 1
    FROM team_members tm
    INNER JOIN users u ON tm.user_id = u.id
    WHERE tm.team_id = @teamid::char(28)
      AND u.church_id = @churchid::char(28)
);

-- name: GetTeamProjectID :one
SELECT project_id
FROM teams
WHERE id = @teamid::char(28);

-- name: RemoveUserFromTeamsInProject :exec
DELETE FROM team_members
WHERE user_id = @userid::char(28)
  AND team_id IN (
    SELECT id FROM teams WHERE project_id = @projectid::char(28)
  );

-- name: RemoveTeamLeadRole :exec
DELETE FROM user_roles
WHERE user_id = @userid::char(28)
  AND team_id = @teamid::char(28)
  AND role = 'TEAM_LEAD';

-- name: AssignTeamLeadRole :one
INSERT INTO user_roles (id, user_id, role, team_id, assigned_by, assigned_at)
VALUES (@id::text, @userid::text, 'TEAM_LEAD', @teamid::text, @assignedby::text, @assignedat::timestamptz)
RETURNING id, user_id, role, church_id, project_id, team_id, assigned_by, assigned_at;

-- name: GetTeamLeadUserID :one
SELECT user_id
FROM user_roles
WHERE team_id = @teamid::char(28)
  AND role = 'TEAM_LEAD'
LIMIT 1;

-- name: IsUserTeamMember :one
SELECT EXISTS(
    SELECT 1
    FROM team_members
    WHERE team_id = @teamid::char(28)
      AND user_id = @userid::char(28)
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
WHERE tm.team_id = @team_id::char(28)
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
    WHERE tm.user_id = @userid::char(28)
      AND t.project_id = @projectid::char(28)
) AS is_member;

-- name: IsUserInAnySuperTeamInProject :one
SELECT EXISTS(
    SELECT 1
    FROM team_members tm
    INNER JOIN teams t ON tm.team_id = t.id
    INNER JOIN super_teams st ON t.super_team_id = st.id
    WHERE tm.user_id = @userid::char(28)
      AND st.project_id = @projectid::char(28)
) AS is_member;

-- name: GetUserIDsInTeams :many
SELECT DISTINCT user_id
FROM team_members
WHERE team_id = ANY(@teamids::char(28)[]);

-- name: GetUserIDsByTeamIDs :many
-- Returns user_id and team_id for grouping by team
SELECT team_id, user_id
FROM team_members
WHERE team_id = ANY(@teamids::char(28)[]);

-- name: GetUserIDsInSuperTeams :many
SELECT DISTINCT tm.user_id
FROM team_members tm
JOIN teams t ON t.id = tm.team_id
WHERE t.super_team_id = ANY(@superteamids::char(28)[]);

-- name: GetUserIDsBySuperTeamIDs :many
-- Returns user_id and super_team_id for grouping by super team
SELECT t.super_team_id, tm.user_id
FROM team_members tm
JOIN teams t ON t.id = tm.team_id
WHERE t.super_team_id = ANY(@superteamids::char(28)[]);

-- name: GetTeamAverageAge :one
SELECT COALESCE(
    AVG(EXTRACT(YEAR FROM CURRENT_DATE) - EXTRACT(YEAR FROM u.birthdate)),
    0
)::float AS average_age
FROM team_members tm
INNER JOIN users u ON tm.user_id = u.id
WHERE tm.team_id = @teamid::char(28)
  AND u.birthdate IS NOT NULL;

-- name: UpdateTeamLeaderboardExcluded :one
UPDATE teams
SET
    leaderboard_excluded = @leaderboardexcluded::bool,
    updated_at = now()
WHERE id = @id::char(28)
RETURNING id, project_id, name, description, join_code, super_team_id, leaderboard_excluded, created_at, updated_at;

-- name: GetTeamCreatorChurchID :one
SELECT u.church_id
FROM teams t
INNER JOIN users u ON t.created_by_user_id = u.id
WHERE t.id = @teamid::char(28);
