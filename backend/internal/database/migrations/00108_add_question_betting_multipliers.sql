-- +goose Up
-- +goose StatementBegin

-- Payout multipliers for bets on a question.
-- winnings = floor(bet * multiplier); net points = winnings - bet.
-- NULL means the application default (correct: 2.0, wrong: 0).
-- At most 2 decimals and 0..100, matching the fixed-point payout in services/betting.
ALTER TABLE quiz_questions
ADD COLUMN betting_multiplier_correct NUMERIC(5, 2),
ADD COLUMN betting_multiplier_wrong NUMERIC(5, 2);

ALTER TABLE quiz_questions
ADD CONSTRAINT quiz_questions_betting_multiplier_correct_range
CHECK (betting_multiplier_correct IS NULL OR (betting_multiplier_correct >= 0 AND betting_multiplier_correct <= 100));

ALTER TABLE quiz_questions
ADD CONSTRAINT quiz_questions_betting_multiplier_wrong_range
CHECK (betting_multiplier_wrong IS NULL OR (betting_multiplier_wrong >= 0 AND betting_multiplier_wrong <= 100));

ALTER TABLE quiz_questions
ADD CONSTRAINT quiz_questions_betting_multiplier_wrong_le_correct
CHECK (betting_multiplier_correct IS NULL OR betting_multiplier_wrong IS NULL OR betting_multiplier_wrong <= betting_multiplier_correct);

-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin

ALTER TABLE quiz_questions
DROP CONSTRAINT IF EXISTS quiz_questions_betting_multiplier_wrong_le_correct;

ALTER TABLE quiz_questions
DROP CONSTRAINT IF EXISTS quiz_questions_betting_multiplier_wrong_range;

ALTER TABLE quiz_questions
DROP CONSTRAINT IF EXISTS quiz_questions_betting_multiplier_correct_range;

ALTER TABLE quiz_questions
DROP COLUMN IF EXISTS betting_multiplier_wrong;

ALTER TABLE quiz_questions
DROP COLUMN IF EXISTS betting_multiplier_correct;

-- +goose StatementEnd
