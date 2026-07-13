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
    sa.achievement_id AS streak_achievement_id,
    COALESCE(
        (SELECT jsonb_agg(
            jsonb_build_object(
                'id', sai.id,
                'external_content_id', sai.external_content_id,
                'sort_order', sai.sort_order
            ) ORDER BY sai.sort_order
        )
        FROM streak_achievement_items sai
        WHERE sai.achievement_id = a.id),
        '[]'::jsonb
    ) AS streak_items,
    -- Quiz achievement data
    qa.quiz_id,
    qa.min_score_percentage,
    qa.require_completion
FROM achievements a
LEFT JOIN content_achievements ca ON a.id = ca.achievement_id
LEFT JOIN streak_achievements sa ON a.id = sa.achievement_id
LEFT JOIN quiz_achievements qa ON a.id = qa.achievement_id
WHERE a.id = ANY(@ids::char(28)[]);

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
    sa.achievement_id AS streak_achievement_id,
    -- Quiz achievement data
    qa.quiz_id,
    qa.min_score_percentage,
    qa.require_completion
FROM achievements a
LEFT JOIN content_achievements ca ON a.id = ca.achievement_id
LEFT JOIN streak_achievements sa ON a.id = sa.achievement_id
LEFT JOIN quiz_achievements qa ON a.id = qa.achievement_id
WHERE a.project_id = ANY(@project_ids::char(28)[])
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
    sa.achievement_id AS streak_achievement_id,
    -- Quiz achievement data
    qa.quiz_id,
    qa.min_score_percentage,
    qa.require_completion
FROM achievements a
LEFT JOIN content_achievements ca ON a.id = ca.achievement_id
LEFT JOIN streak_achievements sa ON a.id = sa.achievement_id
LEFT JOIN quiz_achievements qa ON a.id = qa.achievement_id
WHERE a.project_id = @project_id::char(28)
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
    sa.achievement_id AS streak_achievement_id,
    COALESCE(
        (SELECT jsonb_agg(
            jsonb_build_object(
                'id', sai.id,
                'external_content_id', sai.external_content_id,
                'sort_order', sai.sort_order
            ) ORDER BY sai.sort_order
        )
        FROM streak_achievement_items sai
        WHERE sai.achievement_id = a.id),
        '[]'::jsonb
    ) AS streak_items,
    -- Quiz achievement data
    qa.quiz_id,
    qa.min_score_percentage,
    qa.require_completion
FROM achievements a
LEFT JOIN content_achievements ca ON a.id = ca.achievement_id
LEFT JOIN streak_achievements sa ON a.id = sa.achievement_id
LEFT JOIN quiz_achievements qa ON a.id = qa.achievement_id
WHERE
    (@ids::char(28)[] IS NULL OR a.id = ANY(@ids::char(28)[]))
    AND (@projectid::char(28) = '' OR a.project_id = @projectid::char(28))
    AND (@eventid::char(28) = '' OR a.event_id = @eventid::char(28))
    AND (@aftercursor::char(28) = '' OR (a.sort_order, a.id) > (SELECT sort_order, id FROM achievements WHERE id = @aftercursor::char(28)))
    AND (@beforecursor::char(28) = '' OR (a.sort_order, a.id) < (SELECT sort_order, id FROM achievements WHERE id = @beforecursor::char(28)))
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
    (@ids::char(28)[] IS NULL OR a.id = ANY(@ids::char(28)[]))
    AND (@projectid::char(28) = '' OR a.project_id = @projectid::char(28))
    AND (@eventid::char(28) = '' OR a.event_id = @eventid::char(28));

-- name: GetContentItemsByAchievementIDs :many
SELECT id, achievement_id, external_content_id, sort_order
FROM content_achievement_items
WHERE achievement_id = ANY(@achievement_ids::char(28)[])
ORDER BY achievement_id, sort_order;

-- name: GetAchievementCompletionStatus :many
SELECT
    cai.achievement_id,
    COUNT(DISTINCT cai.external_content_id)::int AS item_count,
    COUNT(DISTINCT ucp.external_content_id)::int AS progress_count
FROM content_achievement_items cai
LEFT JOIN user_content_progress ucp
    ON ucp.achievement_id = cai.achievement_id
    AND ucp.user_id = @user_id::char(28)
    AND ucp.external_content_id = cai.external_content_id
WHERE cai.achievement_id = ANY(@achievement_ids::char(28)[])
GROUP BY cai.achievement_id;

-- name: GetContentItemCounts :many
-- Get content item counts per achievement (for caching)
SELECT achievement_id, COUNT(*)::int AS item_count
FROM content_achievement_items
WHERE achievement_id = ANY(@achievement_ids::char(28)[])
GROUP BY achievement_id;

-- name: GetUserProgressCounts :many
-- Get user progress counts per achievement
SELECT achievement_id, COUNT(*)::int AS progress_count
FROM user_content_progress
WHERE user_id = @user_id::char(28)
  AND achievement_id = ANY(@achievement_ids::char(28)[])
GROUP BY achievement_id;

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

-- name: CreateStreakAchievementJunction :exec
INSERT INTO streak_achievements (achievement_id)
VALUES (@achievement_id::text);

-- name: CreateStreakAchievementItem :one
INSERT INTO streak_achievement_items (
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
WHERE id = @id::char(28)
RETURNING *;

-- name: DeleteContentAchievementItems :exec
DELETE FROM content_achievement_items
WHERE achievement_id = @achievement_id::char(28);

-- name: DeleteStreakAchievementItems :exec
DELETE FROM streak_achievement_items
WHERE achievement_id = @achievement_id::char(28);

-- name: UpdateAchievementSortOrder :exec
UPDATE achievements
SET sort_order = @sort_order::int, updated_at = now()
WHERE id = @id::char(28);

-- ==================== Delete Operations ====================

-- name: DeleteAchievement :exec
DELETE FROM achievements
WHERE id = @id::char(28);

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
    WHERE user_id = @user_id::char(28) AND achievement_id = @achievement_id::char(28)
) AS has_achievement;

-- name: GetUserAwardedAchievementIDs :many
-- Get which achievements from a list the user already has
SELECT achievement_id
FROM user_achievements
WHERE user_id = @user_id::char(28)
  AND achievement_id = ANY(@achievement_ids::char(28)[]);

-- name: AwardTeamAchievementBatch :exec
-- Award achievement to all members of a team in a single query
INSERT INTO user_achievements (user_id, achievement_id, achieved_at)
SELECT
    tm.user_id,
    @achievement_id::text,
    COALESCE(@achieved_at::timestamptz, now())
FROM team_members tm
WHERE tm.team_id = @team_id::char(28)
ON CONFLICT (user_id, achievement_id) DO NOTHING;

-- name: AwardUserAchievementsBatch :many
-- Award achievement to multiple users in a single query, returns only newly inserted rows
INSERT INTO user_achievements (user_id, achievement_id, achieved_at)
SELECT
    unnest(@user_ids::text[]),
    @achievement_id::text,
    COALESCE(@achieved_at::timestamptz, now())
ON CONFLICT (user_id, achievement_id) DO NOTHING
RETURNING user_id, achievement_id, achieved_at;

-- name: AwardSuperTeamAchievementBatch :exec
-- Award achievement to all members of a superteam (through all teams in superteam) in a single query
INSERT INTO user_achievements (user_id, achievement_id, achieved_at)
SELECT
    tm.user_id,
    @achievement_id::text,
    COALESCE(@achieved_at::timestamptz, now())
FROM teams t
INNER JOIN team_members tm ON tm.team_id = t.id
WHERE t.super_team_id = @super_team_id::char(28)
ON CONFLICT (user_id, achievement_id) DO NOTHING;

-- name: RevokeUserAchievement :exec
DELETE FROM user_achievements
WHERE user_id = @user_id::char(28)
    AND achievement_id = @achievement_id::char(28);

-- name: RevokeTeamAchievementBatch :exec
-- Revoke achievement from all members of a team in a single query
DELETE FROM user_achievements
WHERE achievement_id = @achievement_id::char(28)
  AND user_id IN (
    SELECT tm.user_id
    FROM team_members tm
    WHERE tm.team_id = @team_id::char(28)
  );

-- name: RevokeSuperTeamAchievementBatch :exec
-- Revoke achievement from all members of a superteam in a single query
DELETE FROM user_achievements
WHERE achievement_id = @achievement_id::char(28)
  AND user_id IN (
    SELECT tm.user_id
    FROM teams t
    INNER JOIN team_members tm ON tm.team_id = t.id
    WHERE t.super_team_id = @super_team_id::char(28)
  );

-- name: GetUserAchievementTimestamps :many
-- Get achieved_at timestamps for a user's achievements (for achievedAt field resolution)
SELECT achievement_id, achieved_at
FROM user_achievements
WHERE user_id = @userid::char(28)
  AND achievement_id = ANY(@achievement_ids::char(28)[]);

-- name: GetBulkUserAchievementTimestamps :many
SELECT user_id, achievement_id, achieved_at
FROM user_achievements
WHERE (user_id, achievement_id) IN (
    SELECT unnest(@user_ids::char(28)[]), unnest(@achievement_ids::char(28)[])
);

-- name: GetBulkUserAchievementCelebratedTimestamps :many
SELECT user_id, achievement_id, celebrated_at
FROM user_achievements
WHERE (user_id, achievement_id) IN (
    SELECT unnest(@user_ids::char(28)[]), unnest(@achievement_ids::char(28)[])
);

-- name: MarkAchievementCelebrated :exec
UPDATE user_achievements
SET celebrated_at = now()
WHERE user_id = @user_id::char(28)
  AND achievement_id = @achievement_id::char(28)
  AND celebrated_at IS NULL;

-- ==================== Content Progress Operations ====================

-- name: GetBulkUserContentProgress :many
SELECT user_id, achievement_id, external_content_id, completed_at
FROM user_content_progress
WHERE (user_id, achievement_id) IN (
    SELECT unnest(@user_ids::char(28)[]), unnest(@achievement_ids::char(28)[])
);

-- name: GetUserContentProgress :many
SELECT user_id, achievement_id, external_content_id, completed_at
FROM user_content_progress
WHERE user_id = @user_id::char(28)
  AND achievement_id = ANY(@achievement_ids::char(28)[]);

-- name: MarkContentItemCompleted :exec
INSERT INTO user_content_progress (user_id, achievement_id, external_content_id, completed_at)
VALUES (@user_id::text, @achievement_id::text, @external_content_id::text, COALESCE(@completed_at::timestamptz, now()))
ON CONFLICT (user_id, achievement_id, external_content_id) DO NOTHING;

-- name: UnmarkContentItemCompleted :exec
DELETE FROM user_content_progress
WHERE user_id = @user_id::char(28)
  AND achievement_id = @achievement_id::char(28)
  AND external_content_id = @external_content_id::char(28);

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
WHERE cai.external_content_id = @external_content_id::char(28);

-- name: MarkContentItemCompletedForAllAchievements :exec
-- Mark content completed for a user across all achievements containing this content
INSERT INTO user_content_progress (user_id, achievement_id, external_content_id, completed_at)
SELECT @user_id::text, ca.achievement_id, @external_content_id::char(28), now()
FROM content_achievements ca
INNER JOIN content_achievement_items cai ON ca.achievement_id = cai.achievement_id
WHERE cai.external_content_id = @external_content_id::char(28)
ON CONFLICT (user_id, achievement_id, external_content_id) DO NOTHING;

-- name: UnmarkContentItemCompletedForAllAchievements :exec
-- Unmark content completed for a user across all achievements containing this content
DELETE FROM user_content_progress
WHERE user_id = @user_id::char(28)
  AND external_content_id = @external_content_id::char(28);

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
WHERE cai.achievement_id = ANY(@achievementids::char(28)[])
ORDER BY cai.achievement_id, cai.sort_order;

-- name: CheckContentItemInAchievement :one
-- Check if a content item exists in a specific achievement
SELECT EXISTS(
    SELECT 1 FROM content_achievement_items
    WHERE achievement_id = @achievement_id::char(28)
      AND external_content_id = @external_content_id::char(28)
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
WHERE a.id = @id::char(28);

-- name: GetContentAchievementForAward :one
-- Get content achievement data for awarding (same fields as GetPublishedContentAchievementsByExternalContent)
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
    ) AS content_items
FROM achievements a
INNER JOIN content_achievements ca ON ca.achievement_id = a.id
WHERE a.id = @achievement_id::char(28) AND a.project_id = @project_id::char(28);

-- name: GetUsersWithUnclaimedContentAchievement :many
-- Find users who completed all items for a content achievement but weren't awarded
SELECT DISTINCT u.id AS user_id
FROM users u
INNER JOIN achievements a ON a.id = @achievement_id::char(28)
INNER JOIN content_achievements ca ON ca.achievement_id = a.id
WHERE a.project_id = @project_id::char(28)
  AND NOT EXISTS (
    SELECT 1 FROM user_achievements ua
    WHERE ua.user_id = u.id AND ua.achievement_id = a.id
  )
  AND NOT EXISTS (
    SELECT 1 FROM content_achievement_items cai
    WHERE cai.achievement_id = a.id
    AND NOT EXISTS (
      SELECT 1 FROM user_content_progress ucp
      WHERE ucp.user_id = u.id
      AND ucp.achievement_id = a.id
      AND ucp.external_content_id = cai.external_content_id
    )
  )
  AND EXISTS (
    SELECT 1 FROM content_achievement_items cai
    WHERE cai.achievement_id = a.id
  );

-- ==================== Streak Progress Operations ====================

-- name: GetStreakItemsByAchievementIDs :many
SELECT id, achievement_id, external_content_id, sort_order
FROM streak_achievement_items
WHERE achievement_id = ANY(@achievement_ids::char(28)[])
ORDER BY achievement_id, sort_order;

-- name: GetStreakAchievementCompletionStatus :many
SELECT
    sai.achievement_id,
    COUNT(DISTINCT sai.external_content_id)::int AS item_count,
    COUNT(DISTINCT usp.external_content_id)::int AS progress_count
FROM streak_achievement_items sai
LEFT JOIN user_streak_progress usp
    ON usp.achievement_id = sai.achievement_id
    AND usp.user_id = @user_id::char(28)
    AND usp.external_content_id = sai.external_content_id
WHERE sai.achievement_id = ANY(@achievement_ids::char(28)[])
GROUP BY sai.achievement_id;

-- name: GetStreakItemCounts :many
-- Get streak item counts per achievement (for caching)
SELECT achievement_id, COUNT(*)::int AS item_count
FROM streak_achievement_items
WHERE achievement_id = ANY(@achievement_ids::char(28)[])
GROUP BY achievement_id;

-- name: GetUserStreakProgressCounts :many
-- Get user progress counts per streak achievement
SELECT achievement_id, COUNT(*)::int AS progress_count
FROM user_streak_progress
WHERE user_id = @user_id::char(28)
  AND achievement_id = ANY(@achievement_ids::char(28)[])
GROUP BY achievement_id;

-- name: GetBulkUserStreakProgress :many
SELECT user_id, achievement_id, external_content_id, completed_at
FROM user_streak_progress
WHERE (user_id, achievement_id) IN (
    SELECT unnest(@user_ids::char(28)[]), unnest(@achievement_ids::char(28)[])
);

-- name: GetUserStreakProgress :many
SELECT user_id, achievement_id, external_content_id, completed_at
FROM user_streak_progress
WHERE user_id = @user_id::char(28)
  AND achievement_id = ANY(@achievement_ids::char(28)[]);

-- name: MarkStreakItemCompleted :exec
INSERT INTO user_streak_progress (user_id, achievement_id, external_content_id, completed_at)
VALUES (@user_id::text, @achievement_id::text, @external_content_id::text, COALESCE(@completed_at::timestamptz, now()))
ON CONFLICT (user_id, achievement_id, external_content_id) DO NOTHING;

-- name: UnmarkStreakItemCompleted :exec
DELETE FROM user_streak_progress
WHERE user_id = @user_id::char(28)
  AND achievement_id = @achievement_id::char(28)
  AND external_content_id = @external_content_id::char(28);

-- name: GetPublishedStreakAchievementsByExternalContent :many
-- Get all streak achievements that contain a specific external content
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
                'id', sai2.id,
                'external_content_id', sai2.external_content_id,
                'sort_order', sai2.sort_order
            ) ORDER BY sai2.sort_order
        )
        FROM streak_achievement_items sai2
        WHERE sai2.achievement_id = a.id),
        '[]'::jsonb
    ) AS streak_items
FROM achievements a
INNER JOIN streak_achievements sa ON a.id = sa.achievement_id
INNER JOIN streak_achievement_items sai ON sa.achievement_id = sai.achievement_id
WHERE sai.external_content_id = @external_content_id::char(28);

-- name: MarkStreakItemCompletedForAllAchievements :exec
-- Mark content completed for a user across all streak achievements containing this content
INSERT INTO user_streak_progress (user_id, achievement_id, external_content_id, completed_at)
SELECT @user_id::text, sa.achievement_id, @external_content_id::char(28), now()
FROM streak_achievements sa
INNER JOIN streak_achievement_items sai ON sa.achievement_id = sai.achievement_id
WHERE sai.external_content_id = @external_content_id::char(28)
ON CONFLICT (user_id, achievement_id, external_content_id) DO NOTHING;

-- name: UnmarkStreakItemCompletedForAllAchievements :exec
-- Unmark content completed for a user across all streak achievements containing this content
DELETE FROM user_streak_progress
WHERE user_id = @user_id::char(28)
  AND external_content_id = @external_content_id::char(28);

-- name: GetStreakAchievementForAward :one
-- Get streak achievement data for awarding
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
    COALESCE(
        (SELECT jsonb_agg(
            jsonb_build_object(
                'id', sai.id,
                'external_content_id', sai.external_content_id,
                'sort_order', sai.sort_order
            ) ORDER BY sai.sort_order
        )
        FROM streak_achievement_items sai
        WHERE sai.achievement_id = a.id),
        '[]'::jsonb
    ) AS streak_items
FROM achievements a
INNER JOIN streak_achievements sa ON sa.achievement_id = a.id
WHERE a.id = @achievement_id::char(28) AND a.project_id = @project_id::char(28);

-- name: GetUsersWithUnclaimedStreakAchievement :many
-- Find users who completed all items for a streak achievement but weren't awarded
SELECT DISTINCT u.id AS user_id
FROM users u
INNER JOIN achievements a ON a.id = @achievement_id::char(28)
INNER JOIN streak_achievements sa ON sa.achievement_id = a.id
WHERE a.project_id = @project_id::char(28)
  AND NOT EXISTS (
    SELECT 1 FROM user_achievements ua
    WHERE ua.user_id = u.id AND ua.achievement_id = a.id
  )
  AND NOT EXISTS (
    SELECT 1 FROM streak_achievement_items sai
    WHERE sai.achievement_id = a.id
    AND NOT EXISTS (
      SELECT 1 FROM user_streak_progress usp
      WHERE usp.user_id = u.id
      AND usp.achievement_id = a.id
      AND usp.external_content_id = sai.external_content_id
    )
  )
  AND EXISTS (
    SELECT 1 FROM streak_achievement_items sai
    WHERE sai.achievement_id = a.id
  );

-- name: GetContentItemsWithDeadlines :many
-- Get content achievement items with external content details including complete_by deadline
SELECT cai.id, cai.achievement_id, cai.external_content_id, cai.sort_order, ec.complete_by
FROM content_achievement_items cai
INNER JOIN external_content ec ON cai.external_content_id = ec.id
WHERE cai.achievement_id = @achievement_id::char(28)
ORDER BY cai.sort_order;

-- name: GetStreakItemsWithDeadlines :many
-- Get streak achievement items with external content details including complete_by deadline
SELECT sai.id, sai.achievement_id, sai.external_content_id, sai.sort_order, ec.complete_by
FROM streak_achievement_items sai
INNER JOIN external_content ec ON sai.external_content_id = ec.id
WHERE sai.achievement_id = @achievement_id::char(28)
ORDER BY sai.sort_order;
