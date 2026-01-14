-- name: GetAchievementsByIDs :many
SELECT
    a.id,
    a.achievement_type,
    a.project_id,
    a.event_id,
    a.challenge_id,
    a.name,
    a.description_pending,
    a.description_completed,
    a.notification_text,
    a.image_pending,
    a.image_completed,
    a.points,
    a.hidden,
    a.awardable_from,
    a.created_at,
    a.updated_at,
    -- Content achievement data
    ca.achievement_id AS content_achievement_id,
    COALESCE(
        (SELECT jsonb_agg(
            jsonb_build_object(
                'id', cai.id,
                'external_content_id', cai.external_content_id,
                'sort_order', cai.sort_order
            ) ORDER BY cai.sort_order
        )
        FROM content_achievement_items cai
        WHERE cai.achievement_id = a.id),
        '[]'::jsonb
    ) AS content_items,
    -- Streak achievement data
    sa.streak_id,
    sa.needed_streak
FROM achievements a
LEFT JOIN content_achievements ca ON a.id = ca.achievement_id
LEFT JOIN streak_achievements sa ON a.id = sa.achievement_id
WHERE a.id = ANY(@ids::text[]);

-- name: GetAchievementsByProjectIDs :many
SELECT
    a.id,
    a.achievement_type,
    a.project_id,
    a.event_id,
    a.challenge_id,
    a.name,
    a.description_pending,
    a.description_completed,
    a.notification_text,
    a.image_pending,
    a.image_completed,
    a.points,
    a.hidden,
    a.awardable_from,
    a.created_at,
    a.updated_at,
    -- Type-specific fields needed for model construction
    ca.achievement_id AS content_achievement_id,
    sa.streak_id,
    sa.needed_streak
FROM achievements a
LEFT JOIN content_achievements ca ON a.id = ca.achievement_id
LEFT JOIN streak_achievements sa ON a.id = sa.achievement_id
WHERE a.project_id = ANY(@project_ids::text[])
    AND a.hidden = false
ORDER BY a.project_id, a.sort_order, a.created_at DESC;

-- name: GetAllAchievementsByProjectID :many
-- Returns all achievements for a project including hidden ones, ordered by sort_order
SELECT
    a.id,
    a.achievement_type,
    a.project_id,
    a.event_id,
    a.challenge_id,
    a.name,
    a.description_pending,
    a.description_completed,
    a.notification_text,
    a.image_pending,
    a.image_completed,
    a.points,
    a.hidden,
    a.awardable_from,
    a.sort_order,
    a.created_at,
    a.updated_at,
    ca.achievement_id AS content_achievement_id,
    sa.streak_id,
    sa.needed_streak
FROM achievements a
LEFT JOIN content_achievements ca ON a.id = ca.achievement_id
LEFT JOIN streak_achievements sa ON a.id = sa.achievement_id
WHERE a.project_id = @project_id::text
ORDER BY a.sort_order, a.created_at DESC;

-- name: GetAchievementsFilteredCursor :many
SELECT
    a.id,
    a.achievement_type,
    a.project_id,
    a.event_id,
    a.challenge_id,
    a.name,
    a.description_pending,
    a.description_completed,
    a.notification_text,
    a.image_pending,
    a.image_completed,
    a.points,
    a.hidden,
    a.awardable_from,
    a.sort_order,
    a.created_at,
    a.updated_at,
    -- Content achievement data
    ca.achievement_id AS content_achievement_id,
    COALESCE(
        (SELECT jsonb_agg(
            jsonb_build_object(
                'id', cai.id,
                'external_content_id', cai.external_content_id,
                'sort_order', cai.sort_order
            ) ORDER BY cai.sort_order
        )
        FROM content_achievement_items cai
        WHERE cai.achievement_id = a.id),
        '[]'::jsonb
    ) AS content_items,
    -- Streak achievement data
    sa.streak_id,
    sa.needed_streak
FROM achievements a
LEFT JOIN content_achievements ca ON a.id = ca.achievement_id
LEFT JOIN streak_achievements sa ON a.id = sa.achievement_id
WHERE
    (@ids::text[] IS NULL OR a.id = ANY(@ids::text[]))
    AND (@projectid::text = '' OR a.project_id = @projectid::text)
    AND (@eventid::text = '' OR a.event_id = @eventid::text)
    AND (@aftercursor::text = '' OR (a.sort_order, a.id) > (SELECT sort_order, id FROM achievements WHERE id = @aftercursor::text))
    AND (@beforecursor::text = '' OR (a.sort_order, a.id) < (SELECT sort_order, id FROM achievements WHERE id = @beforecursor::text))
ORDER BY
    CASE WHEN @isbackward::bool = true THEN a.sort_order END DESC,
    CASE WHEN @isbackward::bool = false OR @isbackward::bool IS NULL THEN a.sort_order END ASC,
    CASE WHEN @isbackward::bool = true THEN a.id END DESC,
    CASE WHEN @isbackward::bool = false OR @isbackward::bool IS NULL THEN a.id END ASC
LIMIT CASE WHEN @querylimit::int IS NULL THEN NULL ELSE @querylimit::int END;

-- name: CountAchievementsFiltered :one
SELECT COUNT(DISTINCT a.id)
FROM achievements a
WHERE
    (@ids::text[] IS NULL OR a.id = ANY(@ids::text[]))
    AND (@projectid::text = '' OR a.project_id = @projectid::text)
    AND (@eventid::text = '' OR a.event_id = @eventid::text);

-- name: GetContentItemsByAchievementIDs :many
SELECT id, achievement_id, external_content_id, sort_order
FROM content_achievement_items
WHERE achievement_id = ANY(@achievement_ids::text[])
ORDER BY achievement_id, sort_order;

-- ==================== Create Operations ====================

-- name: CreateAchievement :one
INSERT INTO achievements (
    id,
    achievement_type,
    project_id,
    event_id,
    challenge_id,
    name,
    description_pending,
    description_completed,
    notification_text,
    image_pending,
    image_completed,
    points,
    hidden,
    awardable_from
) VALUES (
    @id::text,
    @achievement_type::text,
    @project_id::text,
    NULLIF(@event_id::text, ''),
    NULLIF(@challenge_id::text, ''),
    @name::text,
    @description_pending::text,
    @description_completed::text,
    @notification_text::text,
    @image_pending::text,
    @image_completed::text,
    @points::int,
    @hidden::bool,
    sqlc.narg('awardable_from')::timestamptz
) RETURNING *;

-- name: CreateContentAchievementJunction :exec
INSERT INTO content_achievements (achievement_id)
VALUES (@achievement_id::text);

-- name: CreateContentAchievementItem :one
INSERT INTO content_achievement_items (
    id,
    achievement_id,
    external_content_id,
    sort_order
) VALUES (
    @id::text,
    @achievement_id::text,
    @external_content_id::text,
    @sort_order::int
) RETURNING *;

-- name: CreateStreakAchievementData :exec
INSERT INTO streak_achievements (
    achievement_id,
    streak_id,
    needed_streak
) VALUES (
    @achievement_id::text,
    @streak_id::text,
    @needed_streak::int
);

-- ==================== Update Operations ====================

-- name: UpdateAchievement :one
UPDATE achievements
SET
    name = CASE WHEN sqlc.narg('name')::text IS NOT NULL THEN sqlc.narg('name')::text ELSE name END,
    description_pending = CASE WHEN sqlc.narg('description_pending')::text IS NOT NULL THEN sqlc.narg('description_pending')::text ELSE description_pending END,
    description_completed = CASE WHEN sqlc.narg('description_completed')::text IS NOT NULL THEN sqlc.narg('description_completed')::text ELSE description_completed END,
    notification_text = CASE WHEN sqlc.narg('notification_text')::text IS NOT NULL THEN sqlc.narg('notification_text')::text ELSE notification_text END,
    image_pending = CASE WHEN sqlc.narg('image_pending')::text IS NOT NULL THEN sqlc.narg('image_pending')::text ELSE image_pending END,
    image_completed = CASE WHEN sqlc.narg('image_completed')::text IS NOT NULL THEN sqlc.narg('image_completed')::text ELSE image_completed END,
    event_id = CASE WHEN sqlc.narg('event_id')::text IS NOT NULL THEN sqlc.narg('event_id')::text ELSE event_id END,
    challenge_id = CASE WHEN sqlc.narg('challenge_id')::text IS NOT NULL THEN sqlc.narg('challenge_id')::text ELSE challenge_id END,
    points = CASE WHEN sqlc.narg('points')::int IS NOT NULL THEN sqlc.narg('points')::int ELSE points END,
    hidden = CASE WHEN sqlc.narg('hidden')::bool IS NOT NULL THEN sqlc.narg('hidden')::bool ELSE hidden END,
    awardable_from = CASE WHEN sqlc.narg('awardable_from')::timestamptz IS NOT NULL THEN sqlc.narg('awardable_from')::timestamptz ELSE awardable_from END,
    updated_at = now()
WHERE id = @id::text
RETURNING *;

-- name: DeleteContentAchievementItems :exec
DELETE FROM content_achievement_items
WHERE achievement_id = @achievement_id::text;

-- name: UpdateStreakAchievementData :exec
UPDATE streak_achievements
SET
    streak_id = CASE WHEN sqlc.narg('streak_id')::text IS NOT NULL THEN sqlc.narg('streak_id')::text ELSE streak_id END,
    needed_streak = CASE WHEN sqlc.narg('needed_streak')::int IS NOT NULL THEN sqlc.narg('needed_streak')::int ELSE needed_streak END
WHERE achievement_id = @achievement_id::text;

-- name: UpdateAchievementSortOrder :exec
UPDATE achievements
SET sort_order = @sort_order::int, updated_at = now()
WHERE id = @id::text;

-- ==================== Delete Operations ====================

-- name: DeleteAchievement :exec
DELETE FROM achievements
WHERE id = @id::text;

-- ==================== Award/Revoke Operations ====================

-- name: AwardUserAchievement :exec
INSERT INTO user_achievements (user_id, achievement_id, achieved_at)
VALUES (@user_id::text, @achievement_id::text, COALESCE(@achieved_at::timestamptz, now()));

-- name: AwardUserAchievementIdempotent :exec
-- Awards achievement to user, silently ignores if already awarded
INSERT INTO user_achievements (user_id, achievement_id, achieved_at)
VALUES (@user_id::text, @achievement_id::text, COALESCE(@achieved_at::timestamptz, now()))
ON CONFLICT (user_id, achievement_id) DO NOTHING;

-- name: CheckUserHasAchievement :one
-- Check if a user already has an achievement
SELECT EXISTS(
    SELECT 1 FROM user_achievements
    WHERE user_id = @user_id::text AND achievement_id = @achievement_id::text
) AS has_achievement;

-- name: AwardTeamAchievementBatch :exec
-- Award achievement to all members of a team in a single query
INSERT INTO user_achievements (user_id, achievement_id, achieved_at)
SELECT
    tm.user_id,
    @achievement_id::text,
    COALESCE(@achieved_at::timestamptz, now())
FROM team_members tm
WHERE tm.team_id = @team_id::text
ON CONFLICT (user_id, achievement_id) DO NOTHING;

-- name: AwardSuperTeamAchievementBatch :exec
-- Award achievement to all members of a superteam (through all teams in superteam) in a single query
INSERT INTO user_achievements (user_id, achievement_id, achieved_at)
SELECT
    tm.user_id,
    @achievement_id::text,
    COALESCE(@achieved_at::timestamptz, now())
FROM teams t
INNER JOIN team_members tm ON tm.team_id = t.id
WHERE t.super_team_id = @super_team_id::text
ON CONFLICT (user_id, achievement_id) DO NOTHING;

-- name: RevokeUserAchievement :exec
DELETE FROM user_achievements
WHERE user_id = @user_id::text
    AND achievement_id = @achievement_id::text;

-- name: RevokeTeamAchievementBatch :exec
-- Revoke achievement from all members of a team in a single query
DELETE FROM user_achievements
WHERE achievement_id = @achievement_id::text
  AND user_id IN (
    SELECT tm.user_id
    FROM team_members tm
    WHERE tm.team_id = @team_id::text
  );

-- name: RevokeSuperTeamAchievementBatch :exec
-- Revoke achievement from all members of a superteam in a single query
DELETE FROM user_achievements
WHERE achievement_id = @achievement_id::text
  AND user_id IN (
    SELECT tm.user_id
    FROM teams t
    INNER JOIN team_members tm ON tm.team_id = t.id
    WHERE t.super_team_id = @super_team_id::text
  );

-- name: GetUserAchievementTimestamps :many
-- Get achieved_at timestamps for a user's achievements (for achievedAt field resolution)
SELECT achievement_id, achieved_at
FROM user_achievements
WHERE user_id = @userid::text
  AND achievement_id = ANY(@achievement_ids::text[]);

-- name: GetBulkUserAchievementTimestamps :many
SELECT user_id, achievement_id, achieved_at
FROM user_achievements
WHERE (user_id, achievement_id) IN (
    SELECT unnest(@user_ids::text[]), unnest(@achievement_ids::text[])
);

-- ==================== Content Progress Operations ====================

-- name: GetUserContentProgress :many
SELECT user_id, achievement_id, external_content_id, completed_at
FROM user_content_progress
WHERE user_id = @user_id::text
  AND achievement_id = ANY(@achievement_ids::text[]);

-- name: GetUserContentProgressForAchievement :many
SELECT user_id, achievement_id, external_content_id, completed_at
FROM user_content_progress
WHERE user_id = @user_id::text
  AND achievement_id = @achievement_id::text;

-- name: MarkContentItemCompleted :exec
INSERT INTO user_content_progress (user_id, achievement_id, external_content_id, completed_at)
VALUES (@user_id::text, @achievement_id::text, @external_content_id::text, COALESCE(@completed_at::timestamptz, now()))
ON CONFLICT (user_id, achievement_id, external_content_id) DO NOTHING;

-- name: UnmarkContentItemCompleted :exec
DELETE FROM user_content_progress
WHERE user_id = @user_id::text
  AND achievement_id = @achievement_id::text
  AND external_content_id = @external_content_id::text;

-- name: GetPublishedContentAchievementsByExternalContent :many
-- Get all content achievements that contain a specific external content
SELECT DISTINCT
    a.id,
    a.achievement_type,
    a.project_id,
    a.event_id,
    a.challenge_id,
    a.name,
    a.description_pending,
    a.description_completed,
    a.notification_text,
    a.image_pending,
    a.image_completed,
    a.points,
    a.hidden,
    a.awardable_from,
    a.created_at,
    a.updated_at,
    COALESCE(
        (SELECT jsonb_agg(
            jsonb_build_object(
                'id', cai2.id,
                'external_content_id', cai2.external_content_id,
                'sort_order', cai2.sort_order
            ) ORDER BY cai2.sort_order
        )
        FROM content_achievement_items cai2
        WHERE cai2.achievement_id = a.id),
        '[]'::jsonb
    ) AS content_items
FROM achievements a
INNER JOIN content_achievements ca ON a.id = ca.achievement_id
INNER JOIN content_achievement_items cai ON ca.achievement_id = cai.achievement_id
WHERE cai.external_content_id = @external_content_id::text;

-- name: MarkContentItemCompletedForAllAchievements :exec
-- Mark content completed for a user across all achievements containing this content
INSERT INTO user_content_progress (user_id, achievement_id, external_content_id, completed_at)
SELECT @user_id::text, ca.achievement_id, @external_content_id::text, now()
FROM content_achievements ca
INNER JOIN content_achievement_items cai ON ca.achievement_id = cai.achievement_id
WHERE cai.external_content_id = @external_content_id::text
ON CONFLICT (user_id, achievement_id, external_content_id) DO NOTHING;

-- name: UnmarkContentItemCompletedForAllAchievements :exec
-- Unmark content completed for a user across all achievements containing this content
DELETE FROM user_content_progress
WHERE user_id = @user_id::text
  AND external_content_id = @external_content_id::text;

-- name: GetContentItemsWithExternalContent :many
-- Get content items with external content joined
SELECT
    cai.id,
    cai.achievement_id,
    cai.external_content_id,
    cai.sort_order,
    ec.plan_id AS external_plan_id,
    ec.task_id AS external_task_id,
    ec.content_id AS external_content_id_value,
    ec.content_type AS external_content_type,
    ec.published_at AS external_published_at,
    ec.source AS external_source
FROM content_achievement_items cai
INNER JOIN external_content ec ON cai.external_content_id = ec.id
WHERE cai.achievement_id = ANY(@achievementids::text[])
ORDER BY cai.achievement_id, cai.sort_order;

-- name: CheckContentItemInAchievement :one
-- Check if a content item exists in a specific achievement
SELECT EXISTS(
    SELECT 1 FROM content_achievement_items
    WHERE achievement_id = @achievement_id::text
      AND external_content_id = @external_content_id::text
) AS exists;

-- name: GetAchievementByID :one
-- Get a single achievement by ID with all details
SELECT
    a.id,
    a.achievement_type,
    a.project_id,
    a.event_id,
    a.challenge_id,
    a.name,
    a.description_pending,
    a.description_completed,
    a.notification_text,
    a.image_pending,
    a.image_completed,
    a.points,
    a.hidden,
    a.awardable_from,
    a.sort_order,
    a.created_at,
    a.updated_at
FROM achievements a
WHERE a.id = @id::text;
