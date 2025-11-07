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
