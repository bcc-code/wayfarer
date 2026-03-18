-- name: JoinEvent :exec
INSERT INTO user_events (user_id, event_id)
VALUES (@userid::text, @eventid::text)
ON CONFLICT (user_id, event_id) DO NOTHING;

-- name: LeaveEvent :exec
DELETE FROM user_events
WHERE user_id = @userid::text AND event_id = @eventid::text;

-- name: GetUserEvents :many
SELECT event_id, joined_at
FROM user_events
WHERE user_id = @userid::text
ORDER BY joined_at DESC;

-- name: IsUserInEvent :one
SELECT EXISTS(
    SELECT 1
    FROM user_events
    WHERE user_id = @userid::text AND event_id = @eventid::text
) AS is_member;

-- name: GetUserIDsByEventID :many
SELECT user_id FROM user_events WHERE event_id = @event_id;
