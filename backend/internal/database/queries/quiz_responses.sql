-- name: GetQuizResponsesBySubmissionID :many
SELECT id, submission_id, question_id, selected_answer_ids, text_response, number_response, json_response, is_correct, answered_at, time_spent_seconds
FROM quiz_responses
WHERE submission_id = @submissionid::text;

-- name: GetQuizResponsesBySubmissionIDs :many
SELECT id, submission_id, question_id, selected_answer_ids, text_response, number_response, json_response, is_correct, answered_at, time_spent_seconds
FROM quiz_responses
WHERE submission_id = ANY(@submission_ids::text[])
ORDER BY submission_id;

-- name: GetQuizResponseBySubmissionAndQuestion :one
SELECT id, submission_id, question_id, selected_answer_ids, text_response, number_response, json_response, is_correct, answered_at, time_spent_seconds
FROM quiz_responses
WHERE submission_id = @submissionid::text
    AND question_id = @questionid::text;

-- name: CreateQuizResponse :one
INSERT INTO quiz_responses (
    id,
    submission_id,
    question_id,
    selected_answer_ids,
    text_response,
    number_response,
    json_response,
    is_correct,
    time_spent_seconds
)
VALUES (
    @id::text,
    @submissionid::text,
    @questionid::text,
    sqlc.narg('selectedanswerids')::jsonb,
    sqlc.narg('textresponse')::text,
    sqlc.narg('numberresponse')::decimal,
    sqlc.narg('jsonresponse')::jsonb,
    sqlc.narg('iscorrect')::bool,
    sqlc.narg('timespentseconds')::int
)
RETURNING id, submission_id, question_id, selected_answer_ids, text_response, number_response, json_response, is_correct, answered_at, time_spent_seconds;

-- name: UpdateQuizResponse :one
UPDATE quiz_responses
SET
    selected_answer_ids = COALESCE(sqlc.narg('selectedanswerids')::jsonb, selected_answer_ids),
    text_response = COALESCE(sqlc.narg('textresponse')::text, text_response),
    number_response = COALESCE(sqlc.narg('numberresponse')::decimal, number_response),
    json_response = COALESCE(sqlc.narg('jsonresponse')::jsonb, json_response),
    is_correct = COALESCE(sqlc.narg('iscorrect')::bool, is_correct),
    time_spent_seconds = COALESCE(sqlc.narg('timespentseconds')::int, time_spent_seconds),
    answered_at = now()
WHERE id = @id::text
RETURNING id, submission_id, question_id, selected_answer_ids, text_response, number_response, json_response, is_correct, answered_at, time_spent_seconds;

-- name: CalculateSubmissionScore :one
SELECT
    COUNT(*) FILTER (WHERE is_correct = true) AS score,
    COUNT(*) FILTER (WHERE is_correct IS NOT NULL) AS max_score
FROM quiz_responses
WHERE submission_id = @submissionid::text;
