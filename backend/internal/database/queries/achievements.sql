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
