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
