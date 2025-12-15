-- name: GetQuizSubmissionByID :one
SELECT id, quiz_id, user_id, started_at, completed_at, expires_at, question_order, score, max_score, points_awarded, created_at
FROM quiz_submissions
WHERE id = @id::text;

-- name: GetQuizSubmissionByIDForUpdate :one
-- Acquires a row-level lock to prevent concurrent finalization
SELECT id, quiz_id, user_id, started_at, completed_at, expires_at, question_order, score, max_score, points_awarded, created_at
FROM quiz_submissions
WHERE id = @id::text
FOR UPDATE;

-- name: GetQuizSubmissionsByIDs :many
SELECT id, quiz_id, user_id, started_at, completed_at, expires_at, question_order, score, max_score, points_awarded, created_at
FROM quiz_submissions
WHERE id = ANY(@ids::text[]);

-- name: GetQuizSubmissionsByQuizID :many
SELECT id, quiz_id, user_id, started_at, completed_at, expires_at, question_order, score, max_score, points_awarded, created_at
FROM quiz_submissions
WHERE quiz_id = @quizid::text
ORDER BY started_at DESC;

-- name: GetQuizSubmissionsByUserID :many
SELECT id, quiz_id, user_id, started_at, completed_at, expires_at, question_order, score, max_score, points_awarded, created_at
FROM quiz_submissions
WHERE user_id = @userid::text
ORDER BY started_at DESC;

-- name: GetQuizSubmissionsByUserAndQuiz :many
SELECT id, quiz_id, user_id, started_at, completed_at, expires_at, question_order, score, max_score, points_awarded, created_at
FROM quiz_submissions
WHERE user_id = @userid::text
    AND quiz_id = @quizid::text
ORDER BY started_at DESC;

-- name: GetQuizSubmissionsByUserIDs :many
SELECT id, quiz_id, user_id, started_at, completed_at, expires_at, question_order, score, max_score, points_awarded, created_at
FROM quiz_submissions
WHERE user_id = ANY(@user_ids::text[])
ORDER BY user_id, started_at DESC;

-- name: GetActiveSubmissionByUserAndQuiz :one
SELECT id, quiz_id, user_id, started_at, completed_at, expires_at, question_order, score, max_score, points_awarded, created_at
FROM quiz_submissions
WHERE user_id = @userid::text
    AND quiz_id = @quizid::text
    AND completed_at IS NULL
    AND (expires_at IS NULL OR expires_at > NOW())
ORDER BY started_at DESC
LIMIT 1;

-- name: GetCompletedSubmissionsByUserAndQuiz :many
SELECT id, quiz_id, user_id, started_at, completed_at, expires_at, question_order, score, max_score, points_awarded, created_at
FROM quiz_submissions
WHERE user_id = @userid::text
    AND quiz_id = @quizid::text
    AND completed_at IS NOT NULL
ORDER BY completed_at DESC;

-- name: GetQuizSubmissionsFilteredCursor :many
SELECT id, quiz_id, user_id, started_at, completed_at, expires_at, question_order, score, max_score, points_awarded, created_at
FROM quiz_submissions
WHERE
    (@quizid::text = '' OR quiz_id = @quizid::text)
    AND (@userid::text = '' OR user_id = @userid::text)
    AND (@aftercursor::text = '' OR id > @aftercursor::text)
    AND (@beforecursor::text = '' OR id < @beforecursor::text)
ORDER BY
    CASE WHEN @isbackward::bool = true THEN id END DESC,
    CASE WHEN @isbackward::bool = false OR @isbackward::bool IS NULL THEN id END ASC
LIMIT CASE WHEN @querylimit::int IS NULL THEN NULL ELSE @querylimit::int END;

-- name: CountQuizSubmissionsFiltered :one
SELECT COUNT(DISTINCT id)
FROM quiz_submissions
WHERE
    (@quizid::text = '' OR quiz_id = @quizid::text)
    AND (@userid::text = '' OR user_id = @userid::text);

-- name: CreateQuizSubmission :one
INSERT INTO quiz_submissions (
    id,
    quiz_id,
    user_id,
    expires_at,
    question_order,
    max_score
)
VALUES (
    @id::text,
    @quizid::text,
    @userid::text,
    sqlc.narg('expiresat')::timestamptz,
    @questionorder::jsonb,
    @maxscore::int
)
RETURNING id, quiz_id, user_id, started_at, completed_at, expires_at, question_order, score, max_score, points_awarded, created_at;

-- name: UpdateQuizSubmission :one
UPDATE quiz_submissions
SET
    completed_at = COALESCE(sqlc.narg('completedat')::timestamptz, completed_at),
    score = COALESCE(sqlc.narg('score')::int, score),
    points_awarded = COALESCE(sqlc.narg('pointsawarded')::int, points_awarded)
WHERE id = @id::text
RETURNING id, quiz_id, user_id, started_at, completed_at, expires_at, question_order, score, max_score, points_awarded, created_at;
