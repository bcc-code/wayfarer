-- name: GetQuizResponseByID :one
SELECT id, submission_id, question_id, selected_answer_ids, text_response, number_response, json_response, is_correct, points_earned, answered_at, time_spent_seconds
FROM quiz_responses
WHERE id = @id::char(28);

-- name: GetQuizResponsesBySubmissionID :many
SELECT id, submission_id, question_id, selected_answer_ids, text_response, number_response, json_response, is_correct, points_earned, answered_at, time_spent_seconds, bet_amount
FROM quiz_responses
WHERE submission_id = @submissionid::char(28);

-- name: GetQuizResponsesBySubmissionIDs :many
SELECT
    r.id, r.submission_id, r.question_id, r.selected_answer_ids,
    r.text_response, r.number_response, r.json_response,
    r.is_correct, r.points_earned, r.answered_at, r.time_spent_seconds,
    r.bet_amount, r.score_journal_id, q.question_type, q.betting_enabled
FROM quiz_responses r
JOIN quiz_questions q ON r.question_id = q.id
WHERE r.submission_id = ANY(@submission_ids::char(28)[])
ORDER BY r.submission_id;

-- name: GetQuizResponseBySubmissionAndQuestion :one
SELECT id, submission_id, question_id, selected_answer_ids, text_response, number_response, json_response, is_correct, points_earned, answered_at, time_spent_seconds, bet_amount
FROM quiz_responses
WHERE submission_id = @submissionid::char(28)
    AND question_id = @questionid::char(28);

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
    points_earned,
    time_spent_seconds,
    bet_amount
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
    sqlc.narg('pointsearned')::int,
    sqlc.narg('timespentseconds')::int,
    sqlc.narg('betamount')::int
)
RETURNING id, submission_id, question_id, selected_answer_ids, text_response, number_response, json_response, is_correct, points_earned, answered_at, time_spent_seconds, bet_amount;

-- name: UpdateQuizResponse :one
UPDATE quiz_responses
SET
    selected_answer_ids = COALESCE(sqlc.narg('selectedanswerids')::jsonb, selected_answer_ids),
    text_response = COALESCE(sqlc.narg('textresponse')::text, text_response),
    number_response = COALESCE(sqlc.narg('numberresponse')::decimal, number_response),
    json_response = COALESCE(sqlc.narg('jsonresponse')::jsonb, json_response),
    is_correct = COALESCE(sqlc.narg('iscorrect')::bool, is_correct),
    points_earned = COALESCE(sqlc.narg('pointsearned')::int, points_earned),
    time_spent_seconds = COALESCE(sqlc.narg('timespentseconds')::int, time_spent_seconds),
    bet_amount = COALESCE(sqlc.narg('betamount')::int, bet_amount),
    answered_at = now()
WHERE id = @id::char(28)
RETURNING id, submission_id, question_id, selected_answer_ids, text_response, number_response, json_response, is_correct, points_earned, answered_at, time_spent_seconds, bet_amount;

-- name: CalculateSubmissionScore :one
SELECT
    COALESCE(SUM(points_earned), 0)::int AS score
FROM quiz_responses
WHERE submission_id = @submissionid::char(28);

-- name: CalculateSubmissionPointsFromResponses :one
SELECT COALESCE(SUM(points_earned), 0)::int AS total_points
FROM quiz_responses
WHERE submission_id = @submissionid::char(28);

-- name: UpdateBetResult :one
UPDATE quiz_responses
SET points_earned = @pointsearned::int
WHERE id = @id::char(28)
RETURNING id, submission_id, question_id, selected_answer_ids, text_response, number_response, json_response, is_correct, points_earned, answered_at, time_spent_seconds, bet_amount;

-- name: UpdateBetResults :many
UPDATE quiz_responses
SET points_earned = data.points_earned
FROM (
    SELECT unnest(@ids::char(28)[]) AS id, unnest(@pointsearned::int[]) AS points_earned
) AS data
WHERE quiz_responses.id = data.id
RETURNING quiz_responses.id, quiz_responses.submission_id, quiz_responses.question_id, quiz_responses.selected_answer_ids, quiz_responses.text_response, quiz_responses.number_response, quiz_responses.json_response, quiz_responses.is_correct, quiz_responses.points_earned, quiz_responses.answered_at, quiz_responses.time_spent_seconds, quiz_responses.bet_amount;

-- name: GetQuizResponseWithContext :one
SELECT
    r.id, r.submission_id, r.question_id, r.points_earned, r.bet_amount, r.score_journal_id,
    s.user_id, s.quiz_id,
    q.project_id, q.name as quiz_name, q.challenge_id,
    c.event_id
FROM quiz_responses r
JOIN quiz_submissions s ON r.submission_id = s.id
JOIN quizzes q ON s.quiz_id = q.id
LEFT JOIN challenges c ON q.challenge_id = c.id
WHERE r.id = @id::char(28);

-- name: GetQuizResponsesWithContext :many
SELECT
    r.id, r.submission_id, r.question_id, r.points_earned, r.bet_amount, r.score_journal_id,
    s.user_id, s.quiz_id,
    q.project_id, q.name as quiz_name, q.challenge_id,
    c.event_id
FROM quiz_responses r
JOIN quiz_submissions s ON r.submission_id = s.id
JOIN quizzes q ON s.quiz_id = q.id
LEFT JOIN challenges c ON q.challenge_id = c.id
WHERE r.id = ANY(@ids::char(28)[]);

-- name: UpdateBetResultWithJournal :one
UPDATE quiz_responses
SET points_earned = @pointsearned::int,
    score_journal_id = @scorejournalid::char(28)
WHERE id = @id::char(28)
RETURNING id, submission_id, question_id, selected_answer_ids, text_response,
          number_response, json_response, is_correct, points_earned, answered_at,
          time_spent_seconds, bet_amount, score_journal_id;

-- name: UpdateBetResultsWithJournal :many
UPDATE quiz_responses
SET points_earned = data.points_earned,
    score_journal_id = data.score_journal_id
FROM (
    SELECT
        unnest(@ids::char(28)[]) AS id,
        unnest(@pointsearned::int[]) AS points_earned,
        unnest(@scorejournalids::char(28)[]) AS score_journal_id
) AS data
WHERE quiz_responses.id = data.id
RETURNING quiz_responses.id, quiz_responses.submission_id, quiz_responses.question_id, quiz_responses.selected_answer_ids, quiz_responses.text_response, quiz_responses.number_response, quiz_responses.json_response, quiz_responses.is_correct, quiz_responses.points_earned, quiz_responses.answered_at, quiz_responses.time_spent_seconds, quiz_responses.bet_amount, quiz_responses.score_journal_id;

-- name: GetQuizResponseScoreJournalID :one
SELECT score_journal_id FROM quiz_responses WHERE id = @id::char(28);
