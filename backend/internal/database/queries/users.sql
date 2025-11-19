-- name: GetUserByID :one
SELECT id, members_id, gender, church_id, birthdate, email, name, avatar_url
FROM users
WHERE id = @id;

-- name: GetUsersByIDs :many
SELECT id, members_id, gender, church_id, birthdate, email, name, avatar_url
FROM users
WHERE id = ANY(@ids::text[]);

-- name: GetUserByMembersID :one
SELECT id, members_id, gender, church_id, birthdate, email, name, avatar_url
FROM users
WHERE members_id = @members_id;

-- name: CreateUser :one
INSERT INTO users (id, members_id, email, name, gender, birthdate, church_id, avatar_url)
VALUES (@id, @members_id, @email, @name, @gender, @birthdate, @church_id, @avatar_url)
RETURNING id, members_id, gender, church_id, birthdate, email, name, avatar_url;

-- name: GetUsersFiltered :many
SELECT DISTINCT u.id, u.members_id, u.gender, u.church_id, u.birthdate, u.email, u.name, u.avatar_url
FROM users u
LEFT JOIN user_projects up ON u.id = up.user_id AND @projectid::text IS NOT NULL
LEFT JOIN user_events ue ON u.id = ue.user_id AND @eventid::text IS NOT NULL
LEFT JOIN team_members tm ON u.id = tm.user_id AND @teamid::text IS NOT NULL
WHERE
    (@churchid::text IS NULL OR u.church_id = @churchid::text)
    AND (@gender::text IS NULL OR u.gender = @gender::text)
    AND (@minage::int IS NULL OR (EXTRACT(YEAR FROM CURRENT_DATE) - EXTRACT(YEAR FROM u.birthdate)) >= @minage::int)
    AND (@maxage::int IS NULL OR (EXTRACT(YEAR FROM CURRENT_DATE) - EXTRACT(YEAR FROM u.birthdate)) <= @maxage::int)
    AND (@projectid::text IS NULL OR up.project_id = @projectid::text)
    AND (@eventid::text IS NULL OR ue.event_id = @eventid::text)
    AND (@teamid::text IS NULL OR tm.team_id = @teamid::text)
    AND (@ids::text[] IS NULL OR u.id = ANY(@ids::text[]))
ORDER BY u.id
LIMIT CASE WHEN @querylimit::int IS NULL THEN NULL ELSE @querylimit::int END
OFFSET CASE WHEN @queryoffset::int IS NULL THEN 0 ELSE @queryoffset::int END;

-- name: GetUsersFilteredCursor :many
SELECT u.id, u.members_id, u.gender, u.church_id, u.birthdate, u.email, u.name, u.avatar_url
FROM users u
WHERE
    (@churchid::text = '' OR u.church_id = @churchid::text)
    AND (@gender::text = '' OR u.gender = @gender::text)
    AND (@minage::int IS NULL OR (EXTRACT(YEAR FROM CURRENT_DATE) - EXTRACT(YEAR FROM u.birthdate)) >= @minage::int)
    AND (@maxage::int IS NULL OR (EXTRACT(YEAR FROM CURRENT_DATE) - EXTRACT(YEAR FROM u.birthdate)) <= @maxage::int)
    AND (@projectid::text = '' OR EXISTS (
        SELECT 1 FROM user_projects up WHERE up.user_id = u.id AND up.project_id = @projectid::text
    ))
    AND (@eventid::text = '' OR EXISTS (
        SELECT 1 FROM user_events ue WHERE ue.user_id = u.id AND ue.event_id = @eventid::text
    ))
    AND (@teamid::text = '' OR EXISTS (
        SELECT 1 FROM team_members tm WHERE tm.user_id = u.id AND tm.team_id = @teamid::text
    ))
    AND (@ids::text[] IS NULL OR u.id = ANY(@ids::text[]))
    AND (@aftercursor::text = '' OR u.id > @aftercursor::text)
    AND (@beforecursor::text = '' OR u.id < @beforecursor::text)
ORDER BY
    CASE WHEN @isbackward::bool = true THEN u.id END DESC,
    CASE WHEN @isbackward::bool = false OR @isbackward::bool IS NULL THEN u.id END ASC
LIMIT CASE WHEN @querylimit::int IS NULL THEN NULL ELSE @querylimit::int END;

-- name: CountUsersFiltered :one
SELECT COUNT(u.id)
FROM users u
WHERE
    (@churchid::text = '' OR u.church_id = @churchid::text)
    AND (@gender::text = '' OR u.gender = @gender::text)
    AND (@minage::int IS NULL OR (EXTRACT(YEAR FROM CURRENT_DATE) - EXTRACT(YEAR FROM u.birthdate)) >= @minage::int)
    AND (@maxage::int IS NULL OR (EXTRACT(YEAR FROM CURRENT_DATE) - EXTRACT(YEAR FROM u.birthdate)) <= @maxage::int)
    AND (@projectid::text = '' OR EXISTS (
        SELECT 1 FROM user_projects up WHERE up.user_id = u.id AND up.project_id = @projectid::text
    ))
    AND (@eventid::text = '' OR EXISTS (
        SELECT 1 FROM user_events ue WHERE ue.user_id = u.id AND ue.event_id = @eventid::text
    ))
    AND (@teamid::text = '' OR EXISTS (
        SELECT 1 FROM team_members tm WHERE tm.user_id = u.id AND tm.team_id = @teamid::text
    ))
    AND (@ids::text[] IS NULL OR u.id = ANY(@ids::text[]));

-- name: GetUsersByTeamIDs :many
SELECT u.id, u.members_id, u.gender, u.church_id, u.birthdate, u.email, u.name, u.avatar_url, tm.team_id, tm.joined_at
FROM users u
INNER JOIN team_members tm ON u.id = tm.user_id
WHERE tm.team_id = ANY(@teamids::text[])
ORDER BY tm.team_id, tm.joined_at;

-- name: GetUsersBySuperTeamIDs :many
SELECT DISTINCT u.id, u.members_id, u.gender, u.church_id, u.birthdate, u.email, u.name, u.avatar_url
FROM users u
INNER JOIN team_members tm ON u.id = tm.user_id
INNER JOIN teams t ON tm.team_id = t.id
WHERE t.super_team_id = ANY(@superteamids::text[])
ORDER BY u.id;

-- name: GetUsersBySuperTeamIDCursor :many
WITH distinct_user_ids AS (
    SELECT DISTINCT u.id
    FROM users u
    INNER JOIN team_members tm ON u.id = tm.user_id
    INNER JOIN teams t ON tm.team_id = t.id
    WHERE t.super_team_id = @superteamid::text
)
SELECT u.id, u.members_id, u.gender, u.church_id, u.birthdate, u.email, u.name, u.avatar_url
FROM distinct_user_ids du
INNER JOIN users u ON du.id = u.id
WHERE (@aftercursor::text = '' OR u.id > @aftercursor::text)
    AND (@beforecursor::text = '' OR u.id < @beforecursor::text)
ORDER BY
    CASE WHEN @isbackward::bool = true THEN u.id END DESC,
    CASE WHEN @isbackward::bool = false OR @isbackward::bool IS NULL THEN u.id END ASC
LIMIT CASE WHEN @querylimit::int IS NULL THEN NULL ELSE @querylimit::int END;

-- name: CountUsersBySuperTeamID :one
SELECT COUNT(DISTINCT u.id)
FROM users u
INNER JOIN team_members tm ON u.id = tm.user_id
INNER JOIN teams t ON tm.team_id = t.id
WHERE t.super_team_id = @superteamid::text;
