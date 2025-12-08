-- name: GetUserStreakActivity :many
SELECT user_id, streak_id, activity_date, created_at
FROM user_streak_activity
WHERE user_id = @userid::text
  AND streak_id = @streakid::text
ORDER BY activity_date DESC;

-- name: GetUserStreakActivitiesForMultipleStreaks :many
SELECT user_id, streak_id, activity_date, created_at
FROM user_streak_activity
WHERE user_id = @userid::text
  AND streak_id = ANY(@streak_ids::text[])
ORDER BY streak_id, activity_date DESC;

-- name: GetBulkUserStreakActivities :many
SELECT user_id, streak_id, activity_date, created_at
FROM user_streak_activity
WHERE (user_id, streak_id) IN (
    SELECT unnest(@user_ids::text[]), unnest(@streak_ids::text[])
)
ORDER BY user_id, streak_id, activity_date DESC;

-- name: GetLatestActivityDate :one
SELECT activity_date
FROM user_streak_activity
WHERE user_id = @userid::text
  AND streak_id = @streakid::text
ORDER BY activity_date DESC
LIMIT 1;

-- name: CountUserStreakActivities :one
SELECT COUNT(*)
FROM user_streak_activity
WHERE user_id = @userid::text
  AND streak_id = @streakid::text;

-- name: RecordUserStreakActivity :exec
INSERT INTO user_streak_activity (user_id, streak_id, activity_date, created_at)
VALUES (@userid::text, @streakid::text, @activitydate::date, NOW())
ON CONFLICT (user_id, streak_id, activity_date) DO NOTHING;
