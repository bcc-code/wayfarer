-- +goose Up
-- How the alternatives of a predefined question are labelled in the UI.
-- The label is derived from answer_order at render time, never stored per
-- answer: a stored one can fall out of step with the order the answers are
-- shown in, and would be translated along with the answer text.
ALTER TABLE quiz_questions
ADD COLUMN IF NOT EXISTS answer_label_style VARCHAR(16) NOT NULL DEFAULT 'NONE';

ALTER TABLE quiz_questions
DROP CONSTRAINT IF EXISTS quiz_questions_answer_label_style_check;

ALTER TABLE quiz_questions
ADD CONSTRAINT quiz_questions_answer_label_style_check
CHECK (answer_label_style IN ('NONE', 'NUMERIC', 'ALPHA'));

-- +goose Down
ALTER TABLE quiz_questions
DROP CONSTRAINT IF EXISTS quiz_questions_answer_label_style_check;

ALTER TABLE quiz_questions DROP COLUMN IF EXISTS answer_label_style;
