-- User Feedback

-- name: CreateUserFeedback :one
INSERT INTO user_feedback (
    id,
    user_id,
    message,
    can_contact_me,
    user_agent,
    platform,
    screen_width,
    screen_height,
    app_version,
    locale,
    project_id,
    timezone,
    context_url,
    tags
) VALUES (
    @id::text,
    @userid::text,
    @message::text,
    @cancontactme::bool,
    @useragent::text,
    @platform::text,
    @screenwidth::int,
    @screenheight::int,
    @appversion::text,
    @locale::text,
    NULLIF(@projectid::text, ''),
    @timezone::text,
    @contexturl::text,
    @tags::text[]
) RETURNING *;

-- name: GetRecentFeedbackCount :one
SELECT COUNT(*) FROM user_feedback
WHERE user_id = @userid::char(28)
  AND created_at > @since::timestamptz;

-- name: GetUserFeedback :many
SELECT * FROM user_feedback
WHERE user_id = @userid::char(28)
ORDER BY created_at DESC;

-- name: GetAllFeedback :many
SELECT * FROM user_feedback
ORDER BY created_at DESC
LIMIT @limitcount::int
OFFSET @offsetcount::int;

-- name: CountAllFeedback :one
SELECT COUNT(*) FROM user_feedback
WHERE
    (@filteruserid::char(28) = '' OR user_id = @filteruserid::char(28))
    AND (cardinality(@filtertags::text[]) = 0 OR tags && @filtertags::text[])
    AND (@filterhandled::text = '' OR (@filterhandled::text = 'true' AND handled_at IS NOT NULL) OR (@filterhandled::text = 'false' AND handled_at IS NULL))
    AND (@filterplatform::text = '' OR platform = @filterplatform::text);

-- name: GetFeedbackCursor :many
SELECT * FROM user_feedback
WHERE
    (@aftercursor::char(28) = '' OR id < @aftercursor::char(28))
    AND (@beforecursor::char(28) = '' OR id > @beforecursor::char(28))
    AND (@filteruserid::char(28) = '' OR user_id = @filteruserid::char(28))
    AND (cardinality(@filtertags::text[]) = 0 OR tags && @filtertags::text[])
    AND (@filterhandled::text = '' OR (@filterhandled::text = 'true' AND handled_at IS NOT NULL) OR (@filterhandled::text = 'false' AND handled_at IS NULL))
    AND (@filterplatform::text = '' OR platform = @filterplatform::text)
ORDER BY
    CASE WHEN @isbackward::bool = true THEN id END ASC,
    CASE WHEN @isbackward::bool = false OR @isbackward::bool IS NULL THEN id END DESC
LIMIT @querylimit::int;

-- name: GetFeedbackByID :one
SELECT * FROM user_feedback WHERE id = @id::char(28);

-- name: SetFeedbackHandledAt :one
UPDATE user_feedback
SET handled_at = @handled_at::timestamptz
WHERE id = @id::char(28)
RETURNING *;

-- name: DeleteFeedback :exec
DELETE FROM user_feedback WHERE id = @id::char(28);

-- name: UpdateFeedbackTags :one
UPDATE user_feedback
SET tags = @tags::text[]
WHERE id = @id::char(28)
RETURNING *;

-- name: GetDistinctFeedbackTags :many
SELECT DISTINCT unnest(tags)::text AS tag FROM user_feedback ORDER BY tag;

-- name: GetDistinctFeedbackPlatforms :many
SELECT DISTINCT platform FROM user_feedback WHERE platform IS NOT NULL ORDER BY platform;
