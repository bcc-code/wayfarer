-- Webhooks

-- name: CreateWebhook :one
INSERT INTO webhooks (
    id,
    project_id,
    name,
    url,
    event_type,
    include_user_data,
    include_event_data,
    active,
    secret
) VALUES (
    @id::text,
    @projectid::text,
    @name::text,
    @url::text,
    @eventtype::text,
    @includeuserdata::bool,
    @includeeventdata::bool,
    @active::bool,
    sqlc.narg('secret')::text
) RETURNING *;

-- name: GetWebhookByID :one
SELECT * FROM webhooks
WHERE id = @id::text;

-- name: GetWebhooksByProjectID :many
SELECT * FROM webhooks
WHERE project_id = @projectid::text
ORDER BY created_at DESC;

-- name: GetActiveWebhooksByProjectAndEvent :many
SELECT * FROM webhooks
WHERE project_id = @projectid::text
  AND event_type = @eventtype::text
  AND active = true
ORDER BY created_at DESC;

-- name: UpdateWebhook :one
UPDATE webhooks
SET
    name = COALESCE(sqlc.narg('name')::text, name),
    url = COALESCE(sqlc.narg('url')::text, url),
    include_user_data = COALESCE(sqlc.narg('includeuserdata')::bool, include_user_data),
    include_event_data = COALESCE(sqlc.narg('includeeventdata')::bool, include_event_data),
    active = COALESCE(sqlc.narg('active')::bool, active),
    secret = CASE WHEN sqlc.narg('updatesecret')::bool = true THEN sqlc.narg('secret')::text ELSE secret END,
    updated_at = now()
WHERE id = @id::text
RETURNING *;

-- name: DeleteWebhook :exec
DELETE FROM webhooks
WHERE id = @id::text;

-- Webhook Logs

-- name: CreateWebhookLog :one
INSERT INTO webhook_logs (
    id,
    webhook_id,
    event_type,
    request_payload,
    response_status_code,
    response_body,
    duration_ms,
    error_message
) VALUES (
    @id::text,
    @webhookid::text,
    @eventtype::text,
    @requestpayload::jsonb,
    sqlc.narg('responsestatuscode')::int,
    sqlc.narg('responsebody')::text,
    @durationms::int,
    sqlc.narg('errormessage')::text
) RETURNING *;

-- name: GetWebhookLogsByWebhookID :many
SELECT * FROM webhook_logs
WHERE webhook_id = @webhookid::text
ORDER BY created_at DESC
LIMIT @limitcount::int;

-- name: CountWebhookLogsByWebhookID :one
SELECT COUNT(*) FROM webhook_logs
WHERE webhook_id = @webhookid::text;
