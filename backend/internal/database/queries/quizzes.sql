-- name: GetQuizByID :one
SELECT id, project_id, challenge_id, name, description, image_url, timeout_seconds, question_timeout_seconds, randomize_questions, reveal_correct_answers, allow_retakes, completion_points, published_at, end_time, created_at, updated_at
FROM quizzes
WHERE id = @id::text;

-- name: GetQuizzesByIDs :many
SELECT id, project_id, challenge_id, name, description, image_url, timeout_seconds, question_timeout_seconds, randomize_questions, reveal_correct_answers, allow_retakes, completion_points, published_at, end_time, created_at, updated_at
FROM quizzes
WHERE id = ANY(@ids::text[]);

-- name: GetQuizzesByProjectIDs :many
SELECT id, project_id, challenge_id, name, description, image_url, timeout_seconds, question_timeout_seconds, randomize_questions, reveal_correct_answers, allow_retakes, completion_points, published_at, end_time, created_at, updated_at
FROM quizzes
WHERE project_id = ANY(@project_ids::text[])
    AND published_at IS NOT NULL
    AND published_at <= NOW()
ORDER BY project_id, published_at DESC;

-- name: GetQuizzesByChallengeIDs :many
SELECT id, project_id, challenge_id, name, description, image_url, timeout_seconds, question_timeout_seconds, randomize_questions, reveal_correct_answers, allow_retakes, completion_points, published_at, end_time, created_at, updated_at
FROM quizzes
WHERE challenge_id = ANY(@challenge_ids::text[])
    AND published_at IS NOT NULL
    AND published_at <= NOW()
ORDER BY challenge_id, published_at DESC;

-- name: GetQuizzesFilteredCursor :many
SELECT id, project_id, challenge_id, name, description, image_url, timeout_seconds, question_timeout_seconds, randomize_questions, reveal_correct_answers, allow_retakes, completion_points, published_at, end_time, created_at, updated_at
FROM quizzes
WHERE
    (@ids::text[] IS NULL OR id = ANY(@ids::text[]))
    AND (@projectid::text = '' OR project_id = @projectid::text)
    AND (@challengeid::text = '' OR challenge_id = @challengeid::text)
    AND (@publishedafter::timestamptz IS NULL OR published_at >= @publishedafter::timestamptz)
    AND (@publishedbefore::timestamptz IS NULL OR published_at <= @publishedbefore::timestamptz)
    AND (@aftercursor::text = '' OR id > @aftercursor::text)
    AND (@beforecursor::text = '' OR id < @beforecursor::text)
ORDER BY
    CASE WHEN @isbackward::bool = true THEN id END DESC,
    CASE WHEN @isbackward::bool = false OR @isbackward::bool IS NULL THEN id END ASC
LIMIT CASE WHEN @querylimit::int IS NULL THEN NULL ELSE @querylimit::int END;

-- name: CountQuizzesFiltered :one
SELECT COUNT(DISTINCT id)
FROM quizzes
WHERE
    (@ids::text[] IS NULL OR id = ANY(@ids::text[]))
    AND (@projectid::text = '' OR project_id = @projectid::text)
    AND (@challengeid::text = '' OR challenge_id = @challengeid::text)
    AND (@publishedafter::timestamptz IS NULL OR published_at >= @publishedafter::timestamptz)
    AND (@publishedbefore::timestamptz IS NULL OR published_at <= @publishedbefore::timestamptz);

-- name: CreateQuiz :one
INSERT INTO quizzes (
    id,
    project_id,
    challenge_id,
    name,
    description,
    image_url,
    timeout_seconds,
    question_timeout_seconds,
    randomize_questions,
    reveal_correct_answers,
    allow_retakes,
    completion_points,
    published_at,
    end_time
)
VALUES (
    @id::text,
    @projectid::text,
    @challengeid::text,
    @name::text,
    @description::text,
    sqlc.narg('imageurl')::text,
    sqlc.narg('timeoutseconds')::int,
    sqlc.narg('questiontimeoutseconds')::int,
    @randomizequestions::bool,
    @revealcorrectanswers::bool,
    @allowretakes::bool,
    @completionpoints::int,
    sqlc.narg('publishedat')::timestamptz,
    sqlc.narg('endtime')::timestamptz
)
RETURNING id, project_id, challenge_id, name, description, image_url, timeout_seconds, question_timeout_seconds, randomize_questions, reveal_correct_answers, allow_retakes, completion_points, published_at, end_time, created_at, updated_at;

-- name: UpdateQuiz :one
UPDATE quizzes
SET
    name = COALESCE(sqlc.narg('name')::text, name),
    description = COALESCE(sqlc.narg('description')::text, description),
    image_url = COALESCE(sqlc.narg('imageurl')::text, image_url),
    timeout_seconds = COALESCE(sqlc.narg('timeoutseconds')::int, timeout_seconds),
    question_timeout_seconds = COALESCE(sqlc.narg('questiontimeoutseconds')::int, question_timeout_seconds),
    randomize_questions = COALESCE(sqlc.narg('randomizequestions')::bool, randomize_questions),
    reveal_correct_answers = COALESCE(sqlc.narg('revealcorrectanswers')::bool, reveal_correct_answers),
    allow_retakes = COALESCE(sqlc.narg('allowretakes')::bool, allow_retakes),
    completion_points = COALESCE(sqlc.narg('completionpoints')::int, completion_points),
    end_time = COALESCE(sqlc.narg('endtime')::timestamptz, end_time),
    updated_at = now()
WHERE id = @id::text
RETURNING id, project_id, challenge_id, name, description, image_url, timeout_seconds, question_timeout_seconds, randomize_questions, reveal_correct_answers, allow_retakes, completion_points, published_at, end_time, created_at, updated_at;

-- name: DeleteQuiz :exec
DELETE FROM quizzes
WHERE id = @id::text;

-- name: PublishQuiz :one
UPDATE quizzes
SET
    published_at = @publishedat::timestamptz,
    updated_at = now()
WHERE id = @id::text
RETURNING id, project_id, challenge_id, name, description, image_url, timeout_seconds, question_timeout_seconds, randomize_questions, reveal_correct_answers, allow_retakes, completion_points, published_at, end_time, created_at, updated_at;

-- name: DeleteQuizTranslations :exec
DELETE FROM quiz_translations
WHERE quiz_id = @quizid::text;
