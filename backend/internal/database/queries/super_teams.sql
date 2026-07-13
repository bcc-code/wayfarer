-- name: GetSuperTeamsByIDs :many
SELECT id, project_id, name, description, image_url, color, created_at, updated_at
FROM super_teams
WHERE id = ANY(@ids::char(28)[]);

-- name: GetSuperTeamsFilteredCursor :many
SELECT st.id, st.project_id, st.name, st.description, st.image_url, st.color, st.created_at, st.updated_at
FROM super_teams st
LEFT JOIN (
    SELECT super_team_id, COUNT(*) as team_count
    FROM teams
    WHERE super_team_id IS NOT NULL
    GROUP BY super_team_id
) t ON st.id = t.super_team_id
LEFT JOIN (
    SELECT t.super_team_id, COUNT(DISTINCT tm.user_id) as member_count
    FROM teams t
    INNER JOIN team_members tm ON t.id = tm.team_id
    WHERE t.super_team_id IS NOT NULL
    GROUP BY t.super_team_id
) m ON st.id = m.super_team_id
WHERE
    (@ids::char(28)[] IS NULL OR st.id = ANY(@ids::char(28)[]))
    AND (@projectid::char(28) = '' OR st.project_id = @projectid::char(28))
    AND (@minteams::int <= 0 OR COALESCE(t.team_count, 0) >= @minteams::int)
    AND (@maxteams::int <= 0 OR COALESCE(t.team_count, 0) <= @maxteams::int)
    AND (@minmembers::int <= 0 OR COALESCE(m.member_count, 0) >= @minmembers::int)
    AND (@maxmembers::int <= 0 OR COALESCE(m.member_count, 0) <= @maxmembers::int)
    AND (@aftercursor::char(28) = '' OR st.id > @aftercursor::char(28))
    AND (@beforecursor::char(28) = '' OR st.id < @beforecursor::char(28))
ORDER BY
    CASE WHEN @isbackward::bool = true THEN st.id END DESC,
    CASE WHEN @isbackward::bool = false OR @isbackward::bool IS NULL THEN st.id END ASC
LIMIT CASE WHEN @querylimit::int IS NULL THEN NULL ELSE @querylimit::int END;

-- name: CountSuperTeamsFiltered :one
SELECT COUNT(DISTINCT st.id)
FROM super_teams st
LEFT JOIN (
    SELECT super_team_id, COUNT(*) as team_count
    FROM teams
    WHERE super_team_id IS NOT NULL
    GROUP BY super_team_id
) t ON st.id = t.super_team_id
LEFT JOIN (
    SELECT t.super_team_id, COUNT(DISTINCT tm.user_id) as member_count
    FROM teams t
    INNER JOIN team_members tm ON t.id = tm.team_id
    WHERE t.super_team_id IS NOT NULL
    GROUP BY t.super_team_id
) m ON st.id = m.super_team_id
WHERE
    (@ids::char(28)[] IS NULL OR st.id = ANY(@ids::char(28)[]))
    AND (@projectid::char(28) = '' OR st.project_id = @projectid::char(28))
    AND (@minteams::int <= 0 OR COALESCE(t.team_count, 0) >= @minteams::int)
    AND (@maxteams::int <= 0 OR COALESCE(t.team_count, 0) <= @maxteams::int)
    AND (@minmembers::int <= 0 OR COALESCE(m.member_count, 0) >= @minmembers::int)
    AND (@maxmembers::int <= 0 OR COALESCE(m.member_count, 0) <= @maxmembers::int);

-- name: GetSuperTeamsByUserIDs :many
SELECT DISTINCT st.id, st.project_id, st.name, st.description, st.image_url, st.color, st.created_at, st.updated_at, tm.user_id
FROM super_teams st
INNER JOIN teams t ON st.id = t.super_team_id
INNER JOIN team_members tm ON t.id = tm.team_id
WHERE tm.user_id = ANY(@userids::char(28)[])
ORDER BY st.name ASC;

-- name: CreateSuperTeam :one
INSERT INTO super_teams (id, project_id, name, description, image_url, color)
VALUES (@id, @project_id, @name, @description, @image_url, @color)
RETURNING *;

-- name: DeleteSuperTeamsByProjectID :exec
DELETE FROM super_teams WHERE project_id = @project_id;

-- name: ClearSuperTeamAssignmentsForProject :exec
UPDATE teams SET super_team_id = NULL WHERE project_id = @project_id;

-- name: AssignTeamToSuperTeam :exec
UPDATE teams SET super_team_id = @super_team_id WHERE id = @team_id;

-- name: UpdateSuperTeam :one
UPDATE super_teams
SET
    name = COALESCE(sqlc.narg('name')::text, name),
    description = COALESCE(sqlc.narg('description')::text, description),
    image_url = COALESCE(sqlc.narg('image_url')::text, image_url),
    color = COALESCE(sqlc.narg('color')::text, color),
    updated_at = now()
WHERE id = @id::char(28)
RETURNING *;

-- name: DeleteSuperTeam :exec
DELETE FROM super_teams WHERE id = @id::char(28);

-- name: ClearTeamsFromSuperTeam :exec
UPDATE teams SET super_team_id = NULL WHERE super_team_id = @super_team_id::char(28);

-- name: GetTeamsWithScoresForDistribution :many
-- Returns teams with total score > 0 and team lead's church
SELECT
    t.id AS team_id,
    t.name AS team_name,
    COALESCE(lead_user.church_id, '') AS church_id,
    COALESCE(c.name, '') AS church_name,
    COALESCE(SUM(sj.points), 0)::bigint AS total_score,
    COUNT(DISTINCT tm.user_id)::int AS member_count
FROM teams t
LEFT JOIN user_roles ur ON ur.team_id = t.id AND ur.role = 'TEAM_LEAD'
LEFT JOIN users lead_user ON ur.user_id = lead_user.id
LEFT JOIN churches c ON lead_user.church_id = c.id
INNER JOIN team_members tm ON t.id = tm.team_id
LEFT JOIN score_journal sj ON sj.user_id = tm.user_id AND sj.project_id = t.project_id
WHERE t.project_id = @project_id
  AND t.leaderboard_excluded = false
GROUP BY t.id, t.name, lead_user.church_id, c.name
HAVING COALESCE(SUM(sj.points), 0) > 0;

-- name: GetTeamsWithScoresAndAttendingForDistribution :many
-- Returns teams with total score > 0, team lead's church, and attending member count
SELECT
    t.id AS team_id,
    t.name AS team_name,
    COALESCE(lead_user.church_id, '') AS church_id,
    COALESCE(c.name, '') AS church_name,
    COALESCE(SUM(sj.points), 0)::bigint AS total_score,
    COUNT(DISTINCT tm.user_id)::int AS member_count,
    COUNT(DISTINCT CASE WHEN tm.user_id = ANY(@attending_user_ids::char(28)[]) THEN tm.user_id END)::int AS attending_count
FROM teams t
LEFT JOIN user_roles ur ON ur.team_id = t.id AND ur.role = 'TEAM_LEAD'
LEFT JOIN users lead_user ON ur.user_id = lead_user.id
LEFT JOIN churches c ON lead_user.church_id = c.id
INNER JOIN team_members tm ON t.id = tm.team_id
LEFT JOIN score_journal sj ON sj.user_id = tm.user_id AND sj.project_id = t.project_id
WHERE t.project_id = @project_id
  AND t.leaderboard_excluded = false
GROUP BY t.id, t.name, lead_user.church_id, c.name
HAVING COALESCE(SUM(sj.points), 0) > 0;
