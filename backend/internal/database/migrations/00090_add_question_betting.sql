-- +goose Up
-- +goose StatementBegin

-- Add betting configuration columns to quiz_questions
ALTER TABLE quiz_questions
ADD COLUMN betting_enabled BOOLEAN DEFAULT false NOT NULL,
ADD COLUMN betting_min_percentage DECIMAL,
ADD COLUMN betting_max_percentage DECIMAL,
ADD COLUMN betting_min_absolute INT,
ADD COLUMN betting_max_absolute INT;

-- Add CHECK constraints for percentage range (0-100)
ALTER TABLE quiz_questions
ADD CONSTRAINT quiz_questions_betting_min_percentage_range
CHECK (betting_min_percentage IS NULL OR (betting_min_percentage >= 0 AND betting_min_percentage <= 100));

ALTER TABLE quiz_questions
ADD CONSTRAINT quiz_questions_betting_max_percentage_range
CHECK (betting_max_percentage IS NULL OR (betting_max_percentage >= 0 AND betting_max_percentage <= 100));

-- Add CHECK constraint for min <= max percentage
ALTER TABLE quiz_questions
ADD CONSTRAINT quiz_questions_betting_percentage_min_max
CHECK (betting_min_percentage IS NULL OR betting_max_percentage IS NULL OR betting_min_percentage <= betting_max_percentage);

-- Add CHECK constraints for absolute values >= 0
ALTER TABLE quiz_questions
ADD CONSTRAINT quiz_questions_betting_min_absolute_positive
CHECK (betting_min_absolute IS NULL OR betting_min_absolute >= 0);

ALTER TABLE quiz_questions
ADD CONSTRAINT quiz_questions_betting_max_absolute_positive
CHECK (betting_max_absolute IS NULL OR betting_max_absolute >= 0);

-- Add CHECK constraint for min <= max absolute
ALTER TABLE quiz_questions
ADD CONSTRAINT quiz_questions_betting_absolute_min_max
CHECK (betting_min_absolute IS NULL OR betting_max_absolute IS NULL OR betting_min_absolute <= betting_max_absolute);

-- Add bet amount column to quiz_responses
ALTER TABLE quiz_responses
ADD COLUMN bet_amount INT;

-- Add CHECK constraint for bet_amount >= 0
ALTER TABLE quiz_responses
ADD CONSTRAINT quiz_responses_bet_amount_positive
CHECK (bet_amount IS NULL OR bet_amount >= 0);

-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin

-- Remove bet_amount from quiz_responses
ALTER TABLE quiz_responses
DROP CONSTRAINT IF EXISTS quiz_responses_bet_amount_positive;

ALTER TABLE quiz_responses
DROP COLUMN IF EXISTS bet_amount;

-- Remove betting columns from quiz_questions
ALTER TABLE quiz_questions
DROP CONSTRAINT IF EXISTS quiz_questions_betting_absolute_min_max;

ALTER TABLE quiz_questions
DROP CONSTRAINT IF EXISTS quiz_questions_betting_max_absolute_positive;

ALTER TABLE quiz_questions
DROP CONSTRAINT IF EXISTS quiz_questions_betting_min_absolute_positive;

ALTER TABLE quiz_questions
DROP CONSTRAINT IF EXISTS quiz_questions_betting_percentage_min_max;

ALTER TABLE quiz_questions
DROP CONSTRAINT IF EXISTS quiz_questions_betting_max_percentage_range;

ALTER TABLE quiz_questions
DROP CONSTRAINT IF EXISTS quiz_questions_betting_min_percentage_range;

ALTER TABLE quiz_questions
DROP COLUMN IF EXISTS betting_max_absolute;

ALTER TABLE quiz_questions
DROP COLUMN IF EXISTS betting_min_absolute;

ALTER TABLE quiz_questions
DROP COLUMN IF EXISTS betting_max_percentage;

ALTER TABLE quiz_questions
DROP COLUMN IF EXISTS betting_min_percentage;

ALTER TABLE quiz_questions
DROP COLUMN IF EXISTS betting_enabled;

-- +goose StatementEnd
