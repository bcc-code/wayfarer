-- Push Subscriptions

-- name: CreatePushSubscription :one
INSERT INTO push_subscriptions (
    id,
    user_id,
    endpoint,
    p256dh_key,
    auth_key,
    user_agent
) VALUES (
    @id::text,
    @userid::text,
    @endpoint::text,
    @p256dhkey::text,
    @authkey::text,
    @useragent::text
) RETURNING *;

-- name: UpdatePushSubscription :one
UPDATE push_subscriptions
SET
    p256dh_key = @p256dhkey::text,
    auth_key = @authkey::text,
    user_agent = @useragent::text,
    updated_at = now()
WHERE endpoint = @endpoint::text
RETURNING *;

-- name: GetPushSubscriptionByEndpoint :one
SELECT * FROM push_subscriptions
WHERE endpoint = @endpoint::text;

-- name: DeletePushSubscriptionByEndpoint :exec
DELETE FROM push_subscriptions
WHERE endpoint = @endpoint::text;

-- name: DeletePushSubscriptionByID :exec
DELETE FROM push_subscriptions
WHERE id = @id::text;

-- name: GetPushSubscriptionsByUserID :many
SELECT * FROM push_subscriptions
WHERE user_id = @userid::text
ORDER BY created_at DESC;

-- name: GetPushSubscriptionsByUserIDs :many
SELECT * FROM push_subscriptions
WHERE user_id = ANY(@userids::text[])
ORDER BY user_id, created_at DESC;

-- name: GetAllPushSubscriptions :many
SELECT * FROM push_subscriptions
ORDER BY created_at DESC;

-- name: CountPushSubscriptions :one
SELECT COUNT(*) FROM push_subscriptions;

-- name: CountPushSubscriptionsByUser :one
SELECT COUNT(*) FROM push_subscriptions
WHERE user_id = @userid::text;

-- User Targeting Queries (for building notification recipient lists)

-- name: GetUserIDsInEvents :many
SELECT DISTINCT user_id
FROM user_events
WHERE event_id = ANY(@eventids::text[]);

-- name: GetUserIDsInProjects :many
SELECT DISTINCT user_id
FROM user_projects
WHERE project_id = ANY(@projectids::text[]);

-- Get subscriptions for users who have the notification type enabled (or no preference = enabled by default)
-- name: GetEnabledSubscriptionsForUsers :many
SELECT ps.*
FROM push_subscriptions ps
WHERE ps.user_id = ANY(@userids::text[])
  AND NOT EXISTS (
    SELECT 1 FROM push_notification_preferences pnp
    WHERE pnp.user_id = ps.user_id
      AND pnp.notification_type = @notificationtype::text
      AND pnp.enabled = false
  )
ORDER BY ps.user_id;

-- Get all subscribed user IDs (users who have at least one subscription)
-- name: GetAllSubscribedUserIDs :many
SELECT DISTINCT user_id
FROM push_subscriptions;

-- Notification Preferences

-- name: GetUserNotificationPreferences :many
SELECT * FROM push_notification_preferences
WHERE user_id = @userid::text
ORDER BY notification_type;

-- name: GetNotificationPreference :one
SELECT * FROM push_notification_preferences
WHERE user_id = @userid::text
  AND notification_type = @notificationtype::text;

-- name: UpsertNotificationPreference :one
INSERT INTO push_notification_preferences (
    user_id,
    notification_type,
    enabled
) VALUES (
    @userid::text,
    @notificationtype::text,
    @enabled::bool
) ON CONFLICT (user_id, notification_type)
DO UPDATE SET
    enabled = @enabled::bool,
    updated_at = now()
RETURNING *;

-- name: DeleteNotificationPreference :exec
DELETE FROM push_notification_preferences
WHERE user_id = @userid::text
  AND notification_type = @notificationtype::text;

-- name: IsNotificationTypeEnabled :one
SELECT COALESCE(
    (SELECT enabled FROM push_notification_preferences
     WHERE user_id = @userid::text AND notification_type = @notificationtype::text),
    true
)::bool AS enabled;

-- Notification Log

-- name: CreatePushNotificationLog :one
INSERT INTO push_notification_log (
    id,
    notification_type,
    title,
    body,
    url,
    data,
    target_criteria,
    sent_by,
    total_recipients,
    successful_deliveries,
    failed_deliveries
) VALUES (
    @id::text,
    @notificationtype::text,
    @title::text,
    @body::text,
    @url::text,
    @data::jsonb,
    @targetcriteria::jsonb,
    @sentby::text,
    @totalrecipients::int,
    @successfuldeliveries::int,
    @faileddeliveries::int
) RETURNING *;

-- name: UpdatePushNotificationLogStats :exec
UPDATE push_notification_log
SET
    total_recipients = @totalrecipients::int,
    successful_deliveries = @successfuldeliveries::int,
    failed_deliveries = @faileddeliveries::int
WHERE id = @id::text;

-- name: GetPushNotificationLog :one
SELECT * FROM push_notification_log
WHERE id = @id::text;

-- name: GetPushNotificationLogs :many
SELECT * FROM push_notification_log
ORDER BY sent_at DESC
LIMIT @limitcount::int
OFFSET @offsetcount::int;

-- name: CountPushNotificationLogs :one
SELECT COUNT(*) FROM push_notification_log;
