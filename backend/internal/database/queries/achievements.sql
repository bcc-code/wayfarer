-- name: GetAchievementsByIDs :many
SELECT
    a.id,
    a.achievement_type,
    a.project_id,
    a.event_id,
    a.challenge_id,
    a.name,
    a.description,
    a.image_url,
    a.points,
    a.hidden,
    a.created_at,
    a.updated_at,
    -- Reading achievement data
    ra.achievement_id AS reading_achievement_id,
    COALESCE(
        (SELECT jsonb_agg(
            jsonb_build_object(
                'id', raa.id,
                'article_id', raa.article_id,
                'title', raa.title,
                'author', raa.author,
                'url', raa.url
            )
        )
        FROM reading_achievement_articles raa
        WHERE raa.achievement_id = a.id),
        '[]'::jsonb
    ) AS reading_articles,
    -- Listening achievement data
    la.achievement_id AS listening_achievement_id,
    COALESCE(
        (SELECT jsonb_agg(
            jsonb_build_object(
                'id', lat.id,
                'track_id', lat.track_id,
                'name', lat.name,
                'description', lat.description,
                'image_url', lat.image_url
            )
        )
        FROM listening_achievement_tracks lat
        WHERE lat.achievement_id = a.id),
        '[]'::jsonb
    ) AS listening_tracks,
    -- Streak achievement data
    sa.streak_id,
    sa.needed_streak
FROM achievements a
LEFT JOIN reading_achievements ra ON a.id = ra.achievement_id
LEFT JOIN listening_achievements la ON a.id = la.achievement_id
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
    a.description,
    a.image_url,
    a.points,
    a.hidden,
    a.created_at,
    a.updated_at,
    -- Type-specific fields needed for model construction
    ra.achievement_id AS reading_achievement_id,
    la.achievement_id AS listening_achievement_id,
    sa.streak_id,
    sa.needed_streak
FROM achievements a
LEFT JOIN reading_achievements ra ON a.id = ra.achievement_id
LEFT JOIN listening_achievements la ON a.id = la.achievement_id
LEFT JOIN streak_achievements sa ON a.id = sa.achievement_id
WHERE a.project_id = ANY(@project_ids::text[])
    AND a.hidden = false
ORDER BY a.project_id, a.created_at DESC;

-- name: GetAchievementsFilteredCursor :many
SELECT
    a.id,
    a.achievement_type,
    a.project_id,
    a.event_id,
    a.challenge_id,
    a.name,
    a.description,
    a.image_url,
    a.points,
    a.hidden,
    a.created_at,
    a.updated_at,
    -- Reading achievement data
    ra.achievement_id AS reading_achievement_id,
    COALESCE(
        (SELECT jsonb_agg(
            jsonb_build_object(
                'id', raa.id,
                'article_id', raa.article_id,
                'title', raa.title,
                'author', raa.author,
                'url', raa.url
            )
        )
        FROM reading_achievement_articles raa
        WHERE raa.achievement_id = a.id),
        '[]'::jsonb
    ) AS reading_articles,
    -- Listening achievement data
    la.achievement_id AS listening_achievement_id,
    COALESCE(
        (SELECT jsonb_agg(
            jsonb_build_object(
                'id', lat.id,
                'track_id', lat.track_id,
                'name', lat.name,
                'description', lat.description,
                'image_url', lat.image_url
            )
        )
        FROM listening_achievement_tracks lat
        WHERE lat.achievement_id = a.id),
        '[]'::jsonb
    ) AS listening_tracks,
    -- Streak achievement data
    sa.streak_id,
    sa.needed_streak
FROM achievements a
LEFT JOIN reading_achievements ra ON a.id = ra.achievement_id
LEFT JOIN listening_achievements la ON a.id = la.achievement_id
LEFT JOIN streak_achievements sa ON a.id = sa.achievement_id
WHERE
    (@ids::text[] IS NULL OR a.id = ANY(@ids::text[]))
    AND (@projectid::text = '' OR a.project_id = @projectid::text)
    AND (@eventid::text = '' OR a.event_id = @eventid::text)
    AND (@aftercursor::text = '' OR a.id > @aftercursor::text)
    AND (@beforecursor::text = '' OR a.id < @beforecursor::text)
ORDER BY
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

-- name: GetArticlesByAchievementIDs :many
SELECT id, achievement_id, article_id, title, author, url
FROM reading_achievement_articles
WHERE achievement_id = ANY(@achievement_ids::text[])
ORDER BY achievement_id;

-- name: GetTracksByAchievementIDs :many
SELECT id, achievement_id, track_id, name, description, image_url
FROM listening_achievement_tracks
WHERE achievement_id = ANY(@achievement_ids::text[])
ORDER BY achievement_id;

-- ==================== Create Operations ====================

-- name: CreateAchievement :one
INSERT INTO achievements (
    id,
    achievement_type,
    project_id,
    event_id,
    challenge_id,
    name,
    description,
    image_url,
    points,
    hidden
) VALUES (
    @id::text,
    @achievement_type::text,
    @project_id::text,
    @event_id::text,
    @challenge_id::text,
    @name::text,
    @description::text,
    @image_url::text,
    @points::int,
    @hidden::bool
) RETURNING *;

-- name: CreateReadingAchievementJunction :exec
INSERT INTO reading_achievements (achievement_id)
VALUES (@achievement_id::text);

-- name: CreateReadingAchievementArticle :one
INSERT INTO reading_achievement_articles (
    id,
    achievement_id,
    article_id,
    title,
    author,
    url
) VALUES (
    @id::text,
    @achievement_id::text,
    @article_id::text,
    @title::text,
    @author::text,
    @url::text
) RETURNING *;

-- name: CreateListeningAchievementJunction :exec
INSERT INTO listening_achievements (achievement_id)
VALUES (@achievement_id::text);

-- name: CreateListeningAchievementTrack :one
INSERT INTO listening_achievement_tracks (
    id,
    achievement_id,
    track_id,
    name,
    description,
    image_url
) VALUES (
    @id::text,
    @achievement_id::text,
    @track_id::text,
    @name::text,
    @description::text,
    @image_url::text
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
    description = CASE WHEN sqlc.narg('description')::text IS NOT NULL THEN sqlc.narg('description')::text ELSE description END,
    image_url = CASE WHEN sqlc.narg('image_url')::text IS NOT NULL THEN sqlc.narg('image_url')::text ELSE image_url END,
    event_id = CASE WHEN sqlc.narg('event_id')::text IS NOT NULL THEN sqlc.narg('event_id')::text ELSE event_id END,
    challenge_id = CASE WHEN sqlc.narg('challenge_id')::text IS NOT NULL THEN sqlc.narg('challenge_id')::text ELSE challenge_id END,
    points = CASE WHEN sqlc.narg('points')::int IS NOT NULL THEN sqlc.narg('points')::int ELSE points END,
    hidden = CASE WHEN sqlc.narg('hidden')::bool IS NOT NULL THEN sqlc.narg('hidden')::bool ELSE hidden END,
    updated_at = now()
WHERE id = @id::text
RETURNING *;

-- name: DeleteReadingAchievementArticles :exec
DELETE FROM reading_achievement_articles
WHERE achievement_id = @achievement_id::text;

-- name: DeleteListeningAchievementTracks :exec
DELETE FROM listening_achievement_tracks
WHERE achievement_id = @achievement_id::text;

-- name: UpdateStreakAchievementData :exec
UPDATE streak_achievements
SET
    streak_id = CASE WHEN sqlc.narg('streak_id')::text IS NOT NULL THEN sqlc.narg('streak_id')::text ELSE streak_id END,
    needed_streak = CASE WHEN sqlc.narg('needed_streak')::int IS NOT NULL THEN sqlc.narg('needed_streak')::int ELSE needed_streak END
WHERE achievement_id = @achievement_id::text;

-- ==================== Delete Operations ====================

-- name: DeleteAchievement :exec
DELETE FROM achievements
WHERE id = @id::text;

-- ==================== Award/Revoke Operations ====================

-- name: AwardUserAchievement :exec
INSERT INTO user_achievements (user_id, achievement_id, achieved_at)
VALUES (@user_id::text, @achievement_id::text, COALESCE(@achieved_at::timestamptz, now()));

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
