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
    app_version
) VALUES (
    @id::text,
    @userid::text,
    @message::text,
    @cancontactme::bool,
    @useragent::text,
    @platform::text,
    @screenwidth::int,
    @screenheight::int,
    @appversion::text
) RETURNING *;

-- name: GetRecentFeedbackCount :one
SELECT COUNT(*) FROM user_feedback
WHERE user_id = @userid::text
  AND created_at > @since::timestamptz;

-- name: GetUserFeedback :many
SELECT * FROM user_feedback
WHERE user_id = @userid::text
ORDER BY created_at DESC;

-- name: GetAllFeedback :many
SELECT * FROM user_feedback
ORDER BY created_at DESC
LIMIT @limitcount::int
OFFSET @offsetcount::int;

-- name: CountAllFeedback :one
SELECT COUNT(*) FROM user_feedback
WHERE (@filteruserid::text = '' OR user_id = @filteruserid::text);

-- name: GetFeedbackCursor :many
SELECT * FROM user_feedback
WHERE
    (@aftercursor::text = '' OR id < @aftercursor::text)
    AND (@beforecursor::text = '' OR id > @beforecursor::text)
    AND (@filteruserid::text = '' OR user_id = @filteruserid::text)
ORDER BY
    CASE WHEN @isbackward::bool = true THEN id END ASC,
    CASE WHEN @isbackward::bool = false OR @isbackward::bool IS NULL THEN id END DESC
LIMIT @querylimit::int;
