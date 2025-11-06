-- name: GetSuperTeamsByIDs :many
SELECT id, project_id, name, description, created_at, updated_at
FROM super_teams
WHERE id = ANY(@ids::text[]);

-- name: GetSuperTeamsFilteredCursor :many
SELECT st.id, st.project_id, st.name, st.description, st.created_at, st.updated_at
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
    (@ids::text[] IS NULL OR st.id = ANY(@ids::text[]))
    AND (@projectid::text = '' OR st.project_id = @projectid::text)
    AND (@minteams::int <= 0 OR COALESCE(t.team_count, 0) >= @minteams::int)
    AND (@maxteams::int <= 0 OR COALESCE(t.team_count, 0) <= @maxteams::int)
    AND (@minmembers::int <= 0 OR COALESCE(m.member_count, 0) >= @minmembers::int)
    AND (@maxmembers::int <= 0 OR COALESCE(m.member_count, 0) <= @maxmembers::int)
    AND (@aftercursor::text = '' OR st.id > @aftercursor::text)
    AND (@beforecursor::text = '' OR st.id < @beforecursor::text)
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
    (@ids::text[] IS NULL OR st.id = ANY(@ids::text[]))
    AND (@projectid::text = '' OR st.project_id = @projectid::text)
    AND (@minteams::int <= 0 OR COALESCE(t.team_count, 0) >= @minteams::int)
    AND (@maxteams::int <= 0 OR COALESCE(t.team_count, 0) <= @maxteams::int)
    AND (@minmembers::int <= 0 OR COALESCE(m.member_count, 0) >= @minmembers::int)
    AND (@maxmembers::int <= 0 OR COALESCE(m.member_count, 0) <= @maxmembers::int);
