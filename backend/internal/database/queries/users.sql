-- name: GetUserByID :one
SELECT id, members_id, person_uuid, gender, church_id, church_locked_until, birthdate, email, name, first_name, last_name, middle_name, display_name, avatar_url, language, created_at
FROM users
WHERE id = @id;

-- name: GetUsersByIDs :many
SELECT id, members_id, person_uuid, gender, church_id, church_locked_until, birthdate, email, name, first_name, last_name, middle_name, display_name, avatar_url, language, created_at
FROM users
WHERE id = ANY(@ids::char(28)[]);

-- name: GetUserByMembersID :one
SELECT id, members_id, person_uuid, gender, church_id, church_locked_until, birthdate, email, name, first_name, last_name, middle_name, display_name, avatar_url, language, created_at
FROM users
WHERE members_id = @members_id;

-- name: CreateUser :one
INSERT INTO users (id, members_id, person_uuid, email, name, first_name, last_name, middle_name, display_name, gender, birthdate, church_id, avatar_url, language)
VALUES (@id, @members_id, @person_uuid, @email, @name, @first_name, @last_name, @middle_name, @display_name, @gender, @birthdate, @church_id, @avatar_url, @language)
RETURNING id, members_id, person_uuid, gender, church_id, church_locked_until, birthdate, email, name, first_name, last_name, middle_name, display_name, avatar_url, language, created_at;

-- name: GetUsersFiltered :many
SELECT DISTINCT u.id, u.members_id, u.person_uuid, u.gender, u.church_id, u.church_locked_until, u.birthdate, u.email, u.name, u.first_name, u.last_name, u.middle_name, u.display_name, u.avatar_url, u.language, u.created_at
FROM users u
LEFT JOIN user_projects up ON u.id = up.user_id AND @projectid::char(28) IS NOT NULL
LEFT JOIN user_events ue ON u.id = ue.user_id AND @eventid::char(28) IS NOT NULL
LEFT JOIN team_members tm ON u.id = tm.user_id AND @teamid::char(28) IS NOT NULL
WHERE
    (@churchid::char(28) IS NULL OR u.church_id = @churchid::char(28))
    AND (@gender::text IS NULL OR u.gender = @gender::text)
    AND (@minage::int IS NULL OR (EXTRACT(YEAR FROM CURRENT_DATE) - EXTRACT(YEAR FROM u.birthdate)) >= @minage::int)
    AND (@maxage::int IS NULL OR (EXTRACT(YEAR FROM CURRENT_DATE) - EXTRACT(YEAR FROM u.birthdate)) <= @maxage::int)
    AND (@projectid::char(28) IS NULL OR up.project_id = @projectid::char(28))
    AND (@eventid::char(28) IS NULL OR ue.event_id = @eventid::char(28))
    AND (@teamid::char(28) IS NULL OR tm.team_id = @teamid::char(28))
    AND (@ids::char(28)[] IS NULL OR u.id = ANY(@ids::char(28)[]))
ORDER BY u.id
LIMIT CASE WHEN @querylimit::int IS NULL THEN NULL ELSE @querylimit::int END
OFFSET CASE WHEN @queryoffset::int IS NULL THEN 0 ELSE @queryoffset::int END;

-- name: GetUsersFilteredCursor :many
SELECT u.id, u.members_id, u.person_uuid, u.gender, u.church_id, u.church_locked_until, u.birthdate, u.email, u.name, u.first_name, u.last_name, u.middle_name, u.display_name, u.avatar_url, u.language, u.created_at
FROM users u
WHERE
    (@query::text = '' OR (u.name ILIKE '%' || @query::text || '%' OR u.email ILIKE '%' || @query::text || '%'))
    AND (@churchid::char(28) = '' OR u.church_id = @churchid::char(28))
    AND (@gender::text = '' OR u.gender = @gender::text)
    AND (@minage::int IS NULL OR (EXTRACT(YEAR FROM CURRENT_DATE) - EXTRACT(YEAR FROM u.birthdate)) >= @minage::int)
    AND (@maxage::int IS NULL OR (EXTRACT(YEAR FROM CURRENT_DATE) - EXTRACT(YEAR FROM u.birthdate)) <= @maxage::int)
    AND (@projectid::char(28) = '' OR EXISTS (
        SELECT 1 FROM user_projects up WHERE up.user_id = u.id AND up.project_id = @projectid::char(28)
    ))
    AND (@eventid::char(28) = '' OR EXISTS (
        SELECT 1 FROM user_events ue WHERE ue.user_id = u.id AND ue.event_id = @eventid::char(28)
    ))
    AND (@teamid::char(28) = '' OR EXISTS (
        SELECT 1 FROM team_members tm WHERE tm.user_id = u.id AND tm.team_id = @teamid::char(28)
    ))
    AND (@ids::char(28)[] IS NULL OR u.id = ANY(@ids::char(28)[]))
    AND (@aftercursor::char(28) = '' OR u.id > @aftercursor::char(28))
    AND (@beforecursor::char(28) = '' OR u.id < @beforecursor::char(28))
ORDER BY
    CASE WHEN @isbackward::bool = true THEN u.id END DESC,
    CASE WHEN @isbackward::bool = false OR @isbackward::bool IS NULL THEN u.id END ASC
LIMIT CASE WHEN @querylimit::int IS NULL THEN NULL ELSE @querylimit::int END;

-- name: CountUsersFiltered :one
SELECT COUNT(u.id)
FROM users u
WHERE
    (@query::text = '' OR (u.name ILIKE '%' || @query::text || '%' OR u.email ILIKE '%' || @query::text || '%'))
    AND (@churchid::char(28) = '' OR u.church_id = @churchid::char(28))
    AND (@gender::text = '' OR u.gender = @gender::text)
    AND (@minage::int IS NULL OR (EXTRACT(YEAR FROM CURRENT_DATE) - EXTRACT(YEAR FROM u.birthdate)) >= @minage::int)
    AND (@maxage::int IS NULL OR (EXTRACT(YEAR FROM CURRENT_DATE) - EXTRACT(YEAR FROM u.birthdate)) <= @maxage::int)
    AND (@projectid::char(28) = '' OR EXISTS (
        SELECT 1 FROM user_projects up WHERE up.user_id = u.id AND up.project_id = @projectid::char(28)
    ))
    AND (@eventid::char(28) = '' OR EXISTS (
        SELECT 1 FROM user_events ue WHERE ue.user_id = u.id AND ue.event_id = @eventid::char(28)
    ))
    AND (@teamid::char(28) = '' OR EXISTS (
        SELECT 1 FROM team_members tm WHERE tm.user_id = u.id AND tm.team_id = @teamid::char(28)
    ))
    AND (@ids::char(28)[] IS NULL OR u.id = ANY(@ids::char(28)[]));

-- name: GetUsersByTeamIDs :many
SELECT
    u.id,
    u.members_id,
    u.person_uuid,
    u.gender,
    u.church_id,
    u.church_locked_until,
    u.birthdate,
    u.email,
    u.name,
    u.first_name,
    u.last_name,
    u.middle_name,
    u.display_name,
    u.avatar_url,
    u.language,
    u.created_at,
    tm.team_id,
    tm.joined_at,
    EXISTS(
        SELECT 1
        FROM user_roles ur
        WHERE ur.user_id = u.id
        AND ur.team_id = tm.team_id
        AND ur.role = 'TEAM_LEAD'
    ) as is_team_lead
FROM users u
INNER JOIN team_members tm ON u.id = tm.user_id
WHERE tm.team_id = ANY(@teamids::char(28)[])
ORDER BY tm.team_id, tm.joined_at;

-- name: GetUsersBySuperTeamIDs :many
SELECT DISTINCT u.id, u.members_id, u.person_uuid, u.gender, u.church_id, u.church_locked_until, u.birthdate, u.email, u.name, u.first_name, u.last_name, u.middle_name, u.display_name, u.avatar_url, u.language, u.created_at
FROM users u
INNER JOIN team_members tm ON u.id = tm.user_id
INNER JOIN teams t ON tm.team_id = t.id
WHERE t.super_team_id = ANY(@superteamids::char(28)[])
ORDER BY u.id;

-- name: GetUsersBySuperTeamIDCursor :many
WITH distinct_user_ids AS (
    SELECT DISTINCT u.id
    FROM users u
    INNER JOIN team_members tm ON u.id = tm.user_id
    INNER JOIN teams t ON tm.team_id = t.id
    WHERE t.super_team_id = @superteamid::char(28)
)
SELECT u.id, u.members_id, u.person_uuid, u.gender, u.church_id, u.church_locked_until, u.birthdate, u.email, u.name, u.first_name, u.last_name, u.middle_name, u.display_name, u.avatar_url, u.language, u.created_at
FROM distinct_user_ids du
INNER JOIN users u ON du.id = u.id
WHERE (@aftercursor::char(28) = '' OR u.id > @aftercursor::char(28))
    AND (@beforecursor::char(28) = '' OR u.id < @beforecursor::char(28))
ORDER BY
    CASE WHEN @isbackward::bool = true THEN u.id END DESC,
    CASE WHEN @isbackward::bool = false OR @isbackward::bool IS NULL THEN u.id END ASC
LIMIT CASE WHEN @querylimit::int IS NULL THEN NULL ELSE @querylimit::int END;

-- name: CountUsersBySuperTeamID :one
SELECT COUNT(DISTINCT u.id)
FROM users u
INNER JOIN team_members tm ON u.id = tm.user_id
INNER JOIN teams t ON tm.team_id = t.id
WHERE t.super_team_id = @superteamid::char(28);

-- name: GetUserByPersonUUID :one
SELECT id, members_id, person_uuid, gender, church_id, church_locked_until, birthdate, email, name, first_name, last_name, middle_name, display_name, avatar_url, language, created_at
FROM users
WHERE person_uuid = @person_uuid::uuid;

-- name: UpdateUserPersonUUID :exec
UPDATE users
SET person_uuid = @person_uuid::uuid, updated_at = now()
WHERE id = @id::char(28);

-- name: GetUsersWithoutPersonUUID :many
SELECT id, members_id
FROM users
WHERE person_uuid IS NULL
  AND members_id ~ '^[0-9]+$'
ORDER BY id
LIMIT @querylimit::int;

-- name: UpdateUserLanguage :exec
UPDATE users
SET language = @language::text, updated_at = now()
WHERE id = @user_id::char(28);

-- name: GetUsersLeastRecentlySynced :many
-- Ordered by updated_at ASC so repeated calls (e.g. from a cron job) cycle
-- through the entire user base over time, not just a fixed subset. Excludes
-- members_id values that aren't a real Members API person ID (e.g. the
-- seed script's "MEM-<n>" placeholders) — those can never resolve via
-- Lookup, and since a failed sync never bumps updated_at, they'd otherwise
-- wedge permanently at the front of the queue.
-- A NULL querylimit means "no limit" (Postgres treats LIMIT NULL as no limit) — the whole
-- matching table is returned so a cron call can sweep everyone in one go.
SELECT id, members_id, person_uuid, gender, church_id, church_locked_until
FROM users
WHERE members_id ~ '^[0-9]+$'
ORDER BY updated_at ASC
LIMIT sqlc.narg('querylimit')::int;

-- name: TouchUserSyncedAt :exec
-- Bumps updated_at without changing any other column. Used when a Members sync attempt
-- fails with a definitive "this person no longer exists in Members" (404), as opposed to
-- a transient error (timeout, 5xx) — without this, a since-deleted member's row would
-- camp at the front of GetUsersLeastRecentlySynced's queue forever, since a failed sync
-- normally never advances updated_at, wasting a batch slot on it every single call.
UPDATE users
SET updated_at = now()
WHERE id = @id::char(28);

-- name: UpdateUserGenderAndChurch :exec
UPDATE users
SET
    gender = COALESCE(NULLIF(@gender::text, ''), gender),
    church_id = COALESCE(NULLIF(@church_id::text, ''), church_id),
    updated_at = now()
WHERE id = @id::char(28);

-- name: UpdateUserProfileFromMembers :exec
-- Empty string / NULL params leave the existing column value untouched,
-- so a partial Members API response never blanks out known-good data.
UPDATE users
SET
    email = COALESCE(NULLIF(@email::text, ''), email),
    name = COALESCE(NULLIF(@name::text, ''), name),
    first_name = COALESCE(NULLIF(@first_name::text, ''), first_name),
    last_name = COALESCE(NULLIF(@last_name::text, ''), last_name),
    middle_name = COALESCE(NULLIF(@middle_name::text, ''), middle_name),
    display_name = COALESCE(NULLIF(@display_name::text, ''), display_name),
    gender = COALESCE(NULLIF(@gender::text, ''), gender),
    birthdate = COALESCE(@birthdate::date, birthdate),
    church_id = COALESCE(NULLIF(@church_id::text, ''), church_id),
    updated_at = now()
WHERE id = @id::char(28);

-- name: LockUserChurch :exec
UPDATE users
SET church_locked_until = @locked_until::timestamptz, updated_at = now()
WHERE id = @id::char(28);

-- name: UnlockUserChurch :exec
UPDATE users
SET church_locked_until = NULL, updated_at = now()
WHERE id = @id::char(28);

-- name: GetUsersByPersonUUIDs :many
SELECT id, members_id, person_uuid, gender, church_id, church_locked_until,
       birthdate, email, name, first_name, last_name, middle_name, display_name,
       avatar_url, language, created_at
FROM users
WHERE person_uuid = ANY(@person_uuids::uuid[]);
