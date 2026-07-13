-- name: JoinEvent :exec
INSERT INTO user_events (user_id, event_id)
VALUES (@userid::text, @eventid::text)
ON CONFLICT (user_id, event_id) DO NOTHING;

-- name: LeaveEvent :exec
DELETE FROM user_events
WHERE user_id = @userid::char(28) AND event_id = @eventid::char(28);

-- name: GetUserEvents :many
SELECT event_id, joined_at
FROM user_events
WHERE user_id = @userid::char(28)
ORDER BY joined_at DESC;

-- name: IsUserInEvent :one
SELECT EXISTS(
    SELECT 1
    FROM user_events
    WHERE user_id = @userid::char(28) AND event_id = @eventid::char(28)
) AS is_member;

-- name: GetUserIDsByEventID :many
SELECT user_id FROM user_events WHERE event_id = @event_id;
