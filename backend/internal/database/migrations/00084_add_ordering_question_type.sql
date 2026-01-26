-- +goose Up
-- +goose StatementBegin

-- Add ORDERING to the quiz_questions question_type check constraint
ALTER TABLE quiz_questions
DROP CONSTRAINT quiz_questions_question_type_check;

ALTER TABLE quiz_questions
ADD CONSTRAINT quiz_questions_question_type_check
CHECK (question_type IN ('PREDEFINED', 'FREE_TEXT', 'NUMBER', 'JSON', 'ORDERING'));

-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin

-- Remove ORDERING from the quiz_questions question_type check constraint
ALTER TABLE quiz_questions
DROP CONSTRAINT quiz_questions_question_type_check;

ALTER TABLE quiz_questions
ADD CONSTRAINT quiz_questions_question_type_check
CHECK (question_type IN ('PREDEFINED', 'FREE_TEXT', 'NUMBER', 'JSON'));

-- +goose StatementEnd
