-- name: GetLeaderboardConfigByID :one
SELECT id, project_id, event_id, name, entity_type, filter, sort_order, is_active, created_at, updated_at, max_entries
FROM leaderboard_configs
WHERE id = @id::char(28);

-- name: GetLeaderboardConfigsByIDs :many
SELECT id, project_id, event_id, name, entity_type, filter, sort_order, is_active, created_at, updated_at, max_entries
FROM leaderboard_configs
WHERE id = ANY(@ids::char(28)[]);

-- name: GetLeaderboardConfigsByProjectIDs :many
-- Returns ALL configs (including inactive) for the given project IDs.
-- Non-admin visibility filtering must be done at the application layer.
SELECT id, project_id, event_id, name, entity_type, filter, sort_order, is_active, created_at, updated_at, max_entries
FROM leaderboard_configs
WHERE project_id = ANY(@project_ids::char(28)[])
ORDER BY project_id, sort_order, id;

-- name: GetLeaderboardConfigsByEventIDs :many
-- Returns ALL configs (including inactive) for the given event IDs.
-- Non-admin visibility filtering must be done at the application layer.
SELECT id, project_id, event_id, name, entity_type, filter, sort_order, is_active, created_at, updated_at, max_entries
FROM leaderboard_configs
WHERE event_id = ANY(@event_ids::char(28)[])
ORDER BY event_id, sort_order, id;

-- name: GetLeaderboardConfigsFilteredCursor :many
SELECT id, project_id, event_id, name, entity_type, filter, sort_order, is_active, created_at, updated_at, max_entries
FROM leaderboard_configs
WHERE
    (@ids::char(28)[] IS NULL OR id = ANY(@ids::char(28)[]))
    AND (@projectid::char(28) = '' OR project_id = @projectid::char(28))
    AND (@eventid::char(28) = '' OR event_id = @eventid::char(28))
    AND (sqlc.narg('isactive')::bool IS NULL OR is_active = sqlc.narg('isactive')::bool)
    AND (
        @aftercursorcreatedat::timestamptz IS NULL
        OR (created_at, id) < (@aftercursorcreatedat::timestamptz, @aftercursorid::char(28))
    )
    AND (
        @beforecursorcreatedat::timestamptz IS NULL
        OR (created_at, id) > (@beforecursorcreatedat::timestamptz, @beforecursorid::char(28))
    )
ORDER BY
    CASE WHEN @isbackward::bool = true THEN created_at END ASC,
    CASE WHEN @isbackward::bool = true THEN id END ASC,
    CASE WHEN @isbackward::bool = false OR @isbackward::bool IS NULL THEN created_at END DESC,
    CASE WHEN @isbackward::bool = false OR @isbackward::bool IS NULL THEN id END DESC
LIMIT CASE WHEN @querylimit::int IS NULL THEN NULL ELSE @querylimit::int END;

-- name: CountLeaderboardConfigsFiltered :one
SELECT COUNT(DISTINCT id)
FROM leaderboard_configs
WHERE
    (@ids::char(28)[] IS NULL OR id = ANY(@ids::char(28)[]))
    AND (@projectid::char(28) = '' OR project_id = @projectid::char(28))
    AND (@eventid::char(28) = '' OR event_id = @eventid::char(28))
    AND (sqlc.narg('isactive')::bool IS NULL OR is_active = sqlc.narg('isactive')::bool);

-- name: CreateLeaderboardConfig :one
INSERT INTO leaderboard_configs (
    id,
    project_id,
    event_id,
    name,
    entity_type,
    filter,
    max_entries,
    sort_order,
    is_active
)
VALUES (
    @id::text,
    @projectid::text,
    sqlc.narg('eventid')::text,
    @name::text,
    @entitytype::text,
    sqlc.narg('filter')::jsonb,
    sqlc.narg('maxentries')::int,
    COALESCE(sqlc.narg('sortorder')::int, 0),
    COALESCE(sqlc.narg('isactive')::bool, true)
)
RETURNING id, project_id, event_id, name, entity_type, filter, sort_order, is_active, created_at, updated_at, max_entries;

-- name: UpdateLeaderboardConfig :one
-- Full-replace update: the caller always sends the complete desired state.
-- Null filter/max_entries values clear those settings.
UPDATE leaderboard_configs
SET
    name = @name::text,
    entity_type = @entitytype::text,
    filter = @filter::jsonb,
    max_entries = sqlc.narg('maxentries')::int,
    sort_order = @sortorder::int,
    is_active = @isactive::bool
WHERE id = @id::char(28)
RETURNING id, project_id, event_id, name, entity_type, filter, sort_order, is_active, created_at, updated_at, max_entries;

-- name: DeleteLeaderboardConfig :exec
DELETE FROM leaderboard_configs
WHERE id = @id::char(28);
