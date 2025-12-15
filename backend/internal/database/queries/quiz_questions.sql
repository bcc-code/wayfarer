-- name: GetQuizQuestionsByQuizID :many
SELECT id, quiz_id, question_type, question_text, question_order, allow_multiple_selection, min_value, max_value, step_value, timeout_seconds, points, created_at, updated_at
FROM quiz_questions
WHERE quiz_id = @quizid::text
ORDER BY question_order ASC;

-- name: GetQuizQuestionsByQuizIDs :many
SELECT id, quiz_id, question_type, question_text, question_order, allow_multiple_selection, min_value, max_value, step_value, timeout_seconds, points, created_at, updated_at
FROM quiz_questions
WHERE quiz_id = ANY(@quiz_ids::text[])
ORDER BY quiz_id, question_order ASC;

-- name: GetQuizQuestionByID :one
SELECT id, quiz_id, question_type, question_text, question_order, allow_multiple_selection, min_value, max_value, step_value, timeout_seconds, points, created_at, updated_at
FROM quiz_questions
WHERE id = @id::text;

-- name: GetQuizQuestionsByIDs :many
SELECT id, quiz_id, question_type, question_text, question_order, allow_multiple_selection, min_value, max_value, step_value, timeout_seconds, points, created_at, updated_at
FROM quiz_questions
WHERE id = ANY(@ids::text[]);

-- name: CreateQuizQuestion :one
INSERT INTO quiz_questions (
    id,
    quiz_id,
    question_type,
    question_text,
    question_order,
    allow_multiple_selection,
    min_value,
    max_value,
    step_value,
    timeout_seconds,
    points
)
VALUES (
    @id::text,
    @quizid::text,
    @questiontype::text,
    @questiontext::text,
    @questionorder::int,
    sqlc.narg('allowmultipleselection')::bool,
    sqlc.narg('minvalue')::decimal,
    sqlc.narg('maxvalue')::decimal,
    sqlc.narg('stepvalue')::decimal,
    sqlc.narg('timeoutseconds')::int,
    sqlc.narg('points')::int
)
RETURNING id, quiz_id, question_type, question_text, question_order, allow_multiple_selection, min_value, max_value, step_value, timeout_seconds, points, created_at, updated_at;

-- name: UpdateQuizQuestion :one
UPDATE quiz_questions
SET
    question_text = COALESCE(sqlc.narg('questiontext')::text, question_text),
    question_order = COALESCE(sqlc.narg('questionorder')::int, question_order),
    allow_multiple_selection = COALESCE(sqlc.narg('allowmultipleselection')::bool, allow_multiple_selection),
    min_value = COALESCE(sqlc.narg('minvalue')::decimal, min_value),
    max_value = COALESCE(sqlc.narg('maxvalue')::decimal, max_value),
    step_value = COALESCE(sqlc.narg('stepvalue')::decimal, step_value),
    timeout_seconds = COALESCE(sqlc.narg('timeoutseconds')::int, timeout_seconds),
    points = COALESCE(sqlc.narg('points')::int, points),
    updated_at = now()
WHERE id = @id::text
RETURNING id, quiz_id, question_type, question_text, question_order, allow_multiple_selection, min_value, max_value, step_value, timeout_seconds, points, created_at, updated_at;

-- name: DeleteQuizQuestion :exec
DELETE FROM quiz_questions
WHERE id = @id::text;

-- name: UpdateQuizQuestionOrder :exec
UPDATE quiz_questions
SET
    question_order = @questionorder::int,
    updated_at = now()
WHERE id = @id::text;

-- name: DeleteQuizQuestionTranslations :exec
DELETE FROM quiz_question_translations
WHERE question_id = @questionid::text;

-- name: GetPredefinedAnswersByQuestionID :many
SELECT id, question_id, answer_text, is_correct, answer_order, created_at
FROM quiz_predefined_answers
WHERE question_id = @questionid::text
ORDER BY answer_order ASC;

-- name: GetPredefinedAnswersByQuestionIDs :many
SELECT id, question_id, answer_text, is_correct, answer_order, created_at
FROM quiz_predefined_answers
WHERE question_id = ANY(@question_ids::text[])
ORDER BY question_id, answer_order ASC;

-- name: GetPredefinedAnswersByIDs :many
SELECT id, question_id, answer_text, is_correct, answer_order, created_at
FROM quiz_predefined_answers
WHERE id = ANY(@ids::text[]);

-- name: CreatePredefinedAnswer :one
INSERT INTO quiz_predefined_answers (
    id,
    question_id,
    answer_text,
    is_correct,
    answer_order
)
VALUES (
    @id::text,
    @questionid::text,
    @answertext::text,
    @iscorrect::bool,
    @answerorder::int
)
RETURNING id, question_id, answer_text, is_correct, answer_order, created_at;

-- name: DeletePredefinedAnswersByQuestion :exec
DELETE FROM quiz_predefined_answers
WHERE question_id = @questionid::text;

-- name: DeletePredefinedAnswerTranslations :exec
DELETE FROM quiz_answer_translations
WHERE answer_id = @answerid::text;
