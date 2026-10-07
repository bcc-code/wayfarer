-- +goose Up
-- +goose StatementBegin

-- Payout multipliers for bets on a question.
-- winnings = bet * multiplier; net points = winnings - bet.
-- NULL means the application default (correct: 2.0, wrong: 0).
ALTER TABLE quiz_questions
ADD COLUMN betting_multiplier_correct DECIMAL,
ADD COLUMN betting_multiplier_wrong DECIMAL;

ALTER TABLE quiz_questions
ADD CONSTRAINT quiz_questions_betting_multiplier_correct_positive
CHECK (betting_multiplier_correct IS NULL OR betting_multiplier_correct >= 0);

ALTER TABLE quiz_questions
ADD CONSTRAINT quiz_questions_betting_multiplier_wrong_positive
CHECK (betting_multiplier_wrong IS NULL OR betting_multiplier_wrong >= 0);

ALTER TABLE quiz_questions
ADD CONSTRAINT quiz_questions_betting_multiplier_wrong_le_correct
CHECK (betting_multiplier_correct IS NULL OR betting_multiplier_wrong IS NULL OR betting_multiplier_wrong <= betting_multiplier_correct);

-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin

ALTER TABLE quiz_questions
DROP CONSTRAINT IF EXISTS quiz_questions_betting_multiplier_wrong_le_correct;

ALTER TABLE quiz_questions
DROP CONSTRAINT IF EXISTS quiz_questions_betting_multiplier_wrong_positive;

ALTER TABLE quiz_questions
DROP CONSTRAINT IF EXISTS quiz_questions_betting_multiplier_correct_positive;

ALTER TABLE quiz_questions
DROP COLUMN IF EXISTS betting_multiplier_wrong;

ALTER TABLE quiz_questions
DROP COLUMN IF EXISTS betting_multiplier_correct;

-- +goose StatementEnd
