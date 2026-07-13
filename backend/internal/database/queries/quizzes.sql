-- name: GetQuizByID :one
SELECT id, project_id, challenge_id, name, description, image_url, timeout_seconds, randomize_questions, reveal_correct_answers, allow_retakes, completion_points, end_time, created_at, updated_at
FROM quizzes
WHERE id = @id::char(28);

-- name: GetQuizzesByIDs :many
SELECT id, project_id, challenge_id, name, description, image_url, timeout_seconds, randomize_questions, reveal_correct_answers, allow_retakes, completion_points, end_time, created_at, updated_at
FROM quizzes
WHERE id = ANY(@ids::char(28)[]);

-- name: GetQuizzesByProjectIDs :many
SELECT id, project_id, challenge_id, name, description, image_url, timeout_seconds, randomize_questions, reveal_correct_answers, allow_retakes, completion_points, end_time, created_at, updated_at
FROM quizzes
WHERE project_id = ANY(@project_ids::char(28)[])
ORDER BY project_id, created_at DESC;

-- name: GetQuizzesByChallengeIDs :many
-- Batch version of GetQuizByChallengeID for dataloader
SELECT id, project_id, challenge_id, name, description, image_url, timeout_seconds, randomize_questions, reveal_correct_answers, allow_retakes, completion_points, end_time, created_at, updated_at
FROM quizzes
WHERE challenge_id = ANY(@challenge_ids::char(28)[])
ORDER BY challenge_id;

-- name: GetQuizzesFilteredCursor :many
SELECT id, project_id, challenge_id, name, description, image_url, timeout_seconds, randomize_questions, reveal_correct_answers, allow_retakes, completion_points, end_time, created_at, updated_at
FROM quizzes
WHERE
    (@ids::char(28)[] IS NULL OR id = ANY(@ids::char(28)[]))
    AND (@projectid::char(28) = '' OR project_id = @projectid::char(28))
    AND (@challengeid::char(28) = '' OR challenge_id = @challengeid::char(28))
    AND (@aftercursor::char(28) = '' OR id > @aftercursor::char(28))
    AND (@beforecursor::char(28) = '' OR id < @beforecursor::char(28))
ORDER BY
    CASE WHEN @isbackward::bool = true THEN id END DESC,
    CASE WHEN @isbackward::bool = false OR @isbackward::bool IS NULL THEN id END ASC
LIMIT CASE WHEN @querylimit::int IS NULL THEN NULL ELSE @querylimit::int END;

-- name: CountQuizzesFiltered :one
SELECT COUNT(DISTINCT id)
FROM quizzes
WHERE
    (@ids::char(28)[] IS NULL OR id = ANY(@ids::char(28)[]))
    AND (@projectid::char(28) = '' OR project_id = @projectid::char(28))
    AND (@challengeid::char(28) = '' OR challenge_id = @challengeid::char(28));

-- name: CreateQuiz :one
INSERT INTO quizzes (
    id,
    project_id,
    challenge_id,
    name,
    description,
    image_url,
    timeout_seconds,
    randomize_questions,
    reveal_correct_answers,
    allow_retakes,
    completion_points,
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
    @randomizequestions::bool,
    @revealcorrectanswers::bool,
    @allowretakes::bool,
    @completionpoints::int,
    sqlc.narg('endtime')::timestamptz
)
RETURNING id, project_id, challenge_id, name, description, image_url, timeout_seconds, randomize_questions, reveal_correct_answers, allow_retakes, completion_points, end_time, created_at, updated_at;

-- name: UpdateQuiz :one
UPDATE quizzes
SET
    name = COALESCE(sqlc.narg('name')::text, name),
    description = COALESCE(sqlc.narg('description')::text, description),
    image_url = COALESCE(sqlc.narg('imageurl')::text, image_url),
    timeout_seconds = COALESCE(sqlc.narg('timeoutseconds')::int, timeout_seconds),
    randomize_questions = COALESCE(sqlc.narg('randomizequestions')::bool, randomize_questions),
    reveal_correct_answers = COALESCE(sqlc.narg('revealcorrectanswers')::bool, reveal_correct_answers),
    allow_retakes = COALESCE(sqlc.narg('allowretakes')::bool, allow_retakes),
    completion_points = COALESCE(sqlc.narg('completionpoints')::int, completion_points),
    end_time = COALESCE(sqlc.narg('endtime')::timestamptz, end_time),
    updated_at = now()
WHERE id = @id::char(28)
RETURNING id, project_id, challenge_id, name, description, image_url, timeout_seconds, randomize_questions, reveal_correct_answers, allow_retakes, completion_points, end_time, created_at, updated_at;

-- name: DeleteQuiz :exec
DELETE FROM quizzes
WHERE id = @id::char(28);

-- name: DeleteQuizTranslations :exec
DELETE FROM quiz_translations
WHERE quiz_id = @quizid::char(28);
