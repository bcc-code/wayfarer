-- name: GetTeamsByIDs :many
SELECT id, project_id, name, description, join_code, super_team_id, created_at, updated_at
FROM teams
WHERE id = ANY(@ids::text[]);

-- name: GetTeamsFilteredCursor :many
SELECT t.id, t.project_id, t.name, t.description, t.join_code, t.super_team_id, t.created_at, t.updated_at
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
SELECT t.id, t.project_id, t.name, t.description, t.join_code, t.super_team_id, t.created_at, t.updated_at, tm.user_id
FROM teams t
INNER JOIN team_members tm ON t.id = tm.team_id
WHERE tm.user_id = ANY(@userids::text[])
ORDER BY t.name ASC;

-- name: GetTeamsBySuperTeamIDs :many
SELECT id, project_id, name, description, join_code, super_team_id, created_at, updated_at
FROM teams
WHERE super_team_id = ANY(@superteamids::text[])
ORDER BY name ASC;

-- name: GetTeamsByProjectIDs :many
SELECT id, project_id, name, description, join_code, super_team_id, created_at, updated_at
FROM teams
WHERE project_id = ANY(@project_ids::text[])
ORDER BY project_id, created_at DESC;

-- name: GetUserTeamByProjectID :one
SELECT t.id, t.project_id, t.name, t.description, t.join_code, t.super_team_id, t.created_at, t.updated_at
FROM teams t
INNER JOIN team_members tm ON t.id = tm.team_id
WHERE tm.user_id = @userid::text
  AND t.project_id = @projectid::text
LIMIT 1;
