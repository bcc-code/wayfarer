-- Aggregated quiz results for the admin panel. Pools all completed submissions
-- for a quiz across every session; nothing here runs per question.
--
-- Numbers, free text and ordering come back as raw rows on purpose: bucketing,
-- grouping and position-matching are shaped in Go, where they are testable.

-- name: GetQuizResultsSummary :one
-- AVG over an empty or all-NULL set is NULL, which a plain ::float8 cast
-- mis-types as non-nullable. COALESCE keeps the scan safe; scored_count is what
-- separates "no scores yet" from "the average is zero".
SELECT
    COUNT(*)::int AS submission_count,
    COUNT(DISTINCT s.user_id)::int AS participant_count,
    COUNT(DISTINCT s.session_id)::int AS session_count,
    COUNT(s.score)::int AS scored_count,
    COALESCE(AVG(s.score), 0)::float8 AS average_score,
    COALESCE(AVG(s.max_score), 0)::float8 AS average_max_score
FROM quiz_submissions s
WHERE s.quiz_id = @quizid::char(28)
    AND s.completed_at IS NOT NULL;

-- name: GetQuizResponseCountsByQuestion :many
SELECT
    r.question_id,
    COUNT(*)::int AS response_count,
    COUNT(*) FILTER (WHERE r.is_correct)::int AS correct_count
FROM quiz_responses r
JOIN quiz_submissions s ON s.id = r.submission_id
WHERE s.quiz_id = @quizid::char(28)
    AND s.completed_at IS NOT NULL
GROUP BY r.question_id;

-- name: GetQuizPredefinedAnswerCounts :many
-- Only answers somebody selected; the resolver fills unpicked ones in as zero.
SELECT
    r.question_id,
    sel.answer_id::char(28) AS answer_id,
    COUNT(*)::int AS count
FROM quiz_responses r
JOIN quiz_submissions s ON s.id = r.submission_id
CROSS JOIN LATERAL jsonb_array_elements_text(r.selected_answer_ids) AS sel(answer_id)
WHERE s.quiz_id = @quizid::char(28)
    AND s.completed_at IS NOT NULL
    AND r.selected_answer_ids IS NOT NULL
GROUP BY r.question_id, sel.answer_id;

-- name: GetQuizNumberResponses :many
SELECT
    r.question_id,
    r.number_response::float8 AS value
FROM quiz_responses r
JOIN quiz_submissions s ON s.id = r.submission_id
WHERE s.quiz_id = @quizid::char(28)
    AND s.completed_at IS NOT NULL
    AND r.number_response IS NOT NULL
ORDER BY r.question_id, r.number_response;

-- name: GetQuizFreeTextResponses :many
SELECT
    r.question_id,
    r.text_response::text AS text_response
FROM quiz_responses r
JOIN quiz_submissions s ON s.id = r.submission_id
WHERE s.quiz_id = @quizid::char(28)
    AND s.completed_at IS NOT NULL
    AND r.text_response IS NOT NULL
    AND btrim(r.text_response) <> ''
ORDER BY r.question_id, r.answered_at;

-- name: GetQuizOrderingResponses :many
-- ORDERING stores the submitted sequence of item IDs in json_response.
SELECT
    r.question_id,
    r.json_response AS submitted_order
FROM quiz_responses r
JOIN quiz_submissions s ON s.id = r.submission_id
JOIN quiz_questions q ON q.id = r.question_id
WHERE s.quiz_id = @quizid::char(28)
    AND s.completed_at IS NOT NULL
    AND q.question_type = 'ORDERING'
    AND r.json_response IS NOT NULL
ORDER BY r.question_id;
