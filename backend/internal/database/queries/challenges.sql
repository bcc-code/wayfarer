-- name: GetChallengesByIDs :many
SELECT id, project_id, event_id, challenge_type, name, description, image_url, url, button_text, published_at, visible_at, started_at, end_time, allow_self_completion, requires_team_membership, requires_super_team_membership, plugin_challenge_id, created_at, updated_at
FROM challenges
WHERE id = ANY(@ids::text[]);

-- name: GetChallengesByProjectIDs :many
SELECT id, project_id, event_id, challenge_type, name, description, image_url, url, button_text, published_at, visible_at, started_at, end_time, allow_self_completion, requires_team_membership, requires_super_team_membership, plugin_challenge_id, created_at, updated_at
FROM challenges
WHERE project_id = ANY(@project_ids::text[])
    AND published_at IS NOT NULL
    AND published_at <= NOW()
ORDER BY project_id, published_at DESC;

-- name: GetAllChallengesByProjectIDs :many
-- Returns ALL challenges for given project IDs, including unpublished.
-- Visibility filtering must be done at the application layer.
SELECT id, project_id, event_id, challenge_type, name, description, image_url, url, button_text, published_at, visible_at, started_at, end_time, allow_self_completion, requires_team_membership, requires_super_team_membership, plugin_challenge_id, created_at, updated_at
FROM challenges
WHERE project_id = ANY(@project_ids::text[])
ORDER BY project_id, COALESCE(published_at, created_at) DESC;

-- name: GetChallengesByEventIDs :many
SELECT id, project_id, event_id, challenge_type, name, description, image_url, url, button_text, published_at, visible_at, started_at, end_time, allow_self_completion, requires_team_membership, requires_super_team_membership, plugin_challenge_id, created_at, updated_at
FROM challenges
WHERE event_id = ANY(@event_ids::text[])
    AND published_at IS NOT NULL
    AND published_at <= NOW()
ORDER BY event_id, published_at DESC;

-- name: GetAllChallengesByEventIDs :many
-- Returns ALL challenges for given event IDs, including unpublished.
-- Visibility filtering must be done at the application layer.
SELECT id, project_id, event_id, challenge_type, name, description, image_url, url, button_text, published_at, visible_at, started_at, end_time, allow_self_completion, requires_team_membership, requires_super_team_membership, plugin_challenge_id, created_at, updated_at
FROM challenges
WHERE event_id = ANY(@event_ids::text[])
ORDER BY event_id, COALESCE(published_at, created_at) DESC;

-- name: GetChallengesFilteredCursor :many
SELECT id, project_id, event_id, challenge_type, name, description, image_url, url, button_text, published_at, visible_at, started_at, end_time, allow_self_completion, requires_team_membership, requires_super_team_membership, plugin_challenge_id, created_at, updated_at
FROM challenges
WHERE
    (@ids::text[] IS NULL OR id = ANY(@ids::text[]))
    AND (@projectid::text = '' OR project_id = @projectid::text)
    AND (@eventid::text = '' OR event_id = @eventid::text)
    AND (@challengetype::text = '' OR challenge_type = @challengetype::text)
    AND (@publishedafter::timestamptz IS NULL OR published_at >= @publishedafter::timestamptz)
    AND (@publishedbefore::timestamptz IS NULL OR published_at <= @publishedbefore::timestamptz)
    AND (@aftercursor::text = '' OR id > @aftercursor::text)
    AND (@beforecursor::text = '' OR id < @beforecursor::text)
ORDER BY
    CASE WHEN @isbackward::bool = true THEN id END DESC,
    CASE WHEN @isbackward::bool = false OR @isbackward::bool IS NULL THEN id END ASC
LIMIT CASE WHEN @querylimit::int IS NULL THEN NULL ELSE @querylimit::int END;

-- name: CountChallengesFiltered :one
SELECT COUNT(DISTINCT id)
FROM challenges
WHERE
    (@ids::text[] IS NULL OR id = ANY(@ids::text[]))
    AND (@projectid::text = '' OR project_id = @projectid::text)
    AND (@eventid::text = '' OR event_id = @eventid::text)
    AND (@challengetype::text = '' OR challenge_type = @challengetype::text)
    AND (@publishedafter::timestamptz IS NULL OR published_at >= @publishedafter::timestamptz)
    AND (@publishedbefore::timestamptz IS NULL OR published_at <= @publishedbefore::timestamptz);

-- name: CreateChallenge :one
INSERT INTO challenges (
    id,
    project_id,
    event_id,
    challenge_type,
    name,
    description,
    image_url,
    url,
    button_text,
    published_at,
    visible_at,
    end_time,
    allow_self_completion,
    requires_team_membership,
    requires_super_team_membership,
    plugin_challenge_id
)
VALUES (
    @id::text,
    @projectid::text,
    sqlc.narg('eventid')::text,
    @challengetype::text,
    @name::text,
    @description::text,
    sqlc.narg('imageurl')::text,
    sqlc.narg('url')::text,
    @buttontext::text,
    @publishedat::timestamptz,
    sqlc.narg('visibleat')::timestamptz,
    @endtime::timestamptz,
    COALESCE(sqlc.narg('allowselfcompletion')::bool, true),
    COALESCE(sqlc.narg('requiresteammembership')::bool, false),
    COALESCE(sqlc.narg('requiressuperteammembership')::bool, false),
    sqlc.narg('pluginchallengeid')::text
)
RETURNING id, project_id, event_id, challenge_type, name, description, image_url, url, button_text, published_at, visible_at, started_at, end_time, allow_self_completion, requires_team_membership, requires_super_team_membership, plugin_challenge_id, created_at, updated_at;

-- name: UpdateChallenge :one
UPDATE challenges
SET
    name = COALESCE(sqlc.narg('name')::text, name),
    description = COALESCE(sqlc.narg('description')::text, description),
    image_url = COALESCE(sqlc.narg('imageurl')::text, image_url),
    url = COALESCE(sqlc.narg('url')::text, url),
    button_text = COALESCE(sqlc.narg('buttontext')::text, button_text),
    event_id = COALESCE(sqlc.narg('eventid')::text, event_id),
    visible_at = COALESCE(sqlc.narg('visibleat')::timestamptz, visible_at),
    started_at = COALESCE(sqlc.narg('startedat')::timestamptz, started_at),
    end_time = COALESCE(sqlc.narg('endtime')::timestamptz, end_time),
    allow_self_completion = COALESCE(sqlc.narg('allowselfcompletion')::bool, allow_self_completion),
    requires_team_membership = COALESCE(sqlc.narg('requiresteammembership')::bool, requires_team_membership),
    requires_super_team_membership = COALESCE(sqlc.narg('requiressuperteammembership')::bool, requires_super_team_membership),
    plugin_challenge_id = COALESCE(sqlc.narg('pluginchallengeid')::text, plugin_challenge_id),
    updated_at = now()
WHERE id = @id::text
RETURNING id, project_id, event_id, challenge_type, name, description, image_url, url, button_text, published_at, visible_at, started_at, end_time, allow_self_completion, requires_team_membership, requires_super_team_membership, plugin_challenge_id, created_at, updated_at;

-- name: DeleteChallenge :exec
DELETE FROM challenges
WHERE id = @id::text;

-- name: PublishChallenge :one
UPDATE challenges
SET
    published_at = @publishedat::timestamptz,
    updated_at = now()
WHERE id = @id::text
RETURNING id, project_id, event_id, challenge_type, name, description, image_url, url, button_text, published_at, visible_at, started_at, end_time, allow_self_completion, requires_team_membership, requires_super_team_membership, plugin_challenge_id, created_at, updated_at;

-- name: AssignChallengeToEvent :one
UPDATE challenges
SET
    event_id = @eventid::text,
    updated_at = now()
WHERE id = @id::text
RETURNING id, project_id, event_id, challenge_type, name, description, image_url, url, button_text, published_at, visible_at, started_at, end_time, allow_self_completion, requires_team_membership, requires_super_team_membership, plugin_challenge_id, created_at, updated_at;

-- name: BulkPublishChallenges :many
UPDATE challenges
SET
    published_at = @publishedat::timestamptz,
    updated_at = now()
WHERE id = ANY(@ids::text[])
RETURNING id, project_id, event_id, challenge_type, name, description, image_url, url, button_text, published_at, visible_at, started_at, end_time, allow_self_completion, requires_team_membership, requires_super_team_membership, plugin_challenge_id, created_at, updated_at;

-- name: BulkCreateChallenges :many
INSERT INTO challenges (
    id,
    project_id,
    event_id,
    challenge_type,
    name,
    description,
    image_url,
    url,
    button_text,
    published_at,
    visible_at,
    end_time,
    allow_self_completion,
    requires_team_membership,
    requires_super_team_membership,
    plugin_challenge_id
)
SELECT
    unnest(@ids::text[]),
    unnest(@projectids::text[]),
    unnest(@eventids::text[]),
    unnest(@challengetypes::text[]),
    unnest(@names::text[]),
    unnest(@descriptions::text[]),
    unnest(@imageurls::text[]),
    unnest(@urls::text[]),
    unnest(@buttontexts::text[]),
    unnest(@publishedats::timestamptz[]),
    unnest(@visibleats::timestamptz[]),
    unnest(@endtimes::timestamptz[]),
    unnest(@allowselfcompletions::bool[]),
    unnest(@requiresteammemberships::bool[]),
    unnest(@requiressuperteammemberships::bool[]),
    unnest(@pluginchallengeids::text[])
RETURNING id, project_id, event_id, challenge_type, name, description, image_url, url, button_text, published_at, visible_at, started_at, end_time, allow_self_completion, requires_team_membership, requires_super_team_membership, plugin_challenge_id, created_at, updated_at;

-- name: UpdateChallengeTimestamps :one
UPDATE challenges
SET
    visible_at = COALESCE(sqlc.narg('visibleat')::timestamptz, visible_at),
    started_at = COALESCE(sqlc.narg('startedat')::timestamptz, started_at),
    updated_at = now()
WHERE id = @id::text
RETURNING id, project_id, event_id, challenge_type, name, description, image_url, url, button_text, published_at, visible_at, started_at, end_time, allow_self_completion, requires_team_membership, requires_super_team_membership, plugin_challenge_id, created_at, updated_at;

-- name: UpdateChallengeRequirements :one
UPDATE challenges
SET
    requires_team_membership = COALESCE(sqlc.narg('requiresteammembership')::bool, requires_team_membership),
    requires_super_team_membership = COALESCE(sqlc.narg('requiressuperteammembership')::bool, requires_super_team_membership),
    updated_at = now()
WHERE id = @id::text
RETURNING id, project_id, event_id, challenge_type, name, description, image_url, url, button_text, published_at, visible_at, started_at, end_time, allow_self_completion, requires_team_membership, requires_super_team_membership, plugin_challenge_id, created_at, updated_at;

-- name: GetQuizByChallengeID :one
SELECT * FROM quizzes WHERE challenge_id = @challengeid::text;

-- name: GetChallengeByPluginChallengeID :one
SELECT id, project_id, event_id, challenge_type, name, description, image_url, url, button_text, published_at, visible_at, started_at, end_time, allow_self_completion, requires_team_membership, requires_super_team_membership, plugin_challenge_id, created_at, updated_at
FROM challenges
WHERE plugin_challenge_id = @pluginchallengeid::text;
