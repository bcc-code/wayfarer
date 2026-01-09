-- ==================== Quiz Session CRUD ====================

-- name: CreateQuizSession :one
INSERT INTO quiz_sessions (
    id,
    quiz_id,
    name,
    state,
    open_at,
    lock_at,
    finish_at,
    created_by
)
VALUES (
    @id::text,
    @quizid::text,
    sqlc.narg('name')::text,
    'DRAFT',
    sqlc.narg('openat')::timestamptz,
    sqlc.narg('lockat')::timestamptz,
    sqlc.narg('finishat')::timestamptz,
    @createdby::text
)
RETURNING *;

-- name: GetQuizSession :one
SELECT *
FROM quiz_sessions
WHERE id = @id::text;

-- name: GetQuizSessionForUpdate :one
SELECT *
FROM quiz_sessions
WHERE id = @id::text
FOR UPDATE;

-- name: GetQuizSessionsByQuiz :many
SELECT *
FROM quiz_sessions
WHERE quiz_id = @quizid::text
    AND (@state::text = '' OR state = @state::text)
ORDER BY created_at DESC;

-- name: GetQuizSessionsByIDs :many
SELECT *
FROM quiz_sessions
WHERE id = ANY(@ids::text[]);

-- name: UpdateQuizSession :one
UPDATE quiz_sessions
SET
    name = COALESCE(sqlc.narg('name')::text, name),
    open_at = sqlc.narg('openat')::timestamptz,
    lock_at = sqlc.narg('lockat')::timestamptz,
    finish_at = sqlc.narg('finishat')::timestamptz,
    updated_at = now()
WHERE id = @id::text
RETURNING *;

-- name: UpdateQuizSessionState :one
UPDATE quiz_sessions
SET
    state = @state::text,
    updated_at = now()
WHERE id = @id::text
RETURNING *;

-- name: DeleteQuizSession :exec
DELETE FROM quiz_sessions
WHERE id = @id::text;

-- ==================== Scheduled Transitions ====================

-- name: GetSessionsPendingOpen :many
SELECT *
FROM quiz_sessions
WHERE state = 'DRAFT'
    AND open_at IS NOT NULL
    AND open_at <= NOW();

-- name: GetSessionsPendingLock :many
SELECT *
FROM quiz_sessions
WHERE state = 'OPEN'
    AND lock_at IS NOT NULL
    AND lock_at <= NOW();

-- name: GetSessionsPendingFinish :many
SELECT *
FROM quiz_sessions
WHERE state = 'LOCKED'
    AND finish_at IS NOT NULL
    AND finish_at <= NOW();

-- ==================== Quiz Session Access ====================

-- name: CreateQuizSessionAccess :one
INSERT INTO quiz_session_access (
    id,
    session_id,
    user_id,
    granted_by,
    source_type,
    source_id
)
VALUES (
    @id::text,
    @sessionid::text,
    @userid::text,
    @grantedby::text,
    @sourcetype::text,
    sqlc.narg('sourceid')::text
)
ON CONFLICT (session_id, user_id) DO NOTHING
RETURNING *;

-- name: HasQuizSessionAccess :one
SELECT EXISTS(
    SELECT 1
    FROM quiz_session_access
    WHERE session_id = @sessionid::text
        AND user_id = @userid::text
) AS has_access;

-- name: GetQuizSessionAccessCount :one
SELECT COUNT(*)::int
FROM quiz_session_access
WHERE session_id = @sessionid::text;

-- name: GetQuizSessionAccessUserIDs :many
SELECT user_id
FROM quiz_session_access
WHERE session_id = @sessionid::text;

-- name: DeleteQuizSessionAccess :exec
DELETE FROM quiz_session_access
WHERE session_id = @sessionid::text
    AND user_id = ANY(@userids::text[]);

-- name: DeleteAllQuizSessionAccess :exec
DELETE FROM quiz_session_access
WHERE session_id = @sessionid::text;

-- ==================== User Resolution Helpers ====================

-- name: GetUserIDsByChurchIDs :many
SELECT DISTINCT id AS user_id
FROM users
WHERE church_id = ANY(@churchids::text[]);

-- name: GetUserIDsByChurchIDsInProject :many
SELECT DISTINCT up.user_id
FROM user_projects up
JOIN users u ON u.id = up.user_id
WHERE u.church_id = ANY(@churchids::text[])
    AND up.project_id = @projectid::text;

-- ==================== Session Submissions ====================

-- name: GetActiveSubmissionsBySessionID :many
SELECT id, quiz_id, user_id, session_id, started_at, completed_at, expires_at,
       question_order, score, max_score, points_awarded, auto_submitted, created_at
FROM quiz_submissions
WHERE session_id = @sessionid::text
    AND completed_at IS NULL;

-- name: GetSubmissionByUserAndSession :one
SELECT id, quiz_id, user_id, session_id, started_at, completed_at, expires_at,
       question_order, score, max_score, points_awarded, auto_submitted, created_at
FROM quiz_submissions
WHERE session_id = @sessionid::text
    AND user_id = @userid::text
ORDER BY started_at DESC
LIMIT 1;

-- name: GetSubmissionsBySessionID :many
SELECT id, quiz_id, user_id, session_id, started_at, completed_at, expires_at,
       question_order, score, max_score, points_awarded, auto_submitted, created_at
FROM quiz_submissions
WHERE session_id = @sessionid::text
ORDER BY started_at DESC;

-- name: AutoSubmitSessionSubmissions :exec
UPDATE quiz_submissions
SET
    completed_at = NOW(),
    auto_submitted = true
WHERE session_id = @sessionid::text
    AND completed_at IS NULL;

-- name: DeleteSubmissionByUserAndSession :exec
DELETE FROM quiz_submissions
WHERE session_id = @sessionid::text
    AND user_id = @userid::text;

-- ==================== User Session Access ====================

-- name: GetUserAccessibleSessions :many
SELECT qs.*
FROM quiz_sessions qs
JOIN quiz_session_access qsa ON qsa.session_id = qs.id
WHERE qsa.user_id = @userid::text
    AND qs.quiz_id = @quizid::text
ORDER BY qs.created_at DESC;

-- name: GetUserAccessibleOpenSessions :many
SELECT qs.*
FROM quiz_sessions qs
JOIN quiz_session_access qsa ON qsa.session_id = qs.id
WHERE qsa.user_id = @userid::text
    AND qs.state = 'OPEN'
ORDER BY qs.created_at DESC;

-- name: GetSessionSubmissionsWithUserData :many
SELECT
    qs.id AS submission_id,
    qs.user_id,
    qs.score,
    qs.max_score,
    CASE WHEN qs.completed_at IS NOT NULL THEN true ELSE false END AS completed,
    qs.auto_submitted,
    u.members_id,
    u.church_id
FROM quiz_submissions qs
JOIN users u ON u.id = qs.user_id
WHERE qs.session_id = @sessionid::text
ORDER BY qs.started_at;
