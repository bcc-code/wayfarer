-- +goose Up
-- +goose StatementBegin

-- Add timeout_seconds to quiz_questions
ALTER TABLE quiz_questions ADD COLUMN timeout_seconds INT;

-- Copy existing quiz-level question timeout to all questions (if set)
UPDATE quiz_questions qq
SET timeout_seconds = q.question_timeout_seconds
FROM quizzes q
WHERE qq.quiz_id = q.id AND q.question_timeout_seconds IS NOT NULL;

-- Remove the CHECK constraint first (it references the column)
ALTER TABLE quizzes DROP CONSTRAINT quizzes_check;

-- Remove question_timeout_seconds from quizzes table
ALTER TABLE quizzes DROP COLUMN question_timeout_seconds;

-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin

ALTER TABLE quizzes ADD COLUMN question_timeout_seconds INT;
ALTER TABLE quizzes ADD CONSTRAINT quizzes_check CHECK (timeout_seconds IS NULL OR question_timeout_seconds IS NULL);
ALTER TABLE quiz_questions DROP COLUMN timeout_seconds;

-- +goose StatementEnd
